package chat

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/ishchibormi/backend/config"
	"github.com/ishchibormi/backend/internal/models"
	"github.com/ishchibormi/backend/internal/notification"
	"github.com/ishchibormi/backend/pkg/httpx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Handler struct {
	Conv   *mongo.Collection
	Msg    *mongo.Collection
	Users  *mongo.Collection
	Hub    *Hub
	Notify *notification.Service
	Cfg    config.Config
}

func NewHandler(db *mongo.Database, hub *Hub, n *notification.Service, cfg config.Config) *Handler {
	return &Handler{
		Conv:   db.Collection("conversations"),
		Msg:    db.Collection("messages"),
		Users:  db.Collection("users"),
		Hub:    hub,
		Notify: n,
		Cfg:    cfg,
	}
}

func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request) {
	uid, _ := primitive.ObjectIDFromHex(httpx.UserID(r))
	cur, err := h.Conv.Find(r.Context(),
		bson.M{"participantIds": uid},
		options.Find().SetSort(bson.D{{Key: "lastMessageAt", Value: -1}}).SetLimit(100))
	if err != nil {
		httpx.Err(w, err)
		return
	}
	defer cur.Close(r.Context())
	out := []models.Conversation{}
	for cur.Next(r.Context()) {
		var c models.Conversation
		if err := cur.Decode(&c); err == nil {
			out = append(out, c)
		}
	}
	httpx.JSON(w, 200, out)
}

type startReq struct {
	UserID string `json:"userId"`
}

func (h *Handler) StartConversation(w http.ResponseWriter, r *http.Request) {
	uid, _ := primitive.ObjectIDFromHex(httpx.UserID(r))
	var req startReq
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Err(w, err)
		return
	}
	tid, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		httpx.Err(w, httpx.NewError(400, "bad_id", "bad userId"))
		return
	}
	if uid == tid {
		httpx.Err(w, httpx.NewError(400, "self_chat", "cannot chat with self"))
		return
	}
	parts := []primitive.ObjectID{uid, tid}
	sort.Slice(parts, func(i, j int) bool { return strings.Compare(parts[i].Hex(), parts[j].Hex()) < 0 })

	now := time.Now()
	res := h.Conv.FindOneAndUpdate(r.Context(),
		bson.M{"participantIds": bson.M{"$all": parts, "$size": 2}},
		bson.M{
			"$setOnInsert": bson.M{
				"participantIds":  parts,
				"createdAt":       now,
				"lastMessageAt":   now,
				"lastMessageText": "",
				"unread":          bson.M{},
			},
		},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	var c models.Conversation
	if err := res.Decode(&c); err != nil {
		httpx.Err(w, err)
		return
	}
	httpx.JSON(w, 200, c)
}

func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	uid, _ := primitive.ObjectIDFromHex(httpx.UserID(r))
	cid, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Err(w, httpx.NewError(400, "bad_id", "bad id"))
		return
	}
	// permission check
	var conv models.Conversation
	if err := h.Conv.FindOne(r.Context(), bson.M{"_id": cid, "participantIds": uid}).Decode(&conv); err != nil {
		httpx.Err(w, httpx.NewError(404, "not_found", "conversation not found"))
		return
	}
	cur, err := h.Msg.Find(r.Context(),
		bson.M{"conversationId": cid},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}).SetLimit(500))
	if err != nil {
		httpx.Err(w, err)
		return
	}
	defer cur.Close(r.Context())
	out := []models.Message{}
	for cur.Next(r.Context()) {
		var m models.Message
		if err := cur.Decode(&m); err == nil {
			out = append(out, m)
		}
	}
	// reset unread for this user
	_, _ = h.Conv.UpdateOne(r.Context(), bson.M{"_id": cid}, bson.M{"$set": bson.M{"unread." + uid.Hex(): 0}})
	httpx.JSON(w, 200, out)
}

