package middleware

import (
	"github.com/gcc798/quick.admin/internal/domain/response"
	"github.com/gcc798/quick.admin/internal/service"
	"github.com/gin-gonic/gin"
)

// Permission 权限检查中间件
// 使用方式:
// 1. 单个权限检查: Permission(casbinService, "user.create")
// 2. 通配符权限: Permission(casbinService, "user.*") - 用户模块所有权限
// 3. 只读权限: Permission(casbinService, "*.read") - 所有模块的读权限
//
// 资源格式: "resource.action"，例如 "org.read", "org.create"
// action 从资源字符串中自动解析: *.read = read 操作, 其他 = write 操作
//
// 注意: 此中间件必须在 Auth 中间件之后使用，因为需要从 context 中获取 userId
func Permission(casbinService service.CasbinServiceV2, resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 context 获取用户信息（由 Auth 中间件设置）
		userIdVal, exists := c.Get("userId")
		if !exists {
			response.Forbidden(c, "用户信息不存在")
			c.Abort()
			return
		}
		userId, ok := userIdVal.(int64)
		if !ok {
			response.Forbidden(c, "用户ID格式错误")
			c.Abort()
			return
		}
		// userId == 1 为超级管理员，跳过权限验证
		if userId == 1 {
			c.Next()
			return
		}
		// 从资源字符串中解析 action
		// 格式: "resource.action"，例如 "org.read", "org.create"
		// *.read = read 操作, 其他 = write 操作
		action := "write"
		if len(resource) > 5 && resource[len(resource)-5:] == ".read" {
			action = "read"
		}

		// 检查权限
		allowed, err := casbinService.CheckPermission(c.Request.Context(), userId, resource, action)
		if err != nil {
			response.InternalServerError(c, "权限检查失败: "+err.Error())
			c.Abort()
			return
		}

		if !allowed {
			response.Forbidden(c, "无权限访问")
			c.Abort()
			return
		}

		c.Next()
	}
}

// PermissionAny 任意权限检查中间件（满足其中一个权限即可）
// 使用方式: PermissionAny(casbinService, []string{"user.read", "user.create"})
func PermissionAny(casbinService service.CasbinServiceV2, resources []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIdVal, exists := c.Get("userId")
		if !exists {
			response.Forbidden(c, "用户信息不存在")
			c.Abort()
			return
		}
		userId, ok := userIdVal.(int64)
		if !ok {
			response.Forbidden(c, "用户ID格式错误")
			c.Abort()
			return
		}
		if userId == 1 {
			c.Next()
			return
		}

		// 检查是否满足任意一个权限
		for _, resource := range resources {
			// 从资源字符串中解析 action
			action := "write"
			if len(resource) > 5 && resource[len(resource)-5:] == ".read" {
				action = "read"
			}

			allowed, err := casbinService.CheckPermission(c.Request.Context(), userId, resource, action)
			if err != nil {
				continue
			}
			if allowed {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "无权限访问")
		c.Abort()
	}
}

// PermissionAll 所有权限检查中间件（必须满足所有权限）
// 使用方式: PermissionAll(casbinService, []string{"user.read", "user.update"})
func PermissionAll(casbinService service.CasbinServiceV2, resources []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIdVal, exists := c.Get("userId")
		if !exists {
			response.Forbidden(c, "用户信息不存在")
			c.Abort()
			return
		}
		userId, ok := userIdVal.(int64)
		if !ok {
			response.Forbidden(c, "用户ID格式错误")
			c.Abort()
			return
		}
		if userId == 1 {
			c.Next()
			return
		}

		// 检查是否满足所有权限
		for _, resource := range resources {
			// 从资源字符串中解析 action
			action := "write"
			if len(resource) > 5 && resource[len(resource)-5:] == ".read" {
				action = "read"
			}

			allowed, err := casbinService.CheckPermission(c.Request.Context(), userId, resource, action)
			if err != nil || !allowed {
				response.Forbidden(c, "无权限访问")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
