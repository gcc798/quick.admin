package protocol

import (
	"fmt"
	"strings"
)

// EdgeDeviceTopicPrefix 服务端下发给自动门设备的 MQTT topic 前缀。
const EdgeDeviceTopicPrefix = "/edge-device/autodoorv1"

// BuildTopic 构造服务端下发给设备的 MQTT topic。
func BuildTopic(netType, mac, sn string) string {
	return fmt.Sprintf("%s/%s/%s", EdgeDeviceTopicPrefix, mac, sn)
}

// ParseTopic 解析服务端下发给设备的 MQTT topic。
func ParseTopic(topic string) (netType, mac, sn string, err error) {
	trimmed := strings.TrimPrefix(topic, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 4 || parts[0] != "edge-device" || parts[1] != "autodoorv1" {
		return "", "", "", fmt.Errorf("invalid topic format: %s", topic)
	}
	return "", parts[2], parts[3], nil
}

// BuildSubscribeTopic 构造服务端下发 topic 的订阅表达式。
func BuildSubscribeTopic(netType, mac, sn string) string {
	if mac == "" {
		return EdgeDeviceTopicPrefix + "/#"
	}
	if sn == "" {
		return fmt.Sprintf("%s/%s/#", EdgeDeviceTopicPrefix, mac)
	}
	return BuildTopic(netType, mac, sn)
}
