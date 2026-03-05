package service

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/force-c/nai-tizi/internal/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CasbinServiceV2 Casbin 权限管理服务接口
type CasbinServiceV2 interface {
	// CheckPermission 检查用户权限
	CheckPermission(ctx context.Context, userId int64, resource, action string) (bool, error)

	// AddRoleForUser 为用户分配角色
	AddRoleForUser(ctx context.Context, userId int64, roleKey string) error

	// DeleteRoleForUser 删除用户的角色
	DeleteRoleForUser(ctx context.Context, userId int64, roleKey string) error

	// GetRolesForUser 获取用户的所有角色
	GetRolesForUser(ctx context.Context, userId int64) ([]string, error)

	// GetUsersForRole 获取拥有指定角色的所有用户
	GetUsersForRole(ctx context.Context, roleKey string) ([]int64, error)

	// AddPermissionForRole 为角色添加权限
	// resource: 资源路径（支持通配符，例如: "user.*", "*.read", "*"）
	// action: 操作类型（支持通配符，例如: "write", "*"）
	AddPermissionForRole(ctx context.Context, roleKey string, resource, action string) error

	// DeletePermissionForRole 删除角色的权限
	DeletePermissionForRole(ctx context.Context, roleKey string, resource, action string) error

	// GetPermissionsForRole 获取角色的所有权限
	GetPermissionsForRole(ctx context.Context, roleKey string) ([][]string, error)

	// ReloadPolicy 重新加载策略（从数据库）
	ReloadPolicy(ctx context.Context) error
}

type casbinServiceV2 struct {
	enforcer *casbin.Enforcer
	db       *gorm.DB
	logger   logger.Logger
}

// NewCasbinServiceV2 创建 Casbin 服务实例
func NewCasbinServiceV2(enforcer *casbin.Enforcer, db *gorm.DB, logger logger.Logger) CasbinServiceV2 {
	return &casbinServiceV2{
		enforcer: enforcer,
		db:       db,
		logger:   logger,
	}
}

// CheckPermission 检查用户权限
func (s *casbinServiceV2) CheckPermission(ctx context.Context, userId int64, resource, action string) (bool, error) {
	sub := fmt.Sprintf("user::%d", userId)

	ok, err := s.enforcer.Enforce(sub, resource, action)
	if err != nil {
		s.logger.Error("权限检查失败", zap.Error(err))
		return false, fmt.Errorf("权限检查失败: %w", err)
	}

	s.logger.Debug("权限检查",
		zap.Int64("userId", userId),
		zap.String("resource", resource),
		zap.String("action", action),
		zap.Bool("allowed", ok))

	return ok, nil
}

// AddRoleForUser 为用户分配角色
func (s *casbinServiceV2) AddRoleForUser(ctx context.Context, userId int64, roleKey string) error {
	sub := fmt.Sprintf("user::%d", userId)
	role := fmt.Sprintf("role::%s", roleKey)

	_, err := s.enforcer.AddGroupingPolicy(sub, role)
	if err != nil {
		return fmt.Errorf("添加用户角色失败: %w", err)
	}

	s.logger.Info("添加用户角色",
		zap.Int64("userId", userId),
		zap.String("roleKey", roleKey))

	return nil
}

// DeleteRoleForUser 删除用户的角色
func (s *casbinServiceV2) DeleteRoleForUser(ctx context.Context, userId int64, roleKey string) error {
	sub := fmt.Sprintf("user::%d", userId)
	role := fmt.Sprintf("role::%s", roleKey)

	_, err := s.enforcer.RemoveGroupingPolicy(sub, role)
	if err != nil {
		return fmt.Errorf("删除用户角色失败: %w", err)
	}

	return nil
}

// GetRolesForUser 获取用户的所有角色
func (s *casbinServiceV2) GetRolesForUser(ctx context.Context, userId int64) ([]string, error) {
	sub := fmt.Sprintf("user::%d", userId)

	roles, err := s.enforcer.GetRolesForUser(sub)
	if err != nil {
		return nil, fmt.Errorf("获取用户角色失败: %w", err)
	}

	// 去除 "role::" 前缀
	var result []string
	for _, role := range roles {
		if len(role) > 6 && role[:6] == "role::" {
			result = append(result, role[6:])
		}
	}

	return result, nil
}

// GetUsersForRole 获取拥有指定角色的所有用户
func (s *casbinServiceV2) GetUsersForRole(ctx context.Context, roleKey string) ([]int64, error) {
	role := fmt.Sprintf("role::%s", roleKey)

	users, err := s.enforcer.GetUsersForRole(role)
	if err != nil {
		return nil, fmt.Errorf("获取角色用户失败: %w", err)
	}

	// 解析用户ID
	var result []int64
	for _, user := range users {
		if len(user) > 6 && user[:6] == "user::" {
			var userId int64
			if _, err := fmt.Sscanf(user[6:], "%d", &userId); err == nil {
				result = append(result, userId)
			}
		}
	}

	return result, nil
}

// AddPermissionForRole 为角色添加权限
func (s *casbinServiceV2) AddPermissionForRole(ctx context.Context, roleKey string, resource, action string) error {
	sub := fmt.Sprintf("role::%s", roleKey)

	_, err := s.enforcer.AddPolicy(sub, resource, action)
	if err != nil {
		return fmt.Errorf("添加角色权限失败: %w", err)
	}

	s.logger.Info("添加角色权限",
		zap.String("roleKey", roleKey),
		zap.String("resource", resource),
		zap.String("action", action))

	return nil
}

// DeletePermissionForRole 删除角色的权限
func (s *casbinServiceV2) DeletePermissionForRole(ctx context.Context, roleKey string, resource, action string) error {
	sub := fmt.Sprintf("role::%s", roleKey)

	_, err := s.enforcer.RemovePolicy(sub, resource, action)
	if err != nil {
		return fmt.Errorf("删除角色权限失败: %w", err)
	}

	return nil
}

// GetPermissionsForRole 获取角色的所有权限
func (s *casbinServiceV2) GetPermissionsForRole(ctx context.Context, roleKey string) ([][]string, error) {
	sub := fmt.Sprintf("role::%s", roleKey)

	permissions, err := s.enforcer.GetPermissionsForUser(sub)
	if err != nil {
		return nil, fmt.Errorf("获取角色权限失败: %w", err)
	}

	return permissions, nil
}

// ReloadPolicy 重新加载策略（从数据库）
func (s *casbinServiceV2) ReloadPolicy(ctx context.Context) error {
	if err := s.enforcer.LoadPolicy(); err != nil {
		s.logger.Error("重新加载策略失败", zap.Error(err))
		return fmt.Errorf("重新加载策略失败: %w", err)
	}

	s.logger.Info("重新加载策略成功")
	return nil
}
