package rabbitmq

import (
	"fmt"
	"time"

	logging "github.com/force-c/nai-tizi/internal/logger"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// Config RabbitMQ配置
type Config struct {
	URL      string `mapstructure:"url"`      // RabbitMQ连接URL
	Exchange string `mapstructure:"exchange"` // 交换机名称
	Enabled  bool   `mapstructure:"enabled"`  // 是否启用RabbitMQ
}

// Connection RabbitMQ连接管理器
type Connection struct {
	config *Config
	conn   *amqp.Connection
	logger logging.Logger
}

// NewConnection 创建RabbitMQ连接
func NewConnection(config *Config, logger logging.Logger) (*Connection, error) {
	if !config.Enabled {
		logger.Info("RabbitMQ已禁用")
		return &Connection{config: config, logger: logger}, nil
	}

	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("连接RabbitMQ失败: %w", err)
	}

	logger.Info("RabbitMQ连接成功", zap.String("url", maskPassword(config.URL)))

	c := &Connection{
		config: config,
		conn:   conn,
		logger: logger,
	}

	// 监听连接关闭事件
	go c.handleConnectionClose()

	return c, nil
}

// GetConnection 获取RabbitMQ连接
func (c *Connection) GetConnection() *amqp.Connection {
	return c.conn
}

// IsEnabled 检查是否启用
func (c *Connection) IsEnabled() bool {
	return c.config.Enabled
}

// CreateChannel 创建通道
func (c *Connection) CreateChannel() (*amqp.Channel, error) {
	if !c.config.Enabled || c.conn == nil {
		return nil, fmt.Errorf("RabbitMQ未启用或未连接")
	}

	ch, err := c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("创建RabbitMQ通道失败: %w", err)
	}

	// 启用Publisher Confirms模式
	if err := ch.Confirm(false); err != nil {
		ch.Close()
		return nil, fmt.Errorf("启用Publisher Confirms失败: %w", err)
	}

	return ch, nil
}

// Close 关闭连接
func (c *Connection) Close() error {
	if c.conn != nil && !c.conn.IsClosed() {
		if err := c.conn.Close(); err != nil {
			c.logger.Error("关闭RabbitMQ连接失败", zap.Error(err))
			return err
		}
		c.logger.Info("RabbitMQ连接已关闭")
	}
	return nil
}

// handleConnectionClose 处理连接关闭事件
func (c *Connection) handleConnectionClose() {
	if c.conn == nil {
		return
	}

	closeErr := make(chan *amqp.Error)
	c.conn.NotifyClose(closeErr)

	err := <-closeErr
	if err != nil {
		c.logger.Error("RabbitMQ连接异常关闭",
			zap.Int("code", err.Code),
			zap.String("reason", err.Reason))

		// 可以在这里实现重连逻辑
		c.reconnect()
	}
}

// reconnect 重新连接（简单实现）
func (c *Connection) reconnect() {
	if !c.config.Enabled {
		return
	}

	for i := 0; i < 5; i++ {
		c.logger.Info("尝试重新连接RabbitMQ", zap.Int("attempt", i+1))
		time.Sleep(time.Duration(i+1) * time.Second)

		conn, err := amqp.Dial(c.config.URL)
		if err != nil {
			c.logger.Error("重连RabbitMQ失败", zap.Error(err))
			continue
		}

		c.conn = conn
		c.logger.Info("RabbitMQ重连成功")
		go c.handleConnectionClose()
		return
	}

	c.logger.Error("RabbitMQ重连失败，已达最大重试次数")
}

// maskPassword 隐藏密码
func maskPassword(url string) string {
	// 简单实现：查找@符号前的最后一个:，替换密码部分
	if len(url) < 10 {
		return "***"
	}
	return url[:10] + "***"
}
