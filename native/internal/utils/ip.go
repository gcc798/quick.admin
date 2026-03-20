package utils

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetClientIP 获取客户端真实IP地址
// 优先从 X-Forwarded-For 和 X-Real-IP 获取，如果都没有则使用 ClientIP()
// 如果是 IPv6 的 localhost (::1)，则转换为 IPv4 的 localhost (127.0.0.1)
func GetClientIP(c *gin.Context) string {
	ip := c.ClientIP()

	// 如果是 IPv6 的 localhost，转换为 IPv4
	if ip == "::1" {
		return "127.0.0.1"
	}

	// 如果是 IPv6 格式的 localhost，也转换为 IPv4
	if ip == "[::1]" {
		return "127.0.0.1"
	}

	// 尝试从 X-Forwarded-For 获取真实IP
	forwardedFor := c.GetHeader("X-Forwarded-For")
	if forwardedFor != "" {
		// X-Forwarded-For 可能包含多个IP，取第一个
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			ip = strings.TrimSpace(ips[0])
			// 验证IP格式
			if net.ParseIP(ip) != nil {
				if ip == "::1" || ip == "[::1]" {
					return "127.0.0.1"
				}
				return ip
			}
		}
	}

	// 尝试从 X-Real-IP 获取真实IP
	realIP := c.GetHeader("X-Real-IP")
	if realIP != "" {
		ip = strings.TrimSpace(realIP)
		// 验证IP格式
		if net.ParseIP(ip) != nil {
			if ip == "::1" || ip == "[::1]" {
				return "127.0.0.1"
			}
			return ip
		}
	}

	return ip
}
