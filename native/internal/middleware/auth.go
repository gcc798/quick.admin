package middleware

import (
	"strings"

	"github.com/gcc798/quick.admin/internal/config"
	"github.com/gcc798/quick.admin/internal/domain/response"
	"github.com/gcc798/quick.admin/internal/service"
	"github.com/gin-gonic/gin"
)

// Auth 认证中间件
// 1. 从配置的请求头读取 AccessToken，WebSocket 路径兼容 query Token
// 2. 验证 AccessToken
// 3. 设置用户信息到 context
func Auth(tokenManager service.TokenManager, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从配置的请求头读取 Token，WebSocket 路径兼容小程序 query 握手。
		tokenHeader := cfg.Auth.TokenHeader
		token := c.GetHeader(tokenHeader)
		if token == "" && isWebSocketPath(c.FullPath()) {
			token = c.Query(tokenHeader)
		}
		if token == "" {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}

		// 去除 "Bearer " 前缀
		token = strings.TrimPrefix(token, "Bearer ")

		// 验证 AccessToken
		claims, err := tokenManager.ValidateAccessToken(c.Request.Context(), token)
		if err != nil {
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		// 可选：验证请求头中的 clientId 是否与 Token 中的一致
		headerClientId := c.GetHeader("clientid")
		if headerClientId == "" {
			headerClientId = c.Query("clientid")
		}
		if headerClientId != "" && claims.ClientId != headerClientId {
			response.Unauthorized(c, "客户端ID与Token不匹配")
			c.Abort()
			return
		}

		// 设置用户信息到 context
		c.Set("userId", claims.UserId)
		c.Set("userName", claims.UserName)
		c.Set("clientId", claims.ClientId)
		c.Set("deviceType", claims.DeviceType)
		c.Next()
	}
}

func isWebSocketPath(path string) bool {
	return path == "/ws" || path == "/resource/websocket"
}
