package server

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocketHub struct {
	mu      sync.RWMutex
	clients map[int64]map[*websocket.Conn]*webSocketClient
}

type WebSocketMessage struct {
	UserID int64  `json:"-"`
	Type   string `json:"type"`
	Data   any    `json:"data"`
}

type webSocketClient struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{clients: make(map[int64]map[*websocket.Conn]*webSocketClient)}
}

func (h *WebSocketHub) Register(userID int64, conn *websocket.Conn) *webSocketClient {
	if h == nil || userID <= 0 || conn == nil {
		return nil
	}
	client := &webSocketClient{conn: conn}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[*websocket.Conn]*webSocketClient)
	}
	h.clients[userID][conn] = client
	return client
}

func (h *WebSocketHub) Unregister(userID int64, conn *websocket.Conn) {
	if h == nil || userID <= 0 || conn == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	connections, ok := h.clients[userID]
	if !ok {
		return
	}
	delete(connections, conn)
	if len(connections) == 0 {
		delete(h.clients, userID)
	}
}

func (h *WebSocketHub) SendToUser(userID int64, msgType string, data any) error {
	if h == nil || userID <= 0 {
		return nil
	}
	payload, err := json.Marshal(WebSocketMessage{Type: msgType, Data: data})
	if err != nil {
		return err
	}
	h.mu.RLock()
	conns := make([]*webSocketClient, 0, len(h.clients[userID]))
	for _, client := range h.clients[userID] {
		conns = append(conns, client)
	}
	h.mu.RUnlock()
	for _, client := range conns {
		if err := client.WriteMessage(websocket.TextMessage, payload); err != nil {
			h.Unregister(userID, client.conn)
			_ = client.conn.Close()
		}
	}
	return nil
}

func (h *WebSocketHub) ConnectionCount(userID int64) int {
	if h == nil || userID <= 0 {
		return 0
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[userID])
}

func (h *WebSocketHub) TotalConnections() int {
	if h == nil {
		return 0
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	total := 0
	for _, conns := range h.clients {
		total += len(conns)
	}
	return total
}

func (c *webSocketClient) WriteJSON(payload any) error {
	if c == nil || c.conn == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteJSON(payload)
}

func (c *webSocketClient) WriteMessage(messageType int, payload []byte) error {
	if c == nil || c.conn == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteMessage(messageType, payload)
}
