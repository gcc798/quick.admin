package handler

import (
	pahoMqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/force-c/nai-tizi/internal/infrastructure/mqtt/protocol"
	logging "github.com/force-c/nai-tizi/internal/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MessageHandler struct {
	db     *gorm.DB
	logger logging.Logger
}

func NewMessageHandler(db *gorm.DB, logger logging.Logger) *MessageHandler {
	return &MessageHandler{
		db:     db,
		logger: logger,
	}
}

// Handle implements mqtt.MessageHandler interface
func (h *MessageHandler) Handle(client pahoMqtt.Client, msg pahoMqtt.Message) {
	if err := h.HandleMessage(msg.Payload()); err != nil {
		h.logger.Error("failed to handle MQTT message", zap.Error(err))
	}
}

// HandleMessage 处理MQTT上行消息
// 这是一个通用示例，记录收到的消息
func (h *MessageHandler) HandleMessage(payload []byte) error {
	// 解析消息
	upMsg, err := protocol.ParseUpMsg(payload)
	if err != nil {
		h.logger.Error("failed to parse MQTT message", zap.Error(err))
		// 如果无法解析为标准协议，也可以选择记录原始payload
		h.logger.Debug("raw payload", zap.ByteString("payload", payload))
		return nil
	}

	h.logger.Debug("received MQTT message",
		zap.Int("optCode", upMsg.OptCode),
		zap.String("msgId", upMsg.MsgId),
		zap.Int64("timestamp", upMsg.Timestamp))

	// TODO: 根据optCode实现业务逻辑
	return nil
}
