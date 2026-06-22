package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/ishchibormi/backend/config"
	"github.com/ishchibormi/backend/internal/models"
	"github.com/ishchibormi/backend/pkg/httpx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Handler struct {
	cfg   config.Config
	otp   *OTPRepo
	users *mongo.Collection
}

func NewHandler(cfg config.Config, db *mongo.Database) *Handler {
	return &Handler{
		cfg:   cfg,
		otp:   NewOTPRepo(db, cfg.OTPTTL, cfg.OTPLength),
		users: db.Collection("users"),
	}
}

type requestOTPResp struct {
	TGToken  string `json:"tgToken"`
	BotURL   string `json:"botUrl"`
	DevCode  string `json:"devCode,omitempty"`
	DevPhone string `json:"devPhone,omitempty"`
}

func (h *Handler) RequestOTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tok, err := h.otp.RequestToken(ctx)
	if err != nil {
		httpx.Err(w, err)
		return
	}
	resp := requestOTPResp{
		TGToken: tok,
		BotURL:  "https://t.me/" + h.cfg.TelegramBotUsername + "?start=" + tok,
	}
	httpx.JSON(w, 200, resp)
}

type verifyReq struct {
	Token string `json:"token"`
	Phone string `json:"phone"`
	Code  string `json:"code" validate:"required"`
}

type verifyResp struct {
	AccessToken  string             `json:"accessToken"`
	RefreshToken string             `json:"refreshToken"`
	User         models.User        `json:"user"`
}

func (h *Handler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req verifyReq
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Err(w, err)
		return
	}
	if req.Code == "" {
		httpx.Err(w, httpx.NewError(400, "bad_request", "code required"))
		return
	}
	ctx := r.Context()
	var (
		phone string
		tgID  int64
		err   error
	)
	if req.Token != "" {
		phone, tgID, err = h.otp.VerifyByToken(ctx, req.Token, req.Code)
		// Fallback: if the bot didn't see the deep-link token (user opened the bot
		// directly), it wrote a phone+code record with no tgToken. Match by code.
		if err != nil {
			phone, tgID, err = h.otp.VerifyByCode(ctx, req.Code)
		}
	} else if req.Phone != "" {
		phone, tgID, err = h.otp.VerifyByPhone(ctx, req.Phone, req.Code)
	} else {
		// No token, no phone: still allow code-only match (covers all bot quirks).
		phone, tgID, err = h.otp.VerifyByCode(ctx, req.Code)
	}
	if err != nil {
		httpx.Err(w, httpx.NewError(401, "invalid_code", "invalid or expired code"))
		return
	}
	if phone == "" {
		httpx.Err(w, httpx.NewError(401, "no_phone_bound", "bot has not bound phone yet"))
		return
	}
	user, err := h.upsertUser(ctx, phone, tgID)
	if err != nil {
		httpx.Err(w, err)
		return
	}
	access, err := httpx.IssueUserToken(h.cfg.JWTAccessSecret, user.ID.Hex(), h.cfg.JWTAccessTTL)
	if err != nil {
		httpx.Err(w, err)
		return
	}
	refresh, err := httpx.IssueUserToken(h.cfg.JWTRefreshSecret, user.ID.Hex(), h.cfg.JWTRefreshTTL)
	if err != nil {
		httpx.Err(w, err)
		return
	}
	httpx.JSON(w, 200, verifyResp{AccessToken: access, RefreshToken: refresh, User: *user})
}

func (h *Handler) upsertUser(ctx context.Context, phone string, tgID int64) (*models.User, error) {
	now := time.Now()
	filter := bson.M{"phone": phone}
	update := bson.M{
		"$setOnInsert": bson.M{
			"createdAt":           now,
			"firstName":           "",
			"lastName":            "",
			"rating":              0.0,
			"reviewsCount":        0,
			"completedJobsCount":  0,
			"langPref":            "latin",
			"themePref":           "light",
			"isPremium":           false,
			"isBlocked":           false,
			"isDeleted":           false,
			"onboardingCompleted": false,
		},
		"$set": bson.M{
			"phone":           phone,
			"telegramId":      tgID,
			"isPhoneVerified": true,
			"updatedAt":       now,
		},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var u models.User
	err := h.users.FindOneAndUpdate(ctx, filter, update, opts).Decode(&u)
	return &u, err
}

type refreshReq struct {
	RefreshToken string `json:"refreshToken"`
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshReq
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Err(w, err)
		return
	}
	uid, err := httpx.ParseUserToken(h.cfg.JWTRefreshSecret, req.RefreshToken)
	if err != nil {
		httpx.Err(w, httpx.NewError(401, "bad_refresh", "invalid refresh token"))
		return
	}
	access, err := httpx.IssueUserToken(h.cfg.JWTAccessSecret, uid, h.cfg.JWTAccessTTL)
	if err != nil {
		httpx.Err(w, err)
		return
	}
	httpx.JSON(w, 200, map[string]string{"accessToken": access})
}

// DevPeekOTP — dev-only endpoint that returns the most recent OTP for a token.
// Guarded by config.OTPDevReturn.
func (h *Handler) DevPeekOTP(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.OTPDevReturn {
		httpx.Err(w, httpx.NewError(404, "not_found", "dev peek disabled"))
		return
	}
	tok := r.URL.Query().Get("token")
	if tok == "" {
		httpx.Err(w, httpx.NewError(400, "bad_request", "token required"))
		return
	}
	doc, err := h.otp.LatestForToken(r.Context(), tok)
	if err != nil {
		httpx.JSON(w, 200, map[string]any{"code": "", "phone": "", "telegramId": 0})
		return
	}
	httpx.JSON(w, 200, map[string]any{
		"code":       doc.Code,
		"phone":      doc.Phone,
		"telegramId": doc.TelegramID,
	})
}

// Used by /api/me to expand the current user object.
func LoadUser(ctx context.Context, users *mongo.Collection, idHex string) (*models.User, error) {
	oid, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return nil, httpx.NewError(401, "bad_token", "bad user id")
	}
	var u models.User
	if err := users.FindOne(ctx, bson.M{"_id": oid}).Decode(&u); err != nil {
		return nil, httpx.NewError(404, "not_found", "user not found")
	}
	return &u, nil
}
