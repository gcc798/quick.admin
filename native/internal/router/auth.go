package router

import (
	"github.com/gcc798/nai-tizi/internal/controller"
	"github.com/gin-gonic/gin"
)

// 注册认证相关路由。
func registerAuthRoutes(r *gin.Engine, ctx *RouterContext) {
	// 初始化controller
	authController := controller.NewAuthController(ctx.Container)

	// 公开路由（无需认证）
	r.POST("/login", authController.Login)               // 统一登录接口
	r.POST("/logout", authController.Logout)             // 登出
	r.POST("/auth/login", authController.Login)          // 统一登录接口
	r.POST("/auth/logout", authController.Logout)        // 登出
	r.POST("/auth/refresh", authController.RefreshToken) // 刷新Token
}
