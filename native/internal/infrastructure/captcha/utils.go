package captcha

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

// generateRandomCode 生成随机字符验证码（字母+数字）
func generateRandomCode(length int) string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 排除易混淆字符
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[n.Int64()]
	}
	return string(result)
}

// generateRandomDigits 生成随机数字验证码
func generateRandomDigits(length int) string {
	const digits = "0123456789"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		result[i] = digits[n.Int64()]
	}
	return string(result)
}

// isValidPhone 验证手机号格式（中国大陆）
func isValidPhone(phone string) bool {
	pattern := `^1[3-9]\d{9}$`
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}

// isValidEmail 验证邮箱格式
func isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// maskPhone 脱敏手机号
func maskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}

// maskEmail 脱敏邮箱
func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}
	username := parts[0]
	domain := parts[1]

	if len(username) <= 2 {
		return username + "***@" + domain
	}
	return username[:2] + "***@" + domain
}

// formatRedisKey 格式化 Redis 键
func formatRedisKey(captchaType CaptchaType, suffix string) string {
	return fmt.Sprintf("captcha:%s:%s", captchaType, suffix)
}
