package websocket

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	logging "github.com/gcc798/nai-tizi/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 生产环境应该检查 Origin。
	},
}

// TextMessageHandler 处理客户端文本消息。
type TextMessageHandler func(client *Client, message []byte) bool

// HeartbeatMessageBuilder 构建服务端主动心跳消息。
type HeartbeatMessageBuilder func(client *Client) interface{}

// Handler WebSocket HTTP 处理器。
type Handler struct {
	hub                 *Hub
	logger              logging.Logger
	onTextMessage       TextMessageHandler
	readTimeout         time.Duration
	writeTimeout        time.Duration
	maxReadTimeouts     int
	heartbeatMsgBuilder HeartbeatMessageBuilder
}

// NewHandler 创建 WebSocket 处理器。
func NewHandler(hub *Hub, logger logging.Logger) *Handler {
	return &Handler{
		hub:    hub,
		logger: logger,
	}
}

// RegisterTextMessageHandler 注册客户端文本消息处理器。
func (h *Handler) RegisterTextMessageHandler(handler TextMessageHandler) {
	h.onTextMessage = handler
}

// ConfigureTimeouts 配置 WebSocket 读写超时时间。
func (h *Handler) ConfigureTimeouts(readTimeout, writeTimeout time.Duration) {
	h.readTimeout = readTimeout
	h.writeTimeout = writeTimeout
}

// RegisterHeartbeatMessageBuilder 注册服务端主动心跳消息构建函数。
func (h *Handler) RegisterHeartbeatMessageBuilder(maxReadTimeouts int, builder HeartbeatMessageBuilder) {
	h.maxReadTimeouts = maxReadTimeouts
	h.heartbeatMsgBuilder = builder
}

// ServeWs 处理 WebSocket 连接请求。
func (h *Handler) ServeWs(c *gin.Context) {
	if h.hub == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "websocket disabled"})
		return
	}

	// 从查询参数或 JWT 中获取用户 ID。
	userIdStr := c.Query("userId")
	if userIdStr == "" {
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

	// 升级 HTTP 连接为 WebSocket。
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("failed to upgrade websocket connection",
			zap.Int64("userId", userId),
			zap.Error(err))
		return
	}

	// 创建客户端并注册。
	client := &Client{
		UserId:       userId,
		Conn:         conn,
		Hub:          h.hub,
		writeTimeout: h.writeTimeout,
		readTimeout:  h.readTimeout,
	}
	client.markActive()

	if err := client.refreshReadDeadline(); err != nil {
		h.logger.Error("failed to set websocket read deadline",
			zap.Int64("userId", userId),
			zap.Error(err))
		conn.Close()
		return
	}

	h.hub.Register(client)

	// 启动读取协程，保持连接并处理客户端消息。
	go h.readPump(client)

	h.logger.Info("websocket connection established",
		zap.Int64("userId", userId),
		zap.String("remoteAddr", c.Request.RemoteAddr))
}

// readPump 读取客户端消息并保持连接。
func (h *Handler) readPump(client *Client) {
	defer func() {
		h.hub.Unregister(client)
	}()

	readTimeouts := 0
	lastActivitySeq := client.activeSequence()
	for {
		messageType, message, err := client.Conn.ReadMessage()
		if err != nil {
			if isTimeoutError(err) && h.heartbeatMsgBuilder != nil {
				if currentSeq := client.activeSequence(); currentSeq != lastActivitySeq {
					readTimeouts = 0
					lastActivitySeq = currentSeq
					if err := client.refreshReadDeadline(); err != nil {
						h.logger.Error("failed to refresh websocket read deadline",
							zap.Int64("userId", client.UserId),
							zap.Error(err))
						break
					}
					continue
				}
				readTimeouts++
				if h.maxReadTimeouts > 0 && readTimeouts >= h.maxReadTimeouts {
					h.logger.Warn("websocket read timeout limit reached",
						zap.Int64("userId", client.UserId),
						zap.Int("readTimeouts", readTimeouts))
					break
				}
				if err := h.sendHeartbeat(client); err != nil {
					h.logger.Error("failed to send websocket heartbeat",
						zap.Int64("userId", client.UserId),
						zap.Error(err))
					break
				}
				if err := client.refreshReadDeadline(); err != nil {
					h.logger.Error("failed to refresh websocket read deadline",
						zap.Int64("userId", client.UserId),
						zap.Error(err))
					break
				}
				continue
			}
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("websocket read error",
					zap.Int64("userId", client.UserId),
					zap.Error(err))
			}
			break
		}

		client.markActive()
		lastActivitySeq = client.activeSequence()
		readTimeouts = 0
		if err := client.refreshReadDeadline(); err != nil {
			h.logger.Error("failed to refresh websocket read deadline",
				zap.Int64("userId", client.UserId),
				zap.Error(err))
			break
		}

		if messageType != websocket.TextMessage {
			h.logger.Debug("received non-text websocket message",
				zap.Int64("userId", client.UserId),
				zap.Int("messageType", messageType))
			continue
		}

		if h.onTextMessage != nil && h.onTextMessage(client, message) {
			continue
		}

		h.logger.Debug("received websocket message from client",
			zap.Int64("userId", client.UserId),
			zap.ByteString("message", message))
	}
}

func (h *Handler) sendHeartbeat(client *Client) error {
	payload := h.heartbeatMsgBuilder(client)
	if payload == nil {
		return nil
	}
	switch typed := payload.(type) {
	case []byte:
		return client.writeMessage(websocket.TextMessage, typed, false)
	case string:
		return client.writeMessage(websocket.TextMessage, []byte(typed), false)
	default:
		data, err := json.Marshal(typed)
		if err != nil {
			return err
		}
		return client.writeMessage(websocket.TextMessage, data, false)
	}
}

func isTimeoutError(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
