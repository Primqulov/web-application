package admin

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ishchibormi/backend/config"
	"github.com/ishchibormi/backend/internal/models"
	"github.com/ishchibormi/backend/internal/notification"
	"github.com/ishchibormi/backend/internal/upload"
	"github.com/ishchibormi/backend/pkg/httpx"
	"github.com/ishchibormi/backend/pkg/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	Cfg      config.Config
	Admins   *mongo.Collection
	Users    *mongo.Collection
	Elons    *mongo.Collection
	Cats     *mongo.Collection
	AuditCol *mongo.Collection
	Notify   *notification.Service
	Apps     *mongo.Collection
	Storage  *storage.Service
}

func NewHandler(cfg config.Config, db *mongo.Database, n *notification.Service, s *storage.Service) *Handler {
	return &Handler{
		Cfg:      cfg,
		Admins:   db.Collection("admins"),
		Users:    db.Collection("users"),
		Elons:    db.Collection("elons"),
		Cats:     db.Collection("categories"),
		AuditCol: db.Collection("admin_audit"),
		Notify:   n,
		Apps:     db.Collection("applications"),
		Storage:  s,
	}
}

type loginReq struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Err(w, err)
		return
	}
	var a models.Admin
	if err := h.Admins.FindOne(r.Context(), bson.M{"username": req.Username, "isActive": true}).Decode(&a); err != nil {
		httpx.Err(w, httpx.NewError(401, "bad_credentials", "invalid credentials"))
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(req.Password)); err != nil {
		httpx.Err(w, httpx.NewError(401, "bad_credentials", "invalid credentials"))
		return
	}
	tok, err := httpx.IssueAdminToken(h.Cfg.JWTAccessSecret, a.ID.Hex(), a.Role, h.Cfg.JWTAccessTTL)
	if err != nil {
		httpx.Err(w, err)
		return
	}
	httpx.JSON(w, 200, map[string]any{"accessToken": tok, "admin": a})
}

func (h *Handler) audit(r *http.Request, action, target, detail string) {
	aid, _ := primitive.ObjectIDFromHex(httpx.AdminID(r))
	_, _ = h.AuditCol.InsertOne(r.Context(), models.AdminAudit{
		AdminID: aid, Action: action, Target: target, Detail: detail, CreatedAt: time.Now(),
	})
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usersCount, _ := h.Users.CountDocuments(ctx, bson.M{"isDeleted": bson.M{"$ne": true}})
	elonsCount, _ := h.Elons.CountDocuments(ctx, bson.M{"isDeleted": bson.M{"$ne": true}})
	completed, _ := h.Apps.CountDocuments(ctx, bson.M{"status": "completed"})
	httpx.JSON(w, 200, map[string]any{
		"users": usersCount, "elons": elonsCount,
		"completed": completed,
	})
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	cur, err := h.Users.Find(r.Context(), bson.M{}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(500))
	if err != nil {
		httpx.Err(w, err)
		return
	}
	defer cur.Close(r.Context())
	out := []models.User{}
	for cur.Next(r.Context()) {
		var u models.User
		if err := cur.Decode(&u); err == nil {
			out = append(out, u)
		}
	}
	httpx.JSON(w, 200, out)
}

type setBlockReq struct {
	IsBlocked bool `json:"isBlocked"`
}

func (h *Handler) BlockUser(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Err(w, httpx.NewError(400, "bad_id", "bad id"))
		return
	}
	var req setBlockReq
	_ = httpx.Decode(r, &req)
	_, err = h.Users.UpdateOne(r.Context(), bson.M{"_id": id}, bson.M{"$set": bson.M{"isBlocked": req.IsBlocked}})
	if err != nil {
		httpx.Err(w, err)
		return
	}
	h.audit(r, "user_block", id.Hex(), "isBlocked=set")
	httpx.JSON(w, 200, map[string]bool{"ok": true})
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Err(w, httpx.NewError(400, "bad_id", "bad id"))
		return
	}
	// Best-effort: remove the user's avatar from S3, plus images of all their elons.
	var u models.User
	if err := h.Users.FindOne(r.Context(), bson.M{"_id": id}).Decode(&u); err == nil {
		go upload.DeleteByURL(h.Storage, u.AvatarURL)
	}
	cur, _ := h.Elons.Find(r.Context(), bson.M{"ownerId": id})
	if cur != nil {
		defer cur.Close(r.Context())
		for cur.Next(r.Context()) {
			var e models.Elon
			if err := cur.Decode(&e); err == nil {
				for _, u := range e.Images {
					go upload.DeleteByURL(h.Storage, u)
				}
			}
		}
	}
	_, err = h.Users.UpdateOne(r.Context(), bson.M{"_id": id}, bson.M{"$set": bson.M{"isDeleted": true}})
	if err != nil {
		httpx.Err(w, err)
		return
	}
	h.audit(r, "user_delete", id.Hex(), "soft-delete")
	httpx.JSON(w, 200, map[string]bool{"ok": true})
}

