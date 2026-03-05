package captcha

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
)

// ImageCaptchaConfig 图形验证码配置
type ImageCaptchaConfig struct {
	Enabled bool `yaml:"enabled"`
	Length  int  `yaml:"length"`
	Width   int  `yaml:"width"`
	Height  int  `yaml:"height"`
	Expire  int  `yaml:"expire"` // 秒
}

// ImageCaptchaProvider 图形验证码提供者
type ImageCaptchaProvider struct {
	config *ImageCaptchaConfig
	redis  *redis.Client
	driver base64Captcha.Driver
}

// NewImageCaptchaProvider 创建图形验证码提供者
func NewImageCaptchaProvider(config *ImageCaptchaConfig, redisClient *redis.Client) *ImageCaptchaProvider {
	// 创建数字验证码驱动
	driver := base64Captcha.NewDriverDigit(
		config.Height,
		config.Width,
		config.Length,
		0.7,
		80,
	)

	return &ImageCaptchaProvider{
		config: config,
		redis:  redisClient,
		driver: driver,
	}
}

func (p *ImageCaptchaProvider) GetType() CaptchaType {
	return CaptchaTypeImage
}

func (p *ImageCaptchaProvider) Generate(ctx context.Context, params interface{}) (*CaptchaData, error) {
	// 生成验证码ID
	captchaID := uuid.New().String()

	// 生成验证码
	captcha := base64Captcha.NewCaptcha(p.driver, base64Captcha.DefaultMemStore)
	id, b64s, _, err := captcha.Generate()
	if err != nil {
		return nil, fmt.Errorf("生成验证码失败: %w", err)
	}

	// 获取验证码答案
	code := base64Captcha.DefaultMemStore.Get(id, true)

	// 存储到 Redis
	key := formatRedisKey(CaptchaTypeImage, captchaID)
	expire := time.Duration(p.config.Expire) * time.Second
	if err := p.redis.Set(ctx, key, code, expire).Err(); err != nil {
		return nil, fmt.Errorf("存储验证码失败: %w", err)
	}

	return &CaptchaData{
		ID:       captchaID,
		Type:     CaptchaTypeImage,
		Data:     b64s, // base64 编码的图片
		ExpireAt: time.Now().Add(expire),
	}, nil
}

func (p *ImageCaptchaProvider) Verify(ctx context.Context, params interface{}) error {
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

	key := formatRedisKey(CaptchaTypeImage, captchaID)

	// 从 Redis 获取验证码
	storedCode, err := p.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return errors.New("验证码已过期或不存在")
	}
	if err != nil {
		return fmt.Errorf("获取验证码失败: %w", err)
	}

	// 验证后删除（一次性使用）
	p.redis.Del(ctx, key)

	// 不区分大小写比较
	if !strings.EqualFold(storedCode, code) {
		return errors.New("验证码错误")
	}

	return nil
}

func (p *ImageCaptchaProvider) IsEnabled() bool {
	return p.config.Enabled
}
