package captcha

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// SMSCaptchaConfig 短信验证码配置
type SMSCaptchaConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Length   int    `yaml:"length"`
	Expire   int    `yaml:"expire"` // 秒
	Template string `yaml:"template"`
	Provider string `yaml:"provider"` // aliyun/tencent
}

// SMSProvider 短信服务提供者接口
type SMSProvider interface {
	SendSMS(phone string, code string, template string) error
}

// SMSCaptchaProvider 短信验证码提供者
type SMSCaptchaProvider struct {
	config      *SMSCaptchaConfig
	redis       *redis.Client
	smsProvider SMSProvider
}

// NewSMSCaptchaProvider 创建短信验证码提供者
func NewSMSCaptchaProvider(config *SMSCaptchaConfig, redisClient *redis.Client, smsProvider SMSProvider) *SMSCaptchaProvider {
	return &SMSCaptchaProvider{
		config:      config,
		redis:       redisClient,
		smsProvider: smsProvider,
	}
}

func (p *SMSCaptchaProvider) GetType() CaptchaType {
	return CaptchaTypeSMS
}

func (p *SMSCaptchaProvider) Generate(ctx context.Context, params interface{}) (*CaptchaData, error) {
	phone, ok := params.(string)
	if !ok {
		return nil, errors.New("手机号参数类型错误")
	}

	// 验证手机号格式
	if !isValidPhone(phone) {
		return nil, errors.New("手机号格式错误")
	}

	// 检查发送频率限制（60秒内只能发送一次）
	rateLimitKey := formatRedisKey(CaptchaTypeSMS, fmt.Sprintf("ratelimit:%s", phone))
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

	// 发送短信
	if p.smsProvider != nil {
		if err := p.smsProvider.SendSMS(phone, code, p.config.Template); err != nil {
			return nil, fmt.Errorf("短信发送失败: %w", err)
		}
	}

	// 存储到 Redis
	key := formatRedisKey(CaptchaTypeSMS, captchaID)
	expire := time.Duration(p.config.Expire) * time.Second

	data := map[string]interface{}{
		"code":  code,
		"phone": phone,
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
		Type:     CaptchaTypeSMS,
		Data:     map[string]interface{}{"phone": maskPhone(phone)},
		ExpireAt: time.Now().Add(expire),
	}, nil
}

func (p *SMSCaptchaProvider) Verify(ctx context.Context, params interface{}) error {
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

	key := formatRedisKey(CaptchaTypeSMS, captchaID)

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

func (p *SMSCaptchaProvider) IsEnabled() bool {
	return p.config.Enabled
}
