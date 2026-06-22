package notification

import (
	"context"
	"net/http"
	"time"

	"github.com/ishchibormi/backend/internal/models"
	"github.com/ishchibormi/backend/pkg/httpx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Pusher pushes a serialized notification to any online WS clients for a user.
type Pusher interface {
	PushUser(userID primitive.ObjectID, kind string, payload any)
}

type Service struct {
	Col    *mongo.Collection
	Pusher Pusher
}

func New(db *mongo.Database) *Service {
	return &Service{Col: db.Collection("notifications")}
}

func (s *Service) AttachPusher(p Pusher) { s.Pusher = p }

func (s *Service) Push(ctx context.Context, userID primitive.ObjectID, typ, title, body string, rel *models.RelatedEntity) {
	n := models.Notification{
		UserID: userID, Type: typ, Title: title, Body: body,
		RelatedEntity: rel, IsRead: false, CreatedAt: time.Now(),
	}
	res, err := s.Col.InsertOne(ctx, n)
	if err == nil {
		if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
			n.ID = oid
		}
	}
	if s.Pusher != nil {
		s.Pusher.PushUser(userID, "notification", n)
	}
}

func (s *Service) List(w http.ResponseWriter, r *http.Request) {
	uid, _ := primitive.ObjectIDFromHex(httpx.UserID(r))
	cur, err := s.Col.Find(r.Context(),
		bson.M{"userId": uid},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(200))
	if err != nil {
		httpx.Err(w, err)
		return
	}
	defer cur.Close(r.Context())
	out := []models.Notification{}
	for cur.Next(r.Context()) {
		var n models.Notification
		if err := cur.Decode(&n); err == nil {
			out = append(out, n)
		}
	}
	httpx.JSON(w, 200, out)
}

func (s *Service) ReadAll(w http.ResponseWriter, r *http.Request) {
	uid, _ := primitive.ObjectIDFromHex(httpx.UserID(r))
	_, _ = s.Col.UpdateMany(r.Context(), bson.M{"userId": uid, "isRead": false}, bson.M{"$set": bson.M{"isRead": true}})
	httpx.JSON(w, 200, map[string]bool{"ok": true})
}
