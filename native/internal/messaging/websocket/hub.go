package websocket

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	logging "github.com/gcc798/quick.admin/internal/logger"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Hub WebSocket 连接管理中心。
type Hub struct {
	clients    map[int64]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	quit       chan struct{}
	stopOnce   sync.Once
	mu         sync.RWMutex
	logger     logging.Logger
}

// Client WebSocket 客户端。
type Client struct {
	UserId       int64
	Conn         *websocket.Conn
	Hub          *Hub
	writeTimeout time.Duration
	readTimeout  time.Duration
	writeMu      sync.Mutex
	activeSeq    atomic.Uint64
}

// Message 推送消息。
type Message struct {
	UserId  int64       `json:"-"`
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	payload []byte      `json:"-"`
}

// NewHub 创建 WebSocket Hub。
func NewHub(logger logging.Logger) *Hub {
	return &Hub{
		clients:    make(map[int64]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, 256),
		quit:       make(chan struct{}),
		logger:     logger,
	}
}

// Name 返回组件名称。
func (h *Hub) Name() string {
	return "WebSocket Hub"
}

// Start 启动组件。
func (h *Hub) Start() error {
	go h.Run()
	return nil
}

// Stop 停止组件。
func (h *Hub) Stop() error {
	h.stopOnce.Do(func() {
		close(h.quit)
	})
	return nil
}

// Run 启动 Hub 处理循环。
func (h *Hub) Run() {
	for {
		select {
		case <-h.quit:
			h.logger.Info("websocket hub stopped")
			return
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.UserId] == nil {
				h.clients[client.UserId] = make(map[*Client]bool)
			}
			h.clients[client.UserId][client] = true
			h.mu.Unlock()
			h.logger.Info("websocket client registered",
				zap.Int64("userId", client.UserId),
				zap.Int("totalConnections", len(h.clients[client.UserId])))

		case client := <-h.unregister:
			h.mu.Lock()
			if connections, ok := h.clients[client.UserId]; ok {
				if _, exists := connections[client]; exists {
					delete(connections, client)
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

func (h *Hub) sendToUser(message *Message) {
	h.mu.RLock()
	connections, ok := h.clients[message.UserId]
	h.mu.RUnlock()

	if !ok || len(connections) == 0 {
		h.logger.Debug("no websocket connections for user",
			zap.Int64("userId", message.UserId))
		return
	}

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

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range connections {
		if err := client.WriteMessage(websocket.TextMessage, message.payload); err != nil {
			h.logger.Error("failed to send websocket message",
				zap.Int64("userId", message.UserId),
				zap.Error(err))
			go func(c *Client) {
				h.unregister <- c
			}(client)
		}
	}

	h.logger.Debug("websocket message sent",
		zap.Int64("userId", message.UserId),
		zap.String("type", message.Type),
		zap.Int("connections", len(connections)))
}

// SendToUser 发送消息给指定用户。
func (h *Hub) SendToUser(userId int64, msgType string, data interface{}) error {
	message := &Message{
		UserId: userId,
		Type:   msgType,
		Data:   data,
	}
	h.broadcast <- message
	return nil
}

// SendJSONToUser 发送已组装好的 JSON 结构给指定用户。
func (h *Hub) SendJSONToUser(userId int64, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	h.broadcast <- &Message{
		UserId:  userId,
		payload: data,
	}
	return nil
}

// Register 注册客户端连接。
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister 注销客户端连接。
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// GetConnectionCount 获取指定用户的连接数。
func (h *Hub) GetConnectionCount(userId int64) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if connections, ok := h.clients[userId]; ok {
		return len(connections)
	}
	return 0
}

// GetTotalConnections 获取总连接数。
func (h *Hub) GetTotalConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	total := 0
	for _, connections := range h.clients {
		total += len(connections)
	}
	return total
}

// WriteMessage 向客户端写入 WebSocket 消息。
func (c *Client) WriteMessage(messageType int, payload []byte) error {
	return c.writeMessage(messageType, payload, true)
}

func (c *Client) writeMessage(messageType int, payload []byte, refreshReadDeadline bool) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if c.writeTimeout > 0 {
		if err := c.Conn.SetWriteDeadline(time.Now().Add(c.writeTimeout)); err != nil {
			return err
		}
	}
	if err := c.Conn.WriteMessage(messageType, payload); err != nil {
		return err
	}
	if refreshReadDeadline {
		c.markActive()
		return c.refreshReadDeadline()
	}
	return nil
}

// WriteJSON 向客户端写入 JSON 文本消息。
func (c *Client) WriteJSON(payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.WriteMessage(websocket.TextMessage, data)
}

func (c *Client) refreshReadDeadline() error {
	if c.readTimeout <= 0 {
		return nil
	}
	return c.Conn.SetReadDeadline(time.Now().Add(c.readTimeout))
}

func (c *Client) markActive() {
	c.activeSeq.Add(1)
}

func (c *Client) activeSequence() uint64 {
	return c.activeSeq.Load()
}
