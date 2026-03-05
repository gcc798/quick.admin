package captcha

import (
	"github.com/force-c/nai-tizi/internal/infrastructure/thirdparty/email"
)

// EmailManagerAdapter 适配器，将 email.Manager 适配为 EmailProvider 接口
type EmailManagerAdapter struct {
	manager *email.Manager
}

// NewEmailManagerAdapter 创建邮件管理器适配器
func NewEmailManagerAdapter(manager *email.Manager) *EmailManagerAdapter {
	return &EmailManagerAdapter{manager: manager}
}

// SendEmail 实现 EmailProvider 接口
func (a *EmailManagerAdapter) SendEmail(emailAddr string, code string, template string) error {
	return a.manager.SendWithTemplate(emailAddr, code, template)
}
