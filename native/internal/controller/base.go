package controller

import (
	"fmt"
	"strings"

	"github.com/force-c/nai-tizi/internal/container"
	"github.com/force-c/nai-tizi/internal/domain/model"
	"github.com/gin-gonic/gin"
)

type BaseController struct {
	ctr container.Container
}

func NewBaseController(c container.Container) *BaseController {
	return &BaseController{ctr: c}
}

// GetUserId 获取当前用户ID
func (b *BaseController) GetUserId(c *gin.Context) (int64, error) {
	if userId, exists := c.Get("userId"); exists {
		if id, ok := userId.(int64); ok {
			return id, nil
		}
	}
	return 0, fmt.Errorf("未登录或用户ID不存在")
}

// GetUserName 获取当前用户名
func (b *BaseController) GetUserName(c *gin.Context) (string, error) {
	if userName, exists := c.Get("userName"); exists {
		if name, ok := userName.(string); ok {
			return name, nil
		}
	}
	return "", fmt.Errorf("未登录或用户名不存在")
}

// GetClientId 获取客户端ID
func (b *BaseController) GetClientId(c *gin.Context) (string, error) {
	if clientId, exists := c.Get("clientId"); exists {
		if id, ok := clientId.(string); ok {
			return id, nil
		}
	}
	return "", fmt.Errorf("客户端ID不存在")
}

// GetDeviceType 获取设备类型
func (b *BaseController) GetDeviceType(c *gin.Context) (string, error) {
	if deviceType, exists := c.Get("deviceType"); exists {
		if dt, ok := deviceType.(string); ok {
			return dt, nil
		}
	}
	return "", fmt.Errorf("设备类型不存在")
}

// CurrentUser 从JWT token解析当前用户信息（不查询数据库）
func (b *BaseController) CurrentUser(c *gin.Context) (*model.User, error) {
	// 先尝试从 context 中获取（middleware 可能已经解析并设置）
	if userId, exists := c.Get("userId"); exists {
		if userName, exists := c.Get("userName"); exists {
			return &model.User{
				ID:       userId.(int64),
				UserName: userName.(string),
			}, nil
		}
	}

	// 如果 context 中没有，从 token 中解析
	tokenHeader := b.ctr.GetConfig().Auth.TokenHeader
	token := c.GetHeader(tokenHeader)
	token = strings.TrimPrefix(token, "Bearer ")
	claims, err := b.ctr.GetJWT().ValidateToken(token)
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:       claims.UserId,
		UserName: claims.UserName,
	}, nil
}
