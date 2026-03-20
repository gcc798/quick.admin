package router

import (
	"github.com/force-c/nai-tizi/internal/constants"
	"github.com/force-c/nai-tizi/internal/controller"
	"github.com/force-c/nai-tizi/internal/middleware"
	"github.com/gin-gonic/gin"
)

// registerUserRoutes 注册用户管理路由
func registerUserRoutes(r *gin.Engine, ctx *RouterContext) {
	// 初始化 controller
	userController := controller.NewUserController(ctx.Container)

	// 用户管理路由组（需要认证和权限）
	users := r.Group("/api/v1/user")
	users.Use(ctx.AuthMiddleware)
	{
		// 用户创建 - 需要 user.create 权限
		users.POST("", middleware.Permission(ctx.CasbinService, constants.ResourceUserCreate), userController.Create)
		users.POST("/import", middleware.Permission(ctx.CasbinService, constants.ResourceUserCreate), userController.BatchImport)

		// 用户查询 - 需要 user.read 权限
		users.POST("/page", middleware.Permission(ctx.CasbinService, constants.ResourceUserRead), userController.PageUser)

		// 批量删除 - 需要 user.delete 权限
		users.DELETE("/batch", middleware.Permission(ctx.CasbinService, constants.ResourceUserDelete), userController.BatchDelete)

		// 用户修改密码 - 不需要特殊权限（用户修改自己的密码）
		users.POST("/password/change", userController.ChangePassword)

		// 用户更新 - 需要 user.update 权限（带参数的路由放在后面）
		users.PUT("/:id", middleware.Permission(ctx.CasbinService, constants.ResourceUserUpdate), userController.Update)
		users.PUT("/:id/password", middleware.Permission(ctx.CasbinService, constants.ResourceUserUpdate), userController.ResetPassword)

		// 用户查询 - 需要 user.read 权限（带参数的路由放在最后）
		users.GET("/:id", middleware.Permission(ctx.CasbinService, constants.ResourceUserRead), userController.GetById)

		// 用户删除 - 需要 user.delete 权限（带参数的路由放在最后）
		users.DELETE("/:id", middleware.Permission(ctx.CasbinService, constants.ResourceUserDelete), userController.Delete)
	}
}