type sendReq struct {
	Text        string                     `json:"text"`
	Attachments []models.MessageAttachment `json:"attachments"`
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	uid, _ := primitive.ObjectIDFromHex(httpx.UserID(r))
	cid, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Err(w, httpx.NewError(400, "bad_id", "bad id"))
		return
	}
	var req sendReq
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Err(w, err)
		return
	}
	text := strings.TrimSpace(req.Text)
	if text == "" && len(req.Attachments) == 0 {
		httpx.Err(w, httpx.NewError(400, "bad_request", "empty message"))
		return
	}
	var conv models.Conversation
	if err := h.Conv.FindOne(r.Context(), bson.M{"_id": cid, "participantIds": uid}).Decode(&conv); err != nil {
		httpx.Err(w, httpx.NewError(404, "not_found", "conversation not found"))
		return
	}
	// block check: if any participant has blocked the other, deny.
	if h.isBlocked(r, uid, conv.ParticipantIDs) {
		httpx.Err(w, httpx.NewError(403, "blocked", "messaging blocked"))
		return
	}
	now := time.Now()
	preview := text
	if preview == "" && len(req.Attachments) > 0 {
		preview = "📎 " + req.Attachments[0].Name
	}
	m := models.Message{
		ConversationID: cid, SenderID: uid, Text: text,
		Attachments: req.Attachments,
		IsRead: false, CreatedAt: now,
	}
	res, err := h.Msg.InsertOne(r.Context(), m)
	if err != nil {
		httpx.Err(w, err)
		return
	}
	m.ID = res.InsertedID.(primitive.ObjectID)
	// increment unread for everyone except the sender
	update := bson.M{
		"$set": bson.M{
			"lastMessageText": preview,
			"lastMessageAt":   now,
			"lastSenderId":    uid,
		},
	}
	inc := bson.M{}
	for _, p := range conv.ParticipantIDs {
		if p != uid {
			inc["unread."+p.Hex()] = 1
		}
	}
	if len(inc) > 0 {
		update["$inc"] = inc
	}
	_, _ = h.Conv.UpdateOne(r.Context(), bson.M{"_id": cid}, update)
	// push WS event to other participants
	for _, p := range conv.ParticipantIDs {
		if p == uid {
			continue
		}
		h.Hub.PushUser(p, "message", m)
		if !h.Hub.IsOnline(p) {
			h.Notify.Push(r.Context(), p, "new_message", "Yangi xabar", text, &models.RelatedEntity{Type: "conversation", ID: cid})
		}
	}
	// echo to sender too for multi-tab sync
	h.Hub.PushUser(uid, "message", m)
	httpx.JSON(w, 201, m)
}

func (h *Handler) isBlocked(r *http.Request, uid primitive.ObjectID, parts []primitive.ObjectID) bool {
	for _, p := range parts {
		if p == uid {
			continue
		}
		var u models.User
		if err := h.Users.FindOne(r.Context(), bson.M{"_id": p}).Decode(&u); err == nil {
			for _, b := range u.BlockedUserIDs {
				if b == uid {
					return true
				}
			}
		}
		var me models.User
		if err := h.Users.FindOne(r.Context(), bson.M{"_id": uid}).Decode(&me); err == nil {
			for _, b := range me.BlockedUserIDs {
				if b == p {
					return true
				}
			}
		}
	}
	return false
}

// WS endpoint: ws upgrades the connection and registers it to the hub.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *Handler) WS(w http.ResponseWriter, r *http.Request) {
	tok := r.URL.Query().Get("token")
	uidHex, err := httpx.ParseUserToken(h.Cfg.JWTAccessSecret, tok)
	if err != nil {
		httpx.Err(w, httpx.NewError(401, "bad_token", "invalid token"))
		return
	}
	uid, err := primitive.ObjectIDFromHex(uidHex)
	if err != nil {
		httpx.Err(w, httpx.NewError(401, "bad_token", "invalid token"))
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &Client{UserID: uid, Send: make(chan []byte, 32)}
	h.Hub.Register(client)
	// hello
	hello, _ := json.Marshal(WSMessage{Kind: "hello", Payload: json.RawMessage(`{"ok":true}`)})
	client.Send <- hello

	go func() {
		defer conn.Close()
		for b := range client.Send {
			if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
				return
			}
		}
	}()
	// reader loop (just consume / detect close)
	defer h.Hub.Unregister(client)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}
