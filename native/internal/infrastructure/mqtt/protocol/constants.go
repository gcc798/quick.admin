package protocol

// 协议说明：使用双向操作码，即上行和下行使用相同的操作码
// 1001: 心跳 (heart) - 双向：下行心跳响应/上行心跳上报
// 1002: 解绑 (unbind) - 双向：下行解绑指令/上行解绑响应
// 1003: 模式控制 (modeControl) - 双向：下行控制指令/上行控制响应
// 1004: 报警 (releaseAlarm/AlarmTrigger) - 双向：下行报警解除/上行报警触发
// 1005: 设备对码 (deviceCodeMatch) - 双向：下行对码指令/上行对码响应
// 1006: 删除对码 (deviceCodeDelete) - 双向：下行删除对码指令/上行删除对码响应
// 1007: 门连接事件 (doorConnectEvent) - 单向上行：门设备连接状态事件
// 8000: 查询设备信息 (queryDeviceInfo) - 双向：下行查询指令/上行设备信息响应
const (
	OptCodeHeart            = 1001 // 心跳 - 双向
	OptCodeUnbind           = 1002 // 解绑 - 双向
	OptCodeModeControl      = 1003 // 模式控制 - 双向
	OptCodeAlarm            = 1004 // 报警解除/触发 - 双向
	OptCodeDeviceCodeMatch  = 1005 // 设备对码 - 双向
	OptCodeDeviceCodeDelete = 1006 // 删除对码 - 双向
	OptCodeDoorConnectEvent = 1007 // 门连接事件 - 上行
	OptCodeQueryDeviceInfo  = 8000 // 查询设备信息 - 双向
)

// 设备网络类型（用于构造MQTT主题）
const (
	NetTypeWiFi     = "wifi"     // Wi-Fi
	NetType4G       = "4g"       // 蜂窝网络
	NetTypeEthernet = "ethernet" // 以太网
)

// 门状态（控制/上报状态值）
const (
	DoorStateOpen     = "OPEN"      // 开门
	DoorStateClose    = "CLOSE"     // 关门
	DoorStateHalfOpen = "HALF_OPEN" // 半开
	DoorStateAuto     = "AUTO"      // 自动模式
)

// 设备在线状态（缓存/监控使用）
const (
	DeviceStateOnline  = "online"  // 在线
	DeviceStateOffline = "offline" // 离线
)

// 门连接状态（缓存值）
const (
	DoorDisconnected = 0 // 断开
	DoorConnected    = 1 // 已连接
)
