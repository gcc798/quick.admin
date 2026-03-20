package server

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketHub struct {
	clients    map[int64]map[*webSocketClient]bool
	register   chan *webSocketClient
	unregister chan *webSocketClient
	broadcast  chan *WebSocketMessage
	quit       chan struct{}
	stopOnce   sync.Once
	mu         sync.RWMutex
}

type WebSocketMessage struct {
	UserID  int64  `json:"-"`
	Type    string `json:"type"`
	Data    any    `json:"data"`
	payload []byte `json:"-"`
}

type webSocketClient struct {
	userID int64
	conn   *websocket.Conn
	hub    *WebSocketHub
	mu     sync.Mutex
}

func NewWebSocketHub() *WebSocketHub {
	hub := &WebSocketHub{
		clients:    make(map[int64]map[*webSocketClient]bool),
		register:   make(chan *webSocketClient),
		unregister: make(chan *webSocketClient),
		broadcast:  make(chan *WebSocketMessage, 256),
		quit:       make(chan struct{}),
	}
	go hub.Run()
	return hub
}

func (h *WebSocketHub) Run() {
	for {
		select {
		case <-h.quit:
			h.closeAll()
			log.Printf("level=INFO msg=%q", "websocket hub stopped")
			return
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.userID] == nil {
				h.clients[client.userID] = make(map[*webSocketClient]bool)
			}
			h.clients[client.userID][client] = true
			count := len(h.clients[client.userID])
			total := h.totalConnectionsLocked()
			h.mu.Unlock()
			log.Printf("level=INFO msg=%q userId=%d connections=%d totalConnections=%d",
				"websocket client registered", client.userID, count, total)
		case client := <-h.unregister:
			h.removeClient(client, true)
		case message := <-h.broadcast:
			h.sendToUser(message)
		}
	}
}

func (h *WebSocketHub) Stop() {
	if h == nil {
		return
	}
	h.stopOnce.Do(func() {
		close(h.quit)
	})
}

func (h *WebSocketHub) Register(userID int64, conn *websocket.Conn) *webSocketClient {
	if h == nil || userID <= 0 || conn == nil {
		return nil
	}
	client := &webSocketClient{userID: userID, conn: conn, hub: h}
	h.register <- client
	return client
}

func (h *WebSocketHub) Unregister(userID int64, conn *websocket.Conn) {
	if h == nil || userID <= 0 || conn == nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients[userID] {
		if client.conn == conn {
			h.unregister <- client
			return
		}
	}
}

func (h *WebSocketHub) SendToUser(userID int64, msgType string, data any) error {
	if h == nil || userID <= 0 {
		return nil
	}
	h.broadcast <- &WebSocketMessage{
		UserID: userID,
		Type:   msgType,
		Data:   data,
	}
	return nil
}

func (h *WebSocketHub) Broadcast(msgType string, data any) error {
	if h == nil {
		return nil
	}
	h.mu.RLock()
	userIDs := make([]int64, 0, len(h.clients))
	for userID := range h.clients {
		userIDs = append(userIDs, userID)
	}
	h.mu.RUnlock()
	for _, userID := range userIDs {
		if err := h.SendToUser(userID, msgType, data); err != nil {
			return err
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
	return h.totalConnectionsLocked()
}

func (h *WebSocketHub) totalConnectionsLocked() int {
	total := 0
	for _, conns := range h.clients {
		total += len(conns)
	}
	return total
}

func (h *WebSocketHub) sendToUser(message *WebSocketMessage) {
	if h == nil || message == nil || message.UserID <= 0 {
		return
	}
	h.mu.RLock()
	connections := make([]*webSocketClient, 0, len(h.clients[message.UserID]))
	for client := range h.clients[message.UserID] {
		connections = append(connections, client)
	}
	h.mu.RUnlock()
	if len(connections) == 0 {
		log.Printf("level=DEBUG msg=%q userId=%d", "no websocket connections for user", message.UserID)
		return
	}
	if message.payload == nil {
		payload, err := json.Marshal(message)
		if err != nil {
			log.Printf("level=ERROR msg=%q userId=%d err=%v", "failed to marshal websocket message", message.UserID, err)
			return
		}
		message.payload = payload
	}
	for _, client := range connections {
		if err := client.WriteMessage(websocket.TextMessage, message.payload); err != nil {
			log.Printf("level=ERROR msg=%q userId=%d err=%v", "failed to send websocket message", message.UserID, err)
			go h.unregisterClient(client)
		}
	}
	log.Printf("level=DEBUG msg=%q userId=%d type=%s connections=%d",
		"websocket message sent", message.UserID, message.Type, len(connections))
}

func (h *WebSocketHub) unregisterClient(client *webSocketClient) {
	if h == nil || client == nil {
		return
	}
	select {
	case h.unregister <- client:
	case <-time.After(time.Second):
		h.removeClient(client, true)
	}
}

func (h *WebSocketHub) removeClient(client *webSocketClient, closeConn bool) {
	if h == nil || client == nil {
		return
	}
	h.mu.Lock()
	connections, ok := h.clients[client.userID]
	if !ok || !connections[client] {
		h.mu.Unlock()
		if closeConn {
			_ = client.conn.Close()
		}
		return
	}
	delete(connections, client)
	if len(connections) == 0 {
		delete(h.clients, client.userID)
	}
	count := len(connections)
	h.mu.Unlock()
	if closeConn {
		_ = client.conn.Close()
	}
	log.Printf("level=INFO msg=%q userId=%d connections=%d",
		"websocket client unregistered", client.userID, count)
}

func (h *WebSocketHub) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for userID, conns := range h.clients {
		for client := range conns {
			_ = client.conn.Close()
			delete(conns, client)
		}
		delete(h.clients, userID)
	}
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

func (c *webSocketClient) WriteControl(messageType int, data []byte, deadline time.Time) error {
	if c == nil || c.conn == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteControl(messageType, data, deadline)
}