func (h *Handler) ListElons(w http.ResponseWriter, r *http.Request) {
	cur, err := h.Elons.Find(r.Context(), bson.M{}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(500))
	if err != nil {
		httpx.Err(w, err)
		return
	}
	defer cur.Close(r.Context())
	out := []models.Elon{}
	for cur.Next(r.Context()) {
		var e models.Elon
		if err := cur.Decode(&e); err == nil {
			out = append(out, e)
		}
	}
	httpx.JSON(w, 200, out)
}

func (h *Handler) DeleteElon(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Err(w, httpx.NewError(400, "bad_id", "bad id"))
		return
	}
	var prev models.Elon
	_ = h.Elons.FindOne(r.Context(), bson.M{"_id": id}).Decode(&prev)
	_, err = h.Elons.UpdateOne(r.Context(), bson.M{"_id": id}, bson.M{"$set": bson.M{"isDeleted": true, "status": "cancelled"}})
	if err != nil {
		httpx.Err(w, err)
		return
	}
	for _, u := range prev.Images {
		go upload.DeleteByURL(h.Storage, u)
	}
	h.audit(r, "elon_delete", id.Hex(), "force")
	httpx.JSON(w, 200, map[string]bool{"ok": true})
}

func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cur, err := h.Cats.Find(r.Context(), bson.M{}, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		httpx.Err(w, err)
		return
	}
	defer cur.Close(r.Context())
	out := []models.Category{}
	for cur.Next(r.Context()) {
		var c models.Category
		if err := cur.Decode(&c); err == nil {
			out = append(out, c)
		}
	}
	httpx.JSON(w, 200, out)
}

type setActiveReq struct {
	IsActive bool `json:"isActive"`
}

func (h *Handler) SetCategoryActive(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Err(w, httpx.NewError(400, "bad_id", "bad id"))
		return
	}
	var req setActiveReq
	_ = httpx.Decode(r, &req)
	_, err = h.Cats.UpdateOne(r.Context(), bson.M{"_id": id}, bson.M{"$set": bson.M{"isActive": req.IsActive}})
	if err != nil {
		httpx.Err(w, err)
		return
	}
	h.audit(r, "category_active", id.Hex(), "")
	httpx.JSON(w, 200, map[string]bool{"ok": true})
}

type broadcastReq struct {
	Title string `json:"title" validate:"required"`
	Body  string `json:"body"`
}

func (h *Handler) Broadcast(w http.ResponseWriter, r *http.Request) {
	var req broadcastReq
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Err(w, err)
		return
	}
	cur, err := h.Users.Find(r.Context(), bson.M{"isDeleted": bson.M{"$ne": true}, "isBlocked": bson.M{"$ne": true}})
	if err != nil {
		httpx.Err(w, err)
		return
	}
	defer cur.Close(r.Context())
	count := 0
	for cur.Next(r.Context()) {
		var u models.User
		if err := cur.Decode(&u); err == nil {
			h.Notify.Push(r.Context(), u.ID, "system", req.Title, req.Body, nil)
			count++
		}
	}
	h.audit(r, "broadcast", req.Title, req.Body)
	httpx.JSON(w, 200, map[string]int{"sent": count})
}

func (h *Handler) Audit(w http.ResponseWriter, r *http.Request) {
	cur, err := h.AuditCol.Find(r.Context(), bson.M{}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(500))
	if err != nil {
		httpx.Err(w, err)
		return
	}
	defer cur.Close(r.Context())
	out := []models.AdminAudit{}
	for cur.Next(r.Context()) {
		var a models.AdminAudit
		if err := cur.Decode(&a); err == nil {
			out = append(out, a)
		}
	}
	httpx.JSON(w, 200, out)
}
