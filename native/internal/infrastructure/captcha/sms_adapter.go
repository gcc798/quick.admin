package captcha

import (
	"github.com/force-c/nai-tizi/internal/infrastructure/thirdparty/sms"
)

// SMSManagerAdapter 适配器，将 sms.Manager 适配为 SMSProvider 接口
type SMSManagerAdapter struct {
	manager *sms.Manager
}

// NewSMSManagerAdapter 创建 SMS 管理器适配器
func NewSMSManagerAdapter(manager *sms.Manager) *SMSManagerAdapter {
	return &SMSManagerAdapter{manager: manager}
}

// SendSMS 实现 SMSProvider 接口
func (a *SMSManagerAdapter) SendSMS(phone string, code string, template string) error {
	// 使用 sms.Manager 的 Send 方法发送自定义短信
	params := map[string]string{
		"code": code,
	}
	return a.manager.Send(phone, template, params)
}
