package router

import (
	"github.com/force-c/nai-tizi/internal/controller"
	"github.com/force-c/nai-tizi/internal/service"
	"github.com/gin-gonic/gin"
)

// registerCaptchaRoutes 注册验证码相关路由
func registerCaptchaRoutes(r *gin.Engine, ctx *RouterContext) {
	captchaService := service.NewCaptchaService(ctx.Container.GetCaptchaManager())
	captchaController := controller.NewCaptchaController(captchaService)

	// 公开路由（无需认证）
	captcha := r.Group("/captcha")
	{
		captcha.GET("/image", captchaController.GenerateImageCaptcha)    // 生成图形验证码
		captcha.POST("/sms", captchaController.SendSMSCaptcha)           // 发送短信验证码
		captcha.POST("/email", captchaController.SendEmailCaptcha)       // 发送邮箱验证码
		captcha.GET("/enabled-types", captchaController.GetEnabledTypes) // 获取启用的验证码类型
	}
}
