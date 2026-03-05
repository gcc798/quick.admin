package router

import (
	"github.com/force-c/nai-tizi/internal/constants"
	"github.com/force-c/nai-tizi/internal/controller"
	"github.com/force-c/nai-tizi/internal/middleware"
	"github.com/gin-gonic/gin"
)

// registerLoginLogRoutes 注册登录日志路由
func registerLoginLogRoutes(r *gin.Engine, ctx *RouterContext) {
	loginLogController := controller.NewLoginLogController(ctx.Container)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		loginLog := v1.Group("/loginLog")
		loginLog.Use(ctx.AuthMiddleware) // 添加认证中间件
		{
			// 创建登录日志 - 需要 login_log.create 权限
			loginLog.POST("", middleware.Permission(ctx.CasbinService, constants.ResourceLoginLogCreate), loginLogController.CreateLoginLog)

			// 分页查询登录日志列表 - 需要 login_log.read 权限
			loginLog.POST("/page", middleware.Permission(ctx.CasbinService, constants.ResourceLoginLogRead), loginLogController.PageLoginLog)

			// 批量删除登录日志 - 需要 login_log.delete 权限
			loginLog.DELETE("/batch", middleware.Permission(ctx.CasbinService, constants.ResourceLoginLogDelete), loginLogController.BatchDeleteLoginLog)

			// 清理登录日志 - 需要 login_log.delete 权限
			loginLog.DELETE("/clean", middleware.Permission(ctx.CasbinService, constants.ResourceLoginLogDelete), loginLogController.CleanLoginLog)

			// 更新、查询和删除登录日志 - 需要 login_log.update/read/delete 权限（带参数的路由放在最后）
			loginLog.PUT("/:id", middleware.Permission(ctx.CasbinService, constants.ResourceLoginLogUpdate), loginLogController.UpdateLoginLog)
			loginLog.GET("/:id", middleware.Permission(ctx.CasbinService, constants.ResourceLoginLogRead), loginLogController.GetLoginLogById)
			loginLog.DELETE("/:id", middleware.Permission(ctx.CasbinService, constants.ResourceLoginLogDelete), loginLogController.DeleteLoginLog)
		}
	}
}
