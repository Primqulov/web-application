package finance

import (
	"net/http"

	"github.com/ishchibormi/backend/internal/models"
	"github.com/ishchibormi/backend/pkg/httpx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Handler struct {
	Col *mongo.Collection
}

func NewHandler(db *mongo.Database) *Handler {
	return &Handler{Col: db.Collection("finance_entries")}
}

type summary struct {
	EarnedSum         int64                 `json:"earnedSum"`
	SpentSum          int64                 `json:"spentSum"`
	CompletedCount    int                   `json:"completedCount"`
	NegotiableCount   int                   `json:"negotiableCount"`
	CancelledCount    int                   `json:"cancelledCount"`
	Entries           []models.FinanceEntry `json:"entries"`
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	uid, _ := primitive.ObjectIDFromHex(httpx.UserID(r))
	cur, err := h.Col.Find(r.Context(), bson.M{"userId": uid}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		httpx.Err(w, err)
		return
	}
	defer cur.Close(r.Context())
	out := summary{Entries: []models.FinanceEntry{}}
	for cur.Next(r.Context()) {
		var e models.FinanceEntry
		if err := cur.Decode(&e); err != nil {
			continue
		}
		out.Entries = append(out.Entries, e)
		if e.Status == "completed" {
			out.CompletedCount++
			if e.IsNegotiable {
				out.NegotiableCount++
			} else {
				if e.Type == "earned" {
					out.EarnedSum += e.Amount
				} else {
					out.SpentSum += e.Amount
				}
			}
		} else if e.Status == "cancelled" {
			out.CancelledCount++
		}
	}
	httpx.JSON(w, 200, out)
}
