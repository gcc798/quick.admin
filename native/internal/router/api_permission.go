package router

import (
	"github.com/gcc798/quick.admin/internal/constants"
	"github.com/gcc798/quick.admin/internal/controller"
	"github.com/gcc798/quick.admin/internal/middleware"
	"github.com/gin-gonic/gin"
)

// registerApiPermissionRoutes 注册 API 权限管理路由。
func registerApiPermissionRoutes(r *gin.Engine, ctx *RouterContext) {
	apiPermissionController := controller.NewApiPermissionController(ctx.Container)

	apiPermissions := r.Group("/api/v1/api-permission")
	apiPermissions.Use(ctx.AuthMiddleware)
	{
		apiPermissions.GET("/tree", middleware.Permission(ctx.CasbinService, constants.ResourceApiPermissionRead), apiPermissionController.Tree)
		apiPermissions.GET("", middleware.Permission(ctx.CasbinService, constants.ResourceApiPermissionRead), apiPermissionController.List)
		apiPermissions.POST("", middleware.Permission(ctx.CasbinService, constants.ResourceApiPermissionCreate), apiPermissionController.Create)
		apiPermissions.PUT("/:id", middleware.Permission(ctx.CasbinService, constants.ResourceApiPermissionUpdate), apiPermissionController.Update)
		apiPermissions.DELETE("/:id", middleware.Permission(ctx.CasbinService, constants.ResourceApiPermissionDelete), apiPermissionController.Delete)
	}

	roles := r.Group("/api/v1/role")
	roles.Use(ctx.AuthMiddleware)
	{
		roles.GET("/:roleId/api-permissions", middleware.Permission(ctx.CasbinService, constants.ResourceApiPermissionAssign), apiPermissionController.GetRolePermissions)
		roles.POST("/:roleId/api-permissions", middleware.Permission(ctx.CasbinService, constants.ResourceApiPermissionAssign), apiPermissionController.AssignRolePermissions)
	}

	users := r.Group("/api/v1/user")
	users.Use(ctx.AuthMiddleware)
	{
		users.GET("/:id/api-permissions", middleware.Permission(ctx.CasbinService, constants.ResourceApiPermissionAssign), apiPermissionController.GetUserPermissions)
		users.POST("/:id/api-permissions", middleware.Permission(ctx.CasbinService, constants.ResourceApiPermissionAssign), apiPermissionController.AssignUserPermissions)
	}
}
