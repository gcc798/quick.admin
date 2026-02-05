package rabbitmq

import (
	"context"
	"fmt"
	"time"

	logging "github.com/force-c/nai-tizi/internal/logger"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// PublishCallback 发布消息结果回调
type PublishCallback func(ctx context.Context, exchange, routingKey, messageID string, err error)

// ProducerService RabbitMQ生产者服务
// 使用常量定义的交换机和路由键，不需要配置生产者
// 支持同步、异步、单向、延时消息发送
type ProducerService struct {
	conn             *Connection
	channel          *amqp.Channel
	logger           logging.Logger
	onPublishSuccess []PublishCallback
	onPublishFailure []PublishCallback
}

// AddOnPublishSuccess 添加发布成功回调
func (p *ProducerService) AddOnPublishSuccess(cb PublishCallback) {
	p.onPublishSuccess = append(p.onPublishSuccess, cb)
}

// AddOnPublishFailure 添加发布失败回调
func (p *ProducerService) AddOnPublishFailure(cb PublishCallback) {
	p.onPublishFailure = append(p.onPublishFailure, cb)
}

// NewProducerService 创建生产者服务
func NewProducerService(conn *Connection, logger logging.Logger) (*ProducerService, error) {
	if !conn.IsEnabled() {
		logger.Info("RabbitMQ未启用，生产者服务将不创建通道")
		return &ProducerService{conn: conn, logger: logger}, nil
	}

	ch, err := conn.CreateChannel()
	if err != nil {
		return nil, fmt.Errorf("创建生产者通道失败: %w", err)
	}

	return &ProducerService{
		conn:    conn,
		channel: ch,
		logger:  logger,
	}, nil
}

// SendSync 同步发送消息（使用指定交换机）
func (p *ProducerService) SendSync(exchange, routingKey, messageID, body string) (bool, error) {
	if !p.conn.IsEnabled() {
		p.logger.Debug("RabbitMQ未启用，跳过发送消息")
		return false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(ConfirmTimeout)*time.Second)
	defer cancel()

	confirms := p.channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	err := p.channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			MessageId:    messageID,
			DeliveryMode: amqp.Persistent, // 持久化
			ContentType:  "application/json",
			Body:         []byte(body),
		},
	)

	if err != nil {
		p.logger.Error("发送同步消息失败",
			zap.String("exchange", exchange),
			zap.String("routingKey", routingKey),
			zap.String("messageId", messageID),
			zap.Error(err))

		// 触发失败回调
		for _, cb := range p.onPublishFailure {
			cb(ctx, exchange, routingKey, messageID, err)
		}

		return false, err
	}

	// 等待确认
	select {
	case confirm := <-confirms:
		if confirm.Ack {
			p.logger.Info("发送同步消息成功",
				zap.String("exchange", exchange),
				zap.String("routingKey", routingKey),
				zap.String("messageId", messageID))

			// 触发成功回调
			for _, cb := range p.onPublishSuccess {
				cb(ctx, exchange, routingKey, messageID, nil)
			}

			return true, nil
		}
		p.logger.Error("发送同步消息失败 - 未确认",
			zap.String("exchange", exchange),
			zap.String("routingKey", routingKey),
			zap.String("messageId", messageID))

		err := fmt.Errorf("消息未被确认")
		// 触发失败回调
		for _, cb := range p.onPublishFailure {
			cb(ctx, exchange, routingKey, messageID, err)
		}

		return false, err
	case <-ctx.Done():
		p.logger.Error("发送同步消息超时",
			zap.String("exchange", exchange),
			zap.String("routingKey", routingKey),
			zap.String("messageId", messageID))

		err := fmt.Errorf("等待确认超时")
		// 触发失败回调
		for _, cb := range p.onPublishFailure {
			cb(ctx, exchange, routingKey, messageID, err)
		}

		return false, err
	}
}

