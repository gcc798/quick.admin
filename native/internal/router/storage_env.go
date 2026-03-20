package router

import (
	"github.com/force-c/nai-tizi/internal/constants"
	"github.com/force-c/nai-tizi/internal/controller"
	"github.com/force-c/nai-tizi/internal/middleware"
	"github.com/gin-gonic/gin"
)

// registerStorageEnvRoutes 注册存储环境管理路由
func registerStorageEnvRoutes(r *gin.Engine, ctx *RouterContext) {
	// 初始化 controller
	storageEnvController := controller.NewStorageEnvController(ctx.Container)

	// 存储环境管理路由组（需要认证和权限）
	storageEnvs := r.Group("/api/v1/storage-env")
	storageEnvs.Use(ctx.AuthMiddleware)
	{
		// 存储环境创建 - 需要 storage_env.create 权限
		storageEnvs.POST("", middleware.Permission(ctx.CasbinService, constants.ResourceStorageEnvCreate), storageEnvController.CreateStorageEnv)

		// 存储环境查询 - 需要 storage_env.read 权限
		storageEnvs.POST("/page", middleware.Permission(ctx.CasbinService, constants.ResourceStorageEnvRead), storageEnvController.PageStorageEnv)
		storageEnvs.GET("/default", middleware.Permission(ctx.CasbinService, constants.ResourceStorageEnvRead), storageEnvController.GetDefaultStorageEnv)

		// 设置默认环境 - 需要 storage_env.manage 权限（高级权限）
		storageEnvs.POST("/default", middleware.Permission(ctx.CasbinService, constants.ResourceStorageEnvManage), storageEnvController.SetDefaultStorageEnv)

		// 存储环境更新、查询和删除 - 需要 storage_env.update/read/delete 权限（带参数的路由放在最后）
		storageEnvs.PUT("/:id", middleware.Permission(ctx.CasbinService, constants.ResourceStorageEnvUpdate), storageEnvController.UpdateStorageEnv)
		storageEnvs.GET("/:id", middleware.Permission(ctx.CasbinService, constants.ResourceStorageEnvRead), storageEnvController.GetStorageEnv)
		storageEnvs.POST("/:id/test", middleware.Permission(ctx.CasbinService, constants.ResourceStorageEnvRead), storageEnvController.TestStorageEnvConnection)
		storageEnvs.DELETE("/:id", middleware.Permission(ctx.CasbinService, constants.ResourceStorageEnvDelete), storageEnvController.DeleteStorageEnv)
	}
}
