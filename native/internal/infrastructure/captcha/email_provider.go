package captcha

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// EmailCaptchaConfig 邮箱验证码配置
type EmailCaptchaConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Length   int    `yaml:"length"`
	Expire   int    `yaml:"expire"` // 秒
	Template string `yaml:"template"`
}

// EmailProvider 邮件服务提供者接口
type EmailProvider interface {
	SendEmail(email string, code string, template string) error
}

// EmailCaptchaProvider 邮箱验证码提供者
type EmailCaptchaProvider struct {
	config        *EmailCaptchaConfig
	redis         *redis.Client
	emailProvider EmailProvider
}

// NewEmailCaptchaProvider 创建邮箱验证码提供者
func NewEmailCaptchaProvider(config *EmailCaptchaConfig, redisClient *redis.Client, emailProvider EmailProvider) *EmailCaptchaProvider {
	return &EmailCaptchaProvider{
		config:        config,
		redis:         redisClient,
		emailProvider: emailProvider,
	}
}

func (p *EmailCaptchaProvider) GetType() CaptchaType {
	return CaptchaTypeEmail
}

func (p *EmailCaptchaProvider) Generate(ctx context.Context, params interface{}) (*CaptchaData, error) {
	email, ok := params.(string)
	if !ok {
		return nil, errors.New("邮箱参数类型错误")
	}

	// 验证邮箱格式
	if !isValidEmail(email) {
		return nil, errors.New("邮箱格式错误")
	}

	// 检查发送频率限制（60秒内只能发送一次）
	rateLimitKey := formatRedisKey(CaptchaTypeEmail, fmt.Sprintf("ratelimit:%s", email))
	exists, err := p.redis.Exists(ctx, rateLimitKey).Result()
	if err != nil {
		return nil, fmt.Errorf("检查频率限制失败: %w", err)
	}
	if exists > 0 {
		return nil, errors.New("发送过于频繁，请稍后再试")
	}

	// 生成验证码ID
	captchaID := uuid.New().String()

	// 生成随机数字验证码
	code := generateRandomDigits(p.config.Length)

	// 发送邮件
	if p.emailProvider != nil {
		if err := p.emailProvider.SendEmail(email, code, p.config.Template); err != nil {
			return nil, fmt.Errorf("邮件发送失败: %w", err)
		}
	}

	// 存储到 Redis
	key := formatRedisKey(CaptchaTypeEmail, captchaID)
	expire := time.Duration(p.config.Expire) * time.Second

	data := map[string]interface{}{
		"code":  code,
		"email": email,
	}
	if err := p.redis.HSet(ctx, key, data).Err(); err != nil {
		return nil, fmt.Errorf("存储验证码失败: %w", err)
	}
	if err := p.redis.Expire(ctx, key, expire).Err(); err != nil {
		return nil, fmt.Errorf("设置过期时间失败: %w", err)
	}

	// 设置发送频率限制
	p.redis.Set(ctx, rateLimitKey, "1", 60*time.Second)

	return &CaptchaData{
		ID:       captchaID,
		Type:     CaptchaTypeEmail,
		Data:     map[string]interface{}{"email": maskEmail(email)},
		ExpireAt: time.Now().Add(expire),
	}, nil
}

func (p *EmailCaptchaProvider) Verify(ctx context.Context, params interface{}) error {
	// 解析参数为 map
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return errors.New("验证参数类型错误")
	}

	captchaID, ok := paramsMap["captchaID"].(string)
	if !ok {
		return errors.New("验证码ID类型错误")
	}
	code, ok := paramsMap["code"].(string)
	if !ok {
		return errors.New("验证码类型错误")
	}

	key := formatRedisKey(CaptchaTypeEmail, captchaID)

	// 从 Redis 获取验证码信息
	data, err := p.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("获取验证码失败: %w", err)
	}
	if len(data) == 0 {
		return errors.New("验证码已过期或不存在")
	}

	storedCode := data["code"]

	// 验证后删除（一次性使用）
	p.redis.Del(ctx, key)

	// 比较验证码
	if storedCode != code {
		return errors.New("验证码错误")
	}

	return nil
}

func (p *EmailCaptchaProvider) IsEnabled() bool {
	return p.config.Enabled
}
