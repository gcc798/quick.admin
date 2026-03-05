package router

import (
	"github.com/force-c/nai-tizi/internal/controller"
	"github.com/gin-gonic/gin"
)

// registerAuthRoutes 注册认证相关路由
func registerAuthRoutes(r *gin.Engine, ctx *RouterContext) {
	// 初始化controller
	authController := controller.NewAuthController(ctx.Container)

	// 公开路由（无需认证）
	r.POST("/login", authController.Login)               // 统一登录接口
	r.POST("/logout", authController.Logout)             // 登出
	r.POST("/auth/refresh", authController.RefreshToken) // 刷新Token

	// 需要认证的路由
	r.GET("/me", ctx.AuthMiddleware, authController.Me)
}
