package request

// LoginRequest 登录请求
//
//	@Description	统一登录请求参数，根据 grantType 使用不同的字段组合
type LoginRequest struct {
	// 客户端认证（必填）
	ClientKey    string `json:"clientKey" binding:"required" msg:"客户端Key不能为空" example:"web-admin"`                         // 客户端Key
	ClientSecret string `json:"clientSecret" binding:"required" msg:"客户端密钥不能为空" example:"web-secret-2024"`                 // 客户端密钥
	GrantType    string `json:"grantType" binding:"required" msg:"授权类型不能为空" example:"password" enums:"password,email,xcx"` // 授权类型：password-密码登录, email-邮箱验证码, xcx-微信小程序

	// 用户凭证（根据 grantType 选填）
	Username    string `json:"username" example:"admin"`             // 用户名（password 必填）
	Password    string `json:"password" example:"admin123"`          // 密码（password 必填）
	Code        string `json:"code" example:"123456"`                // 验证码（email/xcx 必填）
	Phonenumber string `json:"phonenumber" example:"13800138000"`    // 手机号（xcx 必填）
	Email       string `json:"email" example:"admin@example.com"`    // 邮箱（email 必填）
	WxCode      string `json:"wxCode" example:"wx-code-from-wechat"` // 微信code（xcx 必填）
	Uuid        string `json:"uuid" example:"captcha-uuid-12345"`    // 图形验证码UUID（password 可选）
}

// SendSmsCodeRequest 发送短信验证码请求
type SendSmsCodeRequest struct {
	Phonenumber string `json:"phonenumber" binding:"required,len=11" msg:"手机号必须是11位数字"`
}

// SendEmailCodeRequest 发送邮箱验证码请求
type SendEmailCodeRequest struct {
	Email string `json:"email" binding:"required,email" msg:"请输入有效的邮箱地址"`
}
