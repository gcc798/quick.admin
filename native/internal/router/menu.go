package router

import (
	"github.com/force-c/nai-tizi/internal/constants"
	"github.com/force-c/nai-tizi/internal/controller"
	"github.com/force-c/nai-tizi/internal/middleware"
	"github.com/force-c/nai-tizi/internal/service"
	"github.com/gin-gonic/gin"
)

// registerMenuRoutes 注册菜单管理路由
func registerMenuRoutes(r *gin.Engine, ctx *RouterContext) {
	// 创建 MenuService
	menuService := service.NewMenuService(ctx.Container.GetDB())

	// 创建 MenuController
	menuController := controller.NewMenuController(menuService)

	// 菜单管理路由组（需要认证和权限）
	menus := r.Group("/api/v1/menu")
	menus.Use(ctx.AuthMiddleware)
	{
		// 获取当前用户的菜单树（用于前端路由生成）- 所有登录用户都可以访问
		menus.GET("/user/tree", menuController.GetUserMenuTree)

		// 菜单查询 - 需要 menu.read 权限
		menus.GET("/tree", middleware.Permission(ctx.CasbinService, constants.ResourceMenuRead), menuController.GetMenuTree)
		menus.GET("", middleware.Permission(ctx.CasbinService, constants.ResourceMenuRead), menuController.GetMenuList)
		menus.GET("/:id", middleware.Permission(ctx.CasbinService, constants.ResourceMenuRead), menuController.GetMenuById)

		// 菜单创建 - 需要 menu.create 权限
		menus.POST("", middleware.Permission(ctx.CasbinService, constants.ResourceMenuCreate), menuController.CreateMenu)

		// 菜单更新 - 需要 menu.update 权限
		menus.PUT("/:id", middleware.Permission(ctx.CasbinService, constants.ResourceMenuUpdate), menuController.UpdateMenu)

		// 菜单删除 - 需要 menu.delete 权限
		menus.DELETE("/:id", middleware.Permission(ctx.CasbinService, constants.ResourceMenuDelete), menuController.DeleteMenu)
	}
}
