package websocket

import (
	"net/http"
	"strconv"

	logging "github.com/force-c/nai-tizi/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 生产环境应该检查Origin
	},
}

// Handler WebSocket HTTP处理器
type Handler struct {
	hub    *Hub
	logger logging.Logger
}

// NewHandler 创建WebSocket处理器
func NewHandler(hub *Hub, logger logging.Logger) *Handler {
	return &Handler{
		hub:    hub,
		logger: logger,
	}
}

// ServeWs 处理WebSocket连接请求
func (h *Handler) ServeWs(c *gin.Context) {
	// 从查询参数或JWT中获取用户ID
	userIdStr := c.Query("userId")
	if userIdStr == "" {
		// 尝试从JWT token中获取
		if userId, exists := c.Get("userId"); exists {
			if uid, ok := userId.(int64); ok {
				userIdStr = strconv.FormatInt(uid, 10)
			}
		}
	}

	if userIdStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing userId"})
		return
	}

	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId"})
		return
	}

	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("failed to upgrade websocket connection",
			zap.Int64("userId", userId),
			zap.Error(err))
		return
	}

	// 创建客户端并注册
	client := &Client{
		UserId: userId,
		Conn:   conn,
		Hub:    h.hub,
	}

	h.hub.Register(client)

	// 启动读取协程（保持连接并处理客户端消息）
	go h.readPump(client)

	h.logger.Info("websocket connection established",
		zap.Int64("userId", userId),
		zap.String("remoteAddr", c.Request.RemoteAddr))
}

// readPump 读取客户端消息并保持连接
func (h *Handler) readPump(client *Client) {
	defer func() {
		h.hub.Unregister(client)
	}()

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("websocket read error",
					zap.Int64("userId", client.UserId),
					zap.Error(err))
			}
			break
		}

		// 处理客户端发送的消息（如心跳ping/pong等）
		h.logger.Debug("received websocket message from client",
			zap.Int64("userId", client.UserId),
			zap.ByteString("message", message))

		// 可以在这里处理客户端发送的心跳或其他消息
		// 目前仅作为保持连接的机制
	}
}
