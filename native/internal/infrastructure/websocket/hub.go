package websocket

import (
	"encoding/json"
	"sync"

	logging "github.com/force-c/nai-tizi/internal/logger"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Hub WebSocket连接管理中心
type Hub struct {
	clients    map[int64]map[*websocket.Conn]bool // userId -> connections
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	quit       chan struct{}
	stopOnce   sync.Once
	mu         sync.RWMutex
	logger     logging.Logger
}

// Client WebSocket客户端
type Client struct {
	UserId int64
	Conn   *websocket.Conn
	Hub    *Hub
}

// Message 推送消息
type Message struct {
	UserId  int64       `json:"-"`    // 目标用户ID
	Type    string      `json:"type"` // 消息类型：control_resp/match_code_resp/delete_code_resp/device_status等
	Data    interface{} `json:"data"` // 消息数据
	payload []byte      `json:"-"`    // 序列化后的数据
}

// NewHub 创建WebSocket Hub
func NewHub(logger logging.Logger) *Hub {
	return &Hub{
		clients:    make(map[int64]map[*websocket.Conn]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, 256),
		quit:       make(chan struct{}),
		logger:     logger,
	}
}

func (h *Hub) Name() string {
	return "WebSocket Hub"
}

func (h *Hub) Start() error {
	go h.Run()
	return nil
}

func (h *Hub) Stop() error {
	h.stopOnce.Do(func() {
		close(h.quit)
	})
	return nil
}

// Run 启动Hub处理循环
func (h *Hub) Run() {
	for {
		select {
		case <-h.quit:
			h.logger.Info("websocket hub stopped")
			return
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.UserId] == nil {
				h.clients[client.UserId] = make(map[*websocket.Conn]bool)
			}
			h.clients[client.UserId][client.Conn] = true
			h.mu.Unlock()
			h.logger.Info("websocket client registered",
				zap.Int64("userId", client.UserId),
				zap.Int("totalConnections", len(h.clients[client.UserId])))

		case client := <-h.unregister:
			h.mu.Lock()
			if connections, ok := h.clients[client.UserId]; ok {
				if _, exists := connections[client.Conn]; exists {
					delete(connections, client.Conn)
					client.Conn.Close()
					if len(connections) == 0 {
						delete(h.clients, client.UserId)
					}
					h.logger.Info("websocket client unregistered",
						zap.Int64("userId", client.UserId))
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.sendToUser(message)
		}
	}
}

// sendToUser 发送消息给指定用户的所有连接
func (h *Hub) sendToUser(message *Message) {
	h.mu.RLock()
	connections, ok := h.clients[message.UserId]
	h.mu.RUnlock()

	if !ok || len(connections) == 0 {
		h.logger.Debug("no websocket connections for user",
			zap.Int64("userId", message.UserId))
		return
	}

	// 序列化消息（只序列化一次）
	if message.payload == nil {
		payload, err := json.Marshal(message)
		if err != nil {
			h.logger.Error("failed to marshal websocket message",
				zap.Int64("userId", message.UserId),
				zap.Error(err))
			return
		}
		message.payload = payload
	}

	// 发送给该用户的所有连接
	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn := range connections {
		if err := conn.WriteMessage(websocket.TextMessage, message.payload); err != nil {
			h.logger.Error("failed to send websocket message",
				zap.Int64("userId", message.UserId),
				zap.Error(err))
			// 标记为需要注销
			go func(c *Client) {
				h.unregister <- c
			}(&Client{UserId: message.UserId, Conn: conn, Hub: h})
		}
	}

	h.logger.Debug("websocket message sent",
		zap.Int64("userId", message.UserId),
		zap.String("type", message.Type),
		zap.Int("connections", len(connections)))
}

// SendToUser 发送消息给指定用户（外部调用接口）
func (h *Hub) SendToUser(userId int64, msgType string, data interface{}) error {
	message := &Message{
		UserId: userId,
		Type:   msgType,
		Data:   data,
	}
	h.broadcast <- message
	return nil
}

// Register 注册客户端连接
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister 注销客户端连接
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// GetConnectionCount 获取指定用户的连接数
func (h *Hub) GetConnectionCount(userId int64) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if connections, ok := h.clients[userId]; ok {
		return len(connections)
	}
	return 0
}

// GetTotalConnections 获取总连接数
func (h *Hub) GetTotalConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	total := 0
	for _, connections := range h.clients {
		total += len(connections)
	}
	return total
}
