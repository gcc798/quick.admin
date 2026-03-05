package router

import (
	"github.com/force-c/nai-tizi/internal/constants"
	"github.com/force-c/nai-tizi/internal/controller"
	"github.com/force-c/nai-tizi/internal/middleware"
	"github.com/gin-gonic/gin"
)

// registerRoleRoutes 注册角色管理路由
func registerRoleRoutes(r *gin.Engine, ctx *RouterContext) {
	// 初始化 controller
	roleController := controller.NewRoleController(ctx.Container)

	// 角色管理路由组（需要认证和权限）
	roles := r.Group("/api/v1/role")
	roles.Use(ctx.AuthMiddleware)
	{
		// 角色创建 - 需要 role.create 权限
		roles.POST("", middleware.Permission(ctx.CasbinService, constants.ResourceRoleCreate), roleController.CreateRole)

		// 角色查询 - 需要 role.read 权限
		roles.POST("/page", middleware.Permission(ctx.CasbinService, constants.ResourceRoleRead), roleController.PageRole)

		// 用户角色管理 - 需要 role.assign 权限
		roles.POST("/assign", middleware.Permission(ctx.CasbinService, constants.ResourceRoleAssign), roleController.AssignRoleToUser)
		roles.DELETE("/remove", middleware.Permission(ctx.CasbinService, constants.ResourceRoleAssign), roleController.RemoveRoleFromUser)
		roles.GET("/user", middleware.Permission(ctx.CasbinService, constants.ResourceRoleRead), roleController.GetUserRoles)

		// 权限管理 - 需要 role.permission 权限（高级权限）
		roles.POST("/permission", middleware.Permission(ctx.CasbinService, constants.ResourceRolePermission), roleController.AddRolePermission)
		roles.DELETE("/permission", middleware.Permission(ctx.CasbinService, constants.ResourceRolePermission), roleController.DeleteRolePermission)
		roles.GET("/permissions", middleware.Permission(ctx.CasbinService, constants.ResourceRolePermission), roleController.GetRolePermissions)

		// 角色更新、查询和删除 - 需要 role.update/read/delete 权限（带参数的路由放在最后）
		roles.PUT("/:roleId", middleware.Permission(ctx.CasbinService, constants.ResourceRoleUpdate), roleController.UpdateRole)
		roles.GET("/:roleId", middleware.Permission(ctx.CasbinService, constants.ResourceRoleRead), roleController.GetRole)
		roles.DELETE("/:roleId", middleware.Permission(ctx.CasbinService, constants.ResourceRoleDelete), roleController.DeleteRole)
	}
}
