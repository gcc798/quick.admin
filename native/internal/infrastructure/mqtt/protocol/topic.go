package protocol

import (
	"fmt"
	"strings"
)

const TopicPrefix = "NTZ"

func BuildTopic(netType, mac, sn string) string {
	return fmt.Sprintf("%s/%s/%s/%s", TopicPrefix, netType, mac, sn)
}
func ParseTopic(topic string) (netType, mac, sn string, err error) {
	parts := strings.Split(topic, "/")
	if len(parts) != 4 || parts[0] != TopicPrefix {
		return "", "", "", fmt.Errorf("invalid topic format: %s", topic)
	}
	return parts[1], parts[2], parts[3], nil
}
func BuildSubscribeTopic(netType, mac, sn string) string {
	if netType == "" {
		return TopicPrefix + "/#"
	}
	if mac == "" {
		return fmt.Sprintf("%s/%s/#", TopicPrefix, netType)
	}
	if sn == "" {
		return fmt.Sprintf("%s/%s/%s/#", TopicPrefix, netType, mac)
	}
	return BuildTopic(netType, mac, sn)
}
