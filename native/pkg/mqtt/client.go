package mqtt

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	logging "github.com/force-c/nai-tizi/internal/logger"
	"go.uber.org/zap"
)

// Subscription 定义业务数据结构。
type Subscription struct {
	Topic   string
	Handler mqtt.MessageHandler
}

// MqttCallback MQTT事件回调
type MqttCallback func(topic string, payload string, err error)

// ConnectCallback 连接事件回调
type ConnectCallback func(err error)

// Client 定义业务数据结构。
type Client struct {
	client           mqtt.Client
	logger           logging.Logger
	broker, clientID string
	qos              byte
	subscriptions    []Subscription
	onConnect        []ConnectCallback
	onPublishSuccess []MqttCallback
	onPublishFailure []MqttCallback
}

// AddOnConnect 添加连接成功回调
func (c *Client) AddOnConnect(cb ConnectCallback) {
	c.onConnect = append(c.onConnect, cb)
}

// AddOnPublishSuccess 添加发布成功回调
func (c *Client) AddOnPublishSuccess(cb MqttCallback) {
	c.onPublishSuccess = append(c.onPublishSuccess, cb)
}

// AddOnPublishFailure 添加发布失败回调
func (c *Client) AddOnPublishFailure(cb MqttCallback) {
	c.onPublishFailure = append(c.onPublishFailure, cb)
}

// Config 定义业务数据结构。
type Config struct {
	Broker, ClientID, Username, Password string
	QoS                                  byte
}

// NewClient 创建组件实例。
func NewClient(cfg *Config, logger logging.Logger) (*Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetClientID(cfg.ClientID)
	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetConnectTimeout(10 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(10 * time.Second)
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) { logger.Warn("MQTT connection lost", zap.Error(err)) })
	opts.SetOnConnectHandler(func(client mqtt.Client) { logger.Info("MQTT connected successfully") })
	c := &Client{client: mqtt.NewClient(opts), logger: logger, broker: cfg.Broker, clientID: cfg.ClientID, qos: cfg.QoS}
	return c, nil
}

// AddSubscription 执行业务逻辑。
func (c *Client) AddSubscription(topic string, handler mqtt.MessageHandler) {
	c.subscriptions = append(c.subscriptions, Subscription{Topic: topic, Handler: handler})
}

// Name 执行业务逻辑。
func (c *Client) Name() string {
	return "MQTT Client"
}

// Start 启动组件。
func (c *Client) Start() error {
	if err := c.Connect(); err != nil {
		return err
	}
	for _, sub := range c.subscriptions {
		if err := c.Subscribe(sub.Topic, sub.Handler); err != nil {
			return err
		}
	}
	return nil
}

// Stop 停止组件。
func (c *Client) Stop() error {
	c.Disconnect()
	return nil
}

// Connect 执行业务逻辑。
func (c *Client) Connect() error {
	token := c.client.Connect()
	if token.Wait() && token.Error() != nil {
		err := fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
		for _, cb := range c.onConnect {
			cb(err)
		}
		return err
	}
	c.logger.Info("MQTT client connected successfully")
	for _, cb := range c.onConnect {
		cb(nil)
	}
	return nil
}

// Disconnect 执行业务逻辑。
func (c *Client) Disconnect() {
	c.logger.Info("disconnecting from MQTT broker")
	c.client.Disconnect(250)
}

// IsConnected 执行业务逻辑。
func (c *Client) IsConnected() bool { return c.client.IsConnected() }

// Subscribe 执行业务逻辑。
func (c *Client) Subscribe(topic string, callback mqtt.MessageHandler) error {
	token := c.client.Subscribe(topic, c.qos, callback)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, token.Error())
	}
	return nil
}

// Unsubscribe 执行业务逻辑。
func (c *Client) Unsubscribe(topic string) error {
	token := c.client.Unsubscribe(topic)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to unsubscribe from topic %s: %w", topic, token.Error())
	}
	return nil
}

// Publish 执行业务逻辑。
func (c *Client) Publish(topic string, payload string) error {
	if !c.IsConnected() {
		err := fmt.Errorf("MQTT client not connected")
		for _, cb := range c.onPublishFailure {
			cb(topic, payload, err)
		}
		return err
	}
	token := c.client.Publish(topic, c.qos, false, payload)
	if token.Wait() && token.Error() != nil {
		err := fmt.Errorf("failed to publish message: %w", token.Error())
		for _, cb := range c.onPublishFailure {
			cb(topic, payload, err)
		}
		return err
	}

	for _, cb := range c.onPublishSuccess {
		cb(topic, payload, nil)
	}
	return nil
}

// PublishControl 执行业务逻辑。
func (c *Client) PublishControl(netType, mac, sn string, payload string) error {
	topic := fmt.Sprintf("/edge-device/autodoorv1/%s/%s", mac, sn)
	return c.Publish(topic, payload)
}
