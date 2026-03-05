package service

import (
	"context"

	"github.com/force-c/nai-tizi/internal/infrastructure/captcha"
)

// CaptchaService 验证码服务接口
type CaptchaService interface {
	// Generate 生成验证码
	Generate(ctx context.Context, captchaType captcha.CaptchaType, params interface{}) (*captcha.CaptchaData, error)

	// Verify 验证验证码
	Verify(ctx context.Context, captchaType captcha.CaptchaType, params interface{}) error

	// GetEnabledTypes 获取已启用的验证码类型
	GetEnabledTypes() []captcha.CaptchaType
}

// captchaService 验证码服务实现
type captchaService struct {
	manager *captcha.CaptchaManager
}

// NewCaptchaService 创建验证码服务
func NewCaptchaService(manager *captcha.CaptchaManager) CaptchaService {
	return &captchaService{
		manager: manager,
	}
}

func (s *captchaService) Generate(ctx context.Context, captchaType captcha.CaptchaType, params interface{}) (*captcha.CaptchaData, error) {
	return s.manager.Generate(ctx, captchaType, params)
}

func (s *captchaService) Verify(ctx context.Context, captchaType captcha.CaptchaType, params interface{}) error {
	return s.manager.Verify(ctx, captchaType, params)
}

func (s *captchaService) GetEnabledTypes() []captcha.CaptchaType {
	return s.manager.GetEnabledTypes()
}
