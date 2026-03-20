package rabbitmq

// RabbitMQ消息发送常量定义
// 为生产者提供标准化的交换机名称和路由键定义
// 生产者无需配置，直接使用这些常量发送消息
// 当前仅包含设备超时检测业务相关常量

const (
	// DeviceTimeoutExchange 设备超时检测交换机
	DeviceTimeoutExchange = "device.direct"

	// DeviceTimeoutRoutingKey 设备超时检测路由键
	DeviceTimeoutRoutingKey = "timeout.check"

	// BusinessDeviceTimeout 设备超时检测业务标识
	BusinessDeviceTimeout = "device-timeout"

	// DeviceTimeoutQueue 设备超时检测队列名称
	DeviceTimeoutQueue = "queue.device.timeout.check"

	// RetryCountHeader 重试次数header键
	RetryCountHeader = "x-retry-count"

	// MaxRetryCount 最大重试次数
	MaxRetryCount = 3

	// ConfirmTimeout 消息确认超时时间（秒）
	ConfirmTimeout = 6
)
