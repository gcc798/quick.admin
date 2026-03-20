package rabbitmq

import (
	"fmt"

	logging "github.com/force-c/nai-tizi/internal/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Manager RabbitMQ管理器，统一管理所有RabbitMQ相关组件
type Manager struct {
	conn     *Connection
	producer *ProducerService
	logger   logging.Logger
}

// NewManager 创建RabbitMQ管理器
func NewManager(config *Config, db *gorm.DB, logger logging.Logger) (*Manager, error) {
	// 创建连接
	conn, err := NewConnection(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create RabbitMQ connection: %w", err)
	}

	// 创建生产者服务
	producer, err := NewProducerService(conn, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create RabbitMQ producer: %w", err)
	}

	return &Manager{
		conn:     conn,
		producer: producer,
		logger:   logger,
	}, nil
}

// GetConnection 获取连接
func (m *Manager) GetConnection() *Connection {
	return m.conn
}

// GetProducer 获取生产者服务
func (m *Manager) GetProducer() *ProducerService {
	return m.producer
}

func (m *Manager) Name() string {
	return "RabbitMQ Manager"
}

// Start 启动RabbitMQ服务
func (m *Manager) Start() error {
	m.logger.Info("RabbitMQ manager started successfully")
	return nil
}

// Stop 停止RabbitMQ服务（关闭所有组件）
func (m *Manager) Stop() error {
	m.logger.Info("stopping RabbitMQ manager...")

	// 关闭生产者
	if m.producer != nil {
		if err := m.producer.Close(); err != nil {
			m.logger.Error("failed to close producer", zap.Error(err))
		}
	}

	// 关闭连接
	if m.conn != nil {
		if err := m.conn.Close(); err != nil {
			m.logger.Error("failed to close connection", zap.Error(err))
			return err
		}
	}

	m.logger.Info("RabbitMQ manager stopped")
	return nil
}
