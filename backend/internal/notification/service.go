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

type readReq struct {
	RelatedIDs  []string `json:"relatedIds"`
	RelatedType string   `json:"relatedType"`
}

// Read marks a targeted subset of the user's unread notifications read — either
// those tied to specific related entities (relatedIds) or a whole related type
// (relatedType, e.g. "application"). Used to clear the "red dot" indicators as
// the user views the corresponding items.
func (s *Service) Read(w http.ResponseWriter, r *http.Request) {
	uid, _ := primitive.ObjectIDFromHex(httpx.UserID(r))
	var req readReq
	_ = httpx.Decode(r, &req)
	filter := bson.M{"userId": uid, "isRead": false}
	switch {
	case len(req.RelatedIDs) > 0:
		oids := make([]primitive.ObjectID, 0, len(req.RelatedIDs))
		for _, h := range req.RelatedIDs {
			if oid, err := primitive.ObjectIDFromHex(h); err == nil {
				oids = append(oids, oid)
			}
		}
		if len(oids) == 0 {
			httpx.JSON(w, 200, map[string]bool{"ok": true})
			return
		}
		filter["relatedEntity.id"] = bson.M{"$in": oids}
	case req.RelatedType != "":
		filter["relatedEntity.type"] = req.RelatedType
	default:
		httpx.Err(w, httpx.NewError(400, "bad_request", "relatedIds yoki relatedType kerak"))
		return
	}
	_, _ = s.Col.UpdateMany(r.Context(), filter, bson.M{"$set": bson.M{"isRead": true}})
	httpx.JSON(w, 200, map[string]bool{"ok": true})
}
