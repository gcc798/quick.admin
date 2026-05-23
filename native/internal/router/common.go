package router

import (
	"github.com/gcc798/quick.admin/internal/controller"
	"github.com/gcc798/quick.admin/internal/messaging/websocket"
	"github.com/gin-gonic/gin"
)

// 注册公共路由（健康检查等）。
func registerCommonRoutes(r *gin.Engine, ctx *RouterContext) {
	c := ctx.Container
	logger := c.GetLogger()

	// 初始化健康检查控制器
	healthController := controller.NewHealthController(c)

	// 健康检查接口（公开接口，无需认证）
	r.GET("/health", healthController.Health)          // 基础健康检查
	r.GET("/health/ready", healthController.Ready)     // 就绪探针
	r.GET("/health/live", healthController.Live)       // 存活探针
	r.GET("/health/startup", healthController.Startup) // 启动探针

	// 实时连接路由（需要认证）
	wsHub := c.GetWebSocketHub()
	wsHandler := websocket.NewHandler(wsHub, logger)
	if ctx.Bootstrap != nil {
		ctx.Bootstrap.ConfigureWebSocketHandler(wsHandler, c.GetConfig().WebSocket, logger)
	}

	r.GET("/resource/websocket", ctx.AuthMiddleware, wsHandler.ServeWs)
}
