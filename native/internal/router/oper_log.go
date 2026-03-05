package router

import (
	"github.com/force-c/nai-tizi/internal/constants"
	"github.com/force-c/nai-tizi/internal/controller"
	"github.com/force-c/nai-tizi/internal/middleware"
	"github.com/gin-gonic/gin"
)

// registerOperLogRoutes 注册操作日志路由
func registerOperLogRoutes(r *gin.Engine, ctx *RouterContext) {
	operLogController := controller.NewOperLogController(ctx.Container)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		operLog := v1.Group("/operLog")
		operLog.Use(ctx.AuthMiddleware) // 添加认证中间件
		{
			// 创建操作日志 - 需要 oper_log.create 权限
			operLog.POST("", middleware.Permission(ctx.CasbinService, constants.ResourceOperLogCreate), operLogController.CreateOperLog)

			// 分页查询操作日志列表 - 需要 oper_log.read 权限
			operLog.POST("/page", middleware.Permission(ctx.CasbinService, constants.ResourceOperLogRead), operLogController.PageOperLog)

			// 批量删除操作日志 - 需要 oper_log.delete 权限
			operLog.DELETE("/batch", middleware.Permission(ctx.CasbinService, constants.ResourceOperLogDelete), operLogController.BatchDeleteOperLog)

			// 清理操作日志 - 需要 oper_log.delete 权限
			operLog.DELETE("/clean", middleware.Permission(ctx.CasbinService, constants.ResourceOperLogDelete), operLogController.CleanOperLog)

			// 更新、查询和删除操作日志 - 需要 oper_log.update/read/delete 权限（带参数的路由放在最后）
			operLog.PUT("/:id", middleware.Permission(ctx.CasbinService, constants.ResourceOperLogUpdate), operLogController.UpdateOperLog)
			operLog.GET("/:id", middleware.Permission(ctx.CasbinService, constants.ResourceOperLogRead), operLogController.GetOperLogById)
			operLog.DELETE("/:id", middleware.Permission(ctx.CasbinService, constants.ResourceOperLogDelete), operLogController.DeleteOperLog)
		}
	}
}
