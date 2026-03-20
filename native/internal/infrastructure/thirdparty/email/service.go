package email

import (
	"crypto/tls"
	"fmt"
	"time"

	logging "github.com/force-c/nai-tizi/internal/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

const (
	EmailCodeLength   = 6
	EmailCodeExpire   = 5 * time.Minute
	EmailCodeRedisKey = "global:captcha_codes:email:"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type Manager struct {
	config Config
	redis  *redis.Client
	logger logging.Logger
}

func NewManager(config Config, redis *redis.Client, logger logging.Logger) (*Manager, error) {
	return &Manager{
		config: config,
		redis:  redis,
		logger: logger,
	}, nil
}

func (m *Manager) Send(to string, subject string, body string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.config.From)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)

	dialer := gomail.NewDialer(m.config.Host, m.config.Port, m.config.Username, m.config.Password)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := dialer.DialAndSend(msg); err != nil {
		m.logger.Error("failed to send email", zap.Error(err))
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	m.logger.Info("email sent successfully", zap.String("to", to))
	return nil
}

func (m *Manager) SendWithTemplate(to string, code string, template string) error {
	subject := "验证码"
	body := fmt.Sprintf(template, code)
	return m.Send(to, subject, body)
}
