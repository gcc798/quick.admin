package router

import (
	"github.com/force-c/nai-tizi/internal/constants"
	"github.com/force-c/nai-tizi/internal/controller"
	"github.com/force-c/nai-tizi/internal/middleware"
	"github.com/gin-gonic/gin"
)

// registerOrgRoutes 注册组织管理路由
func registerOrgRoutes(r *gin.Engine, ctx *RouterContext) {
	// 初始化 controller
	orgController := controller.NewOrgController(ctx.Container)

	// 组织管理路由组（需要认证和权限）
	orgs := r.Group("/api/v1/org")
	orgs.Use(ctx.AuthMiddleware)
	{
		// 组织创建
		orgs.POST("",
			middleware.Permission(ctx.CasbinService, constants.ResourceOrgCreate),
			orgController.Create) // 创建组织

		// 组织查询
		orgs.POST("/page",
			middleware.Permission(ctx.CasbinService, constants.ResourceOrgRead),
			orgController.PageOrg) // 分页查询组织列表
		orgs.GET("/tree",
			middleware.Permission(ctx.CasbinService, constants.ResourceOrgRead),
			orgController.GetTree) // 获取组织树

		// 批量删除组织
		orgs.DELETE("/batch",
			middleware.Permission(ctx.CasbinService, constants.ResourceOrgDelete),
			orgController.BatchDelete) // 批量删除组织

		// 组织更新（带参数的路由放在后面）
		orgs.PUT("/:id",
			middleware.Permission(ctx.CasbinService, constants.ResourceOrgUpdate),
			orgController.Update) // 更新组织

		// 组织查询（带参数的路由放在最后）
		orgs.GET("/:id",
			middleware.Permission(ctx.CasbinService, constants.ResourceOrgRead),
			orgController.GetById) // 根据ID查询组织

		// 组织删除（带参数的路由放在最后）
		orgs.DELETE("/:id",
			middleware.Permission(ctx.CasbinService, constants.ResourceOrgDelete),
			orgController.Delete) // 删除单个组织
	}
}
