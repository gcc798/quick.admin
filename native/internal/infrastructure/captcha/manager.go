package captcha

import (
	"context"
	"fmt"
)

// CaptchaManager 验证码管理器
type CaptchaManager struct {
	providers map[CaptchaType]CaptchaProvider
}

// NewCaptchaManager 创建验证码管理器
func NewCaptchaManager() *CaptchaManager {
	return &CaptchaManager{
		providers: make(map[CaptchaType]CaptchaProvider),
	}
}

// RegisterProvider 注册验证码提供者
func (m *CaptchaManager) RegisterProvider(provider CaptchaProvider) {
	m.providers[provider.GetType()] = provider
}

// GetProvider 获取验证码提供者
func (m *CaptchaManager) GetProvider(captchaType CaptchaType) (CaptchaProvider, error) {
	provider, ok := m.providers[captchaType]
	if !ok {
		return nil, fmt.Errorf("验证码类型 %s 未启用", captchaType)
	}
	return provider, nil
}

// Generate 生成验证码
func (m *CaptchaManager) Generate(ctx context.Context, captchaType CaptchaType, params interface{}) (*CaptchaData, error) {
	provider, err := m.GetProvider(captchaType)
	if err != nil {
		return nil, err
	}
	return provider.Generate(ctx, params)
}

// Verify 验证验证码
func (m *CaptchaManager) Verify(ctx context.Context, captchaType CaptchaType, params interface{}) error {
	provider, err := m.GetProvider(captchaType)
	if err != nil {
		return err
	}
	return provider.Verify(ctx, params)
}

// IsEnabled 检查验证码类型是否启用
func (m *CaptchaManager) IsEnabled(captchaType CaptchaType) bool {
	provider, ok := m.providers[captchaType]
	if !ok {
		return false
	}
	return provider.IsEnabled()
}

// GetEnabledTypes 获取已启用的验证码类型
func (m *CaptchaManager) GetEnabledTypes() []CaptchaType {
	types := make([]CaptchaType, 0, len(m.providers))
	for captchaType := range m.providers {
		types = append(types, captchaType)
	}
	return types
}
