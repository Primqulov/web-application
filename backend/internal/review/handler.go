package review

import (
	"context"
	"math"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ishchibormi/backend/internal/models"
	"github.com/ishchibormi/backend/internal/notification"
	"github.com/ishchibormi/backend/pkg/httpx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	Reviews *mongo.Collection
	Apps    *mongo.Collection
	Users   *mongo.Collection
	Notify  *notification.Service
}

func NewHandler(db *mongo.Database, n *notification.Service) *Handler {
	return &Handler{
		Reviews: db.Collection("reviews"),
		Apps:    db.Collection("applications"),
		Users:   db.Collection("users"),
		Notify:  n,
	}
}

type createReq struct {
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	uid, _ := primitive.ObjectIDFromHex(httpx.UserID(r))
	appID, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Err(w, httpx.NewError(400, "bad_id", "bad id"))
		return
	}
	var req createReq
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Err(w, err)
		return
	}
	if req.Rating < 1 || req.Rating > 5 {
		httpx.Err(w, httpx.NewError(400, "bad_rating", "rating must be 1-5"))
		return
	}
	var app models.Application
	if err := h.Apps.FindOne(r.Context(), bson.M{"_id": appID}).Decode(&app); err != nil {
		httpx.Err(w, httpx.NewError(404, "not_found", "application not found"))
		return
	}
	if app.Status != "completed" {
		httpx.Err(w, httpx.NewError(400, "bad_state", "only completed applications can be reviewed"))
		return
	}
	var dir string
	var toID primitive.ObjectID
	switch uid {
	case app.WorkerID:
		dir = "worker_to_employer"
		toID = app.EmployerID
	case app.EmployerID:
		dir = "employer_to_worker"
		toID = app.WorkerID
	default:
		httpx.Err(w, httpx.NewError(403, "forbidden", "not your application"))
		return
	}
	rev := models.Review{
		ApplicationID: appID, ElonID: app.ElonID, FromUserID: uid, ToUserID: toID,
		Direction: dir, Rating: req.Rating, Comment: req.Comment, CreatedAt: time.Now(),
	}
	res, err := h.Reviews.InsertOne(r.Context(), rev)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			httpx.Err(w, httpx.NewError(409, "duplicate", "review already submitted"))
			return
		}
		httpx.Err(w, err)
		return
	}
	rev.ID = res.InsertedID.(primitive.ObjectID)
	// recompute rating for toID
	if err := Recompute(r.Context(), h.Reviews, h.Users, toID); err != nil {
		httpx.Err(w, err)
		return
	}
	h.Notify.Push(r.Context(), toID, "new_review", "Yangi baho", "Sizga yangi baho qoldirildi", &models.RelatedEntity{Type: "review", ID: rev.ID})
	httpx.JSON(w, 201, rev)
}

// Recompute averages all reviews for toID and writes the overall rating plus
// the two role-specific ratings:
//   - "employer_to_worker" reviews -> the user is rated AS A WORKER (workerRating)
//   - "worker_to_employer" reviews -> the user is rated AS AN EMPLOYER (employerRating)
func Recompute(ctx context.Context, reviews, users *mongo.Collection, toID primitive.ObjectID) error {
	cur, err := reviews.Find(ctx, bson.M{"toUserId": toID})
	if err != nil {
		return err
	}
	defer cur.Close(ctx)
	var sum, n int
	var wSum, wN int
	var eSum, eN int
	for cur.Next(ctx) {
		var r models.Review
		if err := cur.Decode(&r); err == nil {
			sum += r.Rating
			n++
			switch r.Direction {
			case "employer_to_worker":
				wSum += r.Rating
				wN++
			case "worker_to_employer":
				eSum += r.Rating
				eN++
			}
		}
	}
	_, err = users.UpdateOne(ctx, bson.M{"_id": toID}, bson.M{"$set": bson.M{
		"rating":               avgRound(sum, n),
		"reviewsCount":         n,
		"workerRating":         avgRound(wSum, wN),
		"workerReviewsCount":   wN,
		"employerRating":       avgRound(eSum, eN),
		"employerReviewsCount": eN,
	}})
	return err
}

func avgRound(sum, n int) float64 {
	if n == 0 {
		return 0
	}
	return math.Round(float64(sum)/float64(n)*10) / 10
}

func (h *Handler) ListForUser(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Err(w, httpx.NewError(400, "bad_id", "bad id"))
		return
	}
	cur, err := h.Reviews.Find(r.Context(), bson.M{"toUserId": id})
	if err != nil {
		httpx.Err(w, err)
		return
	}
	defer cur.Close(r.Context())
	out := []models.Review{}
	for cur.Next(r.Context()) {
		var rv models.Review
		if err := cur.Decode(&rv); err == nil {
			out = append(out, rv)
		}
	}
	httpx.JSON(w, 200, out)
}