// SendAsync 异步发送消息（使用指定交换机）
func (p *ProducerService) SendAsync(exchange, routingKey, messageID, body string) error {
	if !p.conn.IsEnabled() {
		p.logger.Debug("RabbitMQ未启用，跳过发送消息")
		return nil
	}

	err := p.channel.Publish(
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			MessageId:    messageID,
			DeliveryMode: amqp.Persistent, // 持久化
			ContentType:  "application/json",
			Body:         []byte(body),
		},
	)

	if err != nil {
		p.logger.Error("发送异步消息失败",
			zap.String("exchange", exchange),
			zap.String("routingKey", routingKey),
			zap.String("messageId", messageID),
			zap.Error(err))
		return err
	}

	p.logger.Info("发送异步消息已提交",
		zap.String("exchange", exchange),
		zap.String("routingKey", routingKey),
		zap.String("messageId", messageID))
	return nil
}

// SendDelaySync 发送延时消息（使用指定交换机）
// 使用RabbitMQ的TTL + DLX实现延时消息
func (p *ProducerService) SendDelaySync(targetExchange, routingKey, messageID, body string, delaySeconds int) (bool, error) {
	if !p.conn.IsEnabled() {
		p.logger.Debug("RabbitMQ未启用，跳过发送延时消息")
		return false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(ConfirmTimeout)*time.Second)
	defer cancel()

	// 构建延时队列名称
	delayQueue := fmt.Sprintf("delay.%s.%s", targetExchange, routingKey)

	// 声明延时队列（带DLX配置）
	_, err := p.channel.QueueDeclare(
		delayQueue,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		amqp.Table{
			"x-dead-letter-exchange":    targetExchange,
			"x-dead-letter-routing-key": routingKey,
		},
	)
	if err != nil {
		p.logger.Error("声明延时队列失败",
			zap.String("queue", delayQueue),
			zap.Error(err))
		return false, fmt.Errorf("声明延时队列失败: %w", err)
	}

	confirms := p.channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	// 发送到延时队列
	err = p.channel.PublishWithContext(
		ctx,
		"",         // exchange（直接发送到队列）
		delayQueue, // routing key（队列名）
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			MessageId:    messageID,
			DeliveryMode: amqp.Persistent, // 持久化
			ContentType:  "application/json",
			Expiration:   fmt.Sprintf("%d", delaySeconds*1000), // TTL（毫秒）
			Body:         []byte(body),
		},
	)

	if err != nil {
		p.logger.Error("发送延时消息失败",
			zap.String("routingKey", routingKey),
			zap.String("messageId", messageID),
			zap.Int("delaySeconds", delaySeconds),
			zap.Error(err))
		return false, err
	}

	// 等待确认
	select {
	case confirm := <-confirms:
		if confirm.Ack {
			p.logger.Info("发送延时消息成功",
				zap.String("routingKey", routingKey),
				zap.String("messageId", messageID),
				zap.Int("delaySeconds", delaySeconds))
			return true, nil
		}
		p.logger.Error("发送延时消息失败 - 未确认",
			zap.String("routingKey", routingKey),
			zap.String("messageId", messageID),
			zap.Int("delaySeconds", delaySeconds))
		return false, fmt.Errorf("消息未被确认")
	case <-ctx.Done():
		p.logger.Error("发送延时消息超时",
			zap.String("routingKey", routingKey),
			zap.String("messageId", messageID),
			zap.Int("delaySeconds", delaySeconds))
		return false, fmt.Errorf("等待确认超时")
	}
}

// SendDelayDeviceTimeoutCheck 发送延时设备超时检测消息
// 便捷方法：使用预定义的交换机和路由键
func (p *ProducerService) SendDelayDeviceTimeoutCheck(messageID, body string, delaySeconds int) (bool, error) {
	return p.SendDelaySync(
		DeviceTimeoutExchange,
		DeviceTimeoutRoutingKey,
		messageID,
		body,
		delaySeconds,
	)
}

// Close 关闭生产者服务
func (p *ProducerService) Close() error {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			p.logger.Error("关闭生产者通道失败", zap.Error(err))
			return err
		}
		p.logger.Info("生产者通道已关闭")
	}
	return nil
}
