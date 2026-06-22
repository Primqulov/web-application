package chat

import (
	"encoding/json"
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WSMessage struct {
	Kind    string          `json:"kind"`
	Payload json.RawMessage `json:"payload"`
}

type Client struct {
	UserID primitive.ObjectID
	Send   chan []byte
}

type Hub struct {
	mu      sync.RWMutex
	clients map[primitive.ObjectID]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: map[primitive.ObjectID]map[*Client]struct{}{}}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c.UserID]; !ok {
		h.clients[c.UserID] = map[*Client]struct{}{}
	}
	h.clients[c.UserID][c] = struct{}{}
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.clients[c.UserID]; ok {
		delete(set, c)
		close(c.Send)
		if len(set) == 0 {
			delete(h.clients, c.UserID)
		}
	}
}

// PushUser implements notification.Pusher.
func (h *Hub) PushUser(userID primitive.ObjectID, kind string, payload any) {
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	envelope, _ := json.Marshal(WSMessage{Kind: kind, Payload: b})
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients[userID] {
		select {
		case c.Send <- envelope:
		default:
			// drop if client is slow
		}
	}
}

func (h *Hub) IsOnline(userID primitive.ObjectID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[userID]) > 0
}
