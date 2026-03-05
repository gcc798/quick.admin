package router

import (
	"github.com/force-c/nai-tizi/internal/controller"
	"github.com/force-c/nai-tizi/internal/infrastructure/websocket"
	"github.com/gin-gonic/gin"
)

// registerCommonRoutes 注册公共路由（健康检查等）
func registerCommonRoutes(r *gin.Engine, ctx *RouterContext) {
	c := ctx.Container
	logger := c.GetLogger()

	// 初始化健康检查控制器
	healthController := controller.NewHealthController(c)

	// 健康检查接口（公开接口，无需认证）
	r.GET("/health", healthController.Health)          // 基础健康检查
	r.GET("/health/ready", healthController.Ready)     // Kubernetes Readiness Probe
	r.GET("/health/live", healthController.Live)       // Kubernetes Liveness Probe
	r.GET("/health/startup", healthController.Startup) // Kubernetes Startup Probe

	// WebSocket路由（需要认证）
	wsHub := c.GetWebSocketHub()
	wsHandler := websocket.NewHandler(wsHub, logger)

	r.GET("/ws", ctx.AuthMiddleware, wsHandler.ServeWs)
}
