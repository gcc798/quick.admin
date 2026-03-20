package retry

import (
	"strconv"
)

// RetryMessageRequest 重试消息请求对象
type RetryMessageRequest struct {
	MessageId      string // 消息ID（业务记录ID）
	DeviceId       int64  // 设备ID
	DeviceMac      string // 设备MAC
	DeviceSnNum    string // 设备序列号
	MessageType    int    // 消息类型（对应OptCodeEnum）
	MessageContent string // MQTT消息内容
	MaxRetryCount  *int   // 最大重试次数
	RetryInterval  *int64 // 重试间隔（毫秒）
}

// FromRedisCache 从Redis缓存数据构建请求对象
func FromRedisCache(messageId string, retryData map[string]interface{}) (*RetryMessageRequest, error) {
	request := &RetryMessageRequest{
		MessageId:      messageId,
		DeviceMac:      safeGetString(retryData, "deviceMac"),
		DeviceSnNum:    safeGetString(retryData, "deviceSnNum"),
		MessageContent: safeGetString(retryData, "messageContent"),
	}

	// 解析 deviceId
	if deviceIdStr := safeGetString(retryData, "deviceId"); deviceIdStr != "" {
		deviceId, err := strconv.ParseInt(deviceIdStr, 10, 64)
		if err != nil {
			return nil, err
		}
		request.DeviceId = deviceId
	}

	// 解析 messageType
	if messageTypeStr := safeGetString(retryData, "messageType"); messageTypeStr != "" {
		messageType, err := strconv.Atoi(messageTypeStr)
		if err != nil {
			return nil, err
		}
		request.MessageType = messageType
	}

	// 解析 maxRetryCount
	if val := safeGetInt(retryData, "maxRetryCount"); val != nil {
		request.MaxRetryCount = val
	}

	// 解析 retryInterval
	if val := safeGetInt64(retryData, "retryInterval"); val != nil {
		request.RetryInterval = val
	}

	return request, nil
}

// safeGetString 安全获取String值
func safeGetString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
		// 如果不是字符串类型，尝试转换
		return strconv.FormatInt(toInt64(val), 10)
	}
	return ""
}

// safeGetInt 安全获取Int值
func safeGetInt(data map[string]interface{}, key string) *int {
	if val, ok := data[key]; ok && val != nil {
		intVal := int(toInt64(val))
		return &intVal
	}
	return nil
}

// safeGetInt64 安全获取Int64值
func safeGetInt64(data map[string]interface{}, key string) *int64 {
	if val, ok := data[key]; ok && val != nil {
		int64Val := toInt64(val)
		return &int64Val
	}
	return nil
}

// toInt64 转换为int64
func toInt64(val interface{}) int64 {
	switch v := val.(type) {
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
	}
	return 0
}
