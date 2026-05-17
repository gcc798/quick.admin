package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/force-c/nai-tizi/internal/domain/model"
	"github.com/force-c/nai-tizi/internal/domain/request"
	"github.com/force-c/nai-tizi/internal/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ApiPermissionTree API 权限树节点。
type ApiPermissionTree struct {
	model.ApiPermission
	Children []*ApiPermissionTree `json:"children,omitempty"`
}

// ApiPermissionService 管理 API 权限资源和授权映射。
type ApiPermissionService interface {
	Tree(ctx context.Context) ([]*ApiPermissionTree, error)
	List(ctx context.Context) ([]model.ApiPermission, error)
	Create(ctx context.Context, req *request.ApiPermissionSaveRequest, userId int64) (*model.ApiPermission, error)
	Update(ctx context.Context, id int64, req *request.ApiPermissionSaveRequest, userId int64) error
	Delete(ctx context.Context, id int64) error
	GetRolePermissionIds(ctx context.Context, roleId int64) ([]int64, error)
	AssignRolePermissions(ctx context.Context, roleId int64, permissionIds []int64, userId int64) error
	GetUserPermissionIds(ctx context.Context, userId int64) ([]int64, error)
	AssignUserPermissions(ctx context.Context, targetUserId int64, permissionIds []int64, operatorId int64) error
}

type apiPermissionService struct {
	db       *gorm.DB
	enforcer *casbin.Enforcer
	logger   logger.Logger
}

func NewApiPermissionService(db *gorm.DB, enforcer *casbin.Enforcer, logger logger.Logger) ApiPermissionService {
	return &apiPermissionService{db: db, enforcer: enforcer, logger: logger}
}

func (s *apiPermissionService) Tree(ctx context.Context) ([]*ApiPermissionTree, error) {
	permissions, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	return buildApiPermissionTree(permissions, 0), nil
}

func (s *apiPermissionService) List(ctx context.Context) ([]model.ApiPermission, error) {
	var permissions []model.ApiPermission
	err := s.db.WithContext(ctx).
		Order("sort ASC, created_time ASC").
		Find(&permissions).Error
	if err != nil {
		return nil, fmt.Errorf("查询 API 权限失败: %w", err)
	}
	return permissions, nil
}

func (s *apiPermissionService) Create(ctx context.Context, req *request.ApiPermissionSaveRequest, userId int64) (*model.ApiPermission, error) {
	permission := &model.ApiPermission{
		ParentId: req.ParentId,
		Module:   req.Module,
		Code:     req.Code,
		Name:     req.Name,
		NodeType: req.NodeType,
		Action:   normalizeAction(req.Code, req.Action),
		Method:   strings.ToUpper(req.Method),
		Path:     req.Path,
		Sort:     req.Sort,
		Status:   req.Status,
		Remark:   req.Remark,
		CreateBy: userId,
		UpdateBy: userId,
	}
	if err := s.validatePermission(ctx, permission, 0); err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Create(permission).Error; err != nil {
		return nil, fmt.Errorf("创建 API 权限失败: %w", err)
	}
	return permission, nil
}

func (s *apiPermissionService) Update(ctx context.Context, id int64, req *request.ApiPermissionSaveRequest, userId int64) error {
	var existing model.ApiPermission
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("API 权限不存在")
		}
		return fmt.Errorf("查询 API 权限失败: %w", err)
	}
	roleIds, userIds, err := s.findAffectedSubjects(ctx, id)
	if err != nil {
		return err
	}
	existing.ParentId = req.ParentId
	existing.Module = req.Module
	existing.Code = req.Code
	existing.Name = req.Name
	existing.NodeType = req.NodeType
	existing.Action = normalizeAction(req.Code, req.Action)
	existing.Method = strings.ToUpper(req.Method)
	existing.Path = req.Path
	existing.Sort = req.Sort
	existing.Status = req.Status
	existing.Remark = req.Remark
	existing.UpdateBy = userId
	if err := s.validatePermission(ctx, &existing, id); err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return fmt.Errorf("更新 API 权限失败: %w", err)
	}
	if err := s.syncAffectedPolicies(ctx, roleIds, userIds); err != nil {
		return err
	}
	return nil
}

func (s *apiPermissionService) Delete(ctx context.Context, id int64) error {
	var childCount int64
	if err := s.db.WithContext(ctx).Model(&model.ApiPermission{}).Where("parent_id = ?", id).Count(&childCount).Error; err != nil {
		return fmt.Errorf("检查子权限失败: %w", err)
	}
	if childCount > 0 {
		return fmt.Errorf("存在子权限，无法删除")
	}
	roleIds, userIds, err := s.findAffectedSubjects(ctx, id)
	if err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := gorm.G[model.MRoleApiPermission](tx).Where("permission_id = ?", id).Delete(ctx); err != nil {
			return fmt.Errorf("删除角色 API 权限关联失败: %w", err)
		}
		if _, err := gorm.G[model.MUserApiPermission](tx).Where("permission_id = ?", id).Delete(ctx); err != nil {
			return fmt.Errorf("删除用户 API 权限关联失败: %w", err)
		}
		if _, err := gorm.G[model.ApiPermission](tx).Where("id = ?", id).Delete(ctx); err != nil {
			return fmt.Errorf("删除 API 权限失败: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := s.syncAffectedPolicies(ctx, roleIds, userIds); err != nil {
		return err
	}
	return nil
}

func (s *apiPermissionService) GetRolePermissionIds(ctx context.Context, roleId int64) ([]int64, error) {
	var rows []model.MRoleApiPermission
	if err := s.db.WithContext(ctx).Where("role_id = ?", roleId).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("查询角色 API 权限失败: %w", err)
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.PermissionId)
	}
	return ids, nil
}

func (s *apiPermissionService) AssignRolePermissions(ctx context.Context, roleId int64, permissionIds []int64, userId int64) error {
	var role model.Role
	if err := s.db.WithContext(ctx).Where("id = ?", roleId).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("角色不存在")
		}
		return fmt.Errorf("查询角色失败: %w", err)
	}
	permissions, normalizedIds, err := s.resolveAssignablePermissions(ctx, permissionIds)
	if err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := gorm.G[model.MRoleApiPermission](tx).Where("role_id = ?", roleId).Delete(ctx); err != nil {
			return fmt.Errorf("删除旧角色 API 权限失败: %w", err)
		}
		rows := make([]model.MRoleApiPermission, 0, len(normalizedIds))
		for _, permissionId := range normalizedIds {
			rows = append(rows, model.MRoleApiPermission{RoleId: roleId, PermissionId: permissionId, CreateBy: userId, UpdateBy: userId})
		}
		if len(rows) > 0 {
			if err := tx.Create(&rows).Error; err != nil {
				return fmt.Errorf("保存角色 API 权限失败: %w", err)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	if err := s.replaceSubjectPolicies(fmt.Sprintf("role::%s", role.RoleKey), permissions); err != nil {
		s.logger.Error("同步角色 API 权限到 Casbin 失败", zap.Int64("roleId", roleId), zap.Error(err))
		return fmt.Errorf("同步 Casbin 失败: %w", err)
	}
	return nil
}

func (s *apiPermissionService) GetUserPermissionIds(ctx context.Context, userId int64) ([]int64, error) {
	var rows []model.MUserApiPermission
	if err := s.db.WithContext(ctx).Where("user_id = ?", userId).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("查询用户 API 权限失败: %w", err)
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.PermissionId)
	}
	return ids, nil
}

func (s *apiPermissionService) AssignUserPermissions(ctx context.Context, targetUserId int64, permissionIds []int64, operatorId int64) error {
	if err := s.db.WithContext(ctx).Where("id = ?", targetUserId).First(&model.User{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("用户不存在")
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}
	permissions, normalizedIds, err := s.resolveAssignablePermissions(ctx, permissionIds)
	if err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := gorm.G[model.MUserApiPermission](tx).Where("user_id = ?", targetUserId).Delete(ctx); err != nil {
			return fmt.Errorf("删除旧用户 API 权限失败: %w", err)
		}
		rows := make([]model.MUserApiPermission, 0, len(normalizedIds))
		for _, permissionId := range normalizedIds {
			rows = append(rows, model.MUserApiPermission{UserId: targetUserId, PermissionId: permissionId, CreateBy: operatorId, UpdateBy: operatorId})
		}
		if len(rows) > 0 {
			if err := tx.Create(&rows).Error; err != nil {
				return fmt.Errorf("保存用户 API 权限失败: %w", err)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	if err := s.replaceSubjectPolicies(fmt.Sprintf("user::%d", targetUserId), permissions); err != nil {
		s.logger.Error("同步用户 API 权限到 Casbin 失败", zap.Int64("userId", targetUserId), zap.Error(err))
		return fmt.Errorf("同步 Casbin 失败: %w", err)
	}
	return nil
}

func (s *apiPermissionService) validatePermission(ctx context.Context, permission *model.ApiPermission, selfId int64) error {
	if permission.ID != 0 && permission.ParentId == permission.ID {
		return fmt.Errorf("不能将自己设置为父级权限")
	}
	if err := s.validateParentChain(ctx, permission.ParentId, selfId); err != nil {
		return err
	}
	var count int64
	query := s.db.WithContext(ctx).Model(&model.ApiPermission{}).Where("code = ?", permission.Code)
	if selfId != 0 {
		query = query.Where("id != ?", selfId)
	}
	if err := query.Count(&count).Error; err != nil {
		return fmt.Errorf("检查权限标识失败: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("权限标识已存在")
	}
	return nil
}

func (s *apiPermissionService) validateParentChain(ctx context.Context, parentId int64, selfId int64) error {
	visited := make(map[int64]struct{})
	for parentId != 0 {
		if _, ok := visited[parentId]; ok {
			return fmt.Errorf("父级权限存在循环引用")
		}
		visited[parentId] = struct{}{}
		if selfId != 0 && parentId == selfId {
			return fmt.Errorf("不能将自己的下级设置为父级权限")
		}
		var parent model.ApiPermission
		if err := s.db.WithContext(ctx).Where("id = ?", parentId).First(&parent).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("父级权限不存在")
			}
			return fmt.Errorf("查询父级权限失败: %w", err)
		}
		parentId = parent.ParentId
	}
	return nil
}

func (s *apiPermissionService) resolveAssignablePermissions(ctx context.Context, permissionIds []int64) ([]model.ApiPermission, []int64, error) {
	if len(permissionIds) == 0 {
		return nil, nil, nil
	}
	uniqueIds := make([]int64, 0, len(permissionIds))
	seen := make(map[int64]struct{}, len(permissionIds))
	for _, id := range permissionIds {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueIds = append(uniqueIds, id)
	}
	var permissions []model.ApiPermission
	if err := s.db.WithContext(ctx).Where("id IN ? AND status = 0", uniqueIds).Find(&permissions).Error; err != nil {
		return nil, nil, fmt.Errorf("查询 API 权限失败: %w", err)
	}
	if len(permissions) != len(uniqueIds) {
		return nil, nil, fmt.Errorf("存在无效或停用的 API 权限")
	}
	normalized := normalizeCoveredPermissions(permissions)
	normalizedIds := make([]int64, 0, len(normalized))
	for _, item := range normalized {
		normalizedIds = append(normalizedIds, item.ID)
	}
	sort.Slice(normalizedIds, func(i, j int) bool { return normalizedIds[i] < normalizedIds[j] })
	return normalized, normalizedIds, nil
}

func (s *apiPermissionService) replaceSubjectPolicies(subject string, permissions []model.ApiPermission) error {
	if _, err := s.enforcer.RemoveFilteredPolicy(0, subject); err != nil {
		return err
	}
	for _, permission := range permissions {
		if permission.Code == "" {
			continue
		}
		action := normalizeAction(permission.Code, permission.Action)
		if _, err := s.enforcer.AddPolicy(subject, permission.Code, action); err != nil {
			return err
		}
	}
	return nil
}

func (s *apiPermissionService) findAffectedSubjects(ctx context.Context, permissionId int64) ([]int64, []int64, error) {
	var roleRows []model.MRoleApiPermission
	if err := s.db.WithContext(ctx).Where("permission_id = ?", permissionId).Find(&roleRows).Error; err != nil {
		return nil, nil, fmt.Errorf("查询受影响角色失败: %w", err)
	}
	var userRows []model.MUserApiPermission
	if err := s.db.WithContext(ctx).Where("permission_id = ?", permissionId).Find(&userRows).Error; err != nil {
		return nil, nil, fmt.Errorf("查询受影响用户失败: %w", err)
	}
	roleIds := make([]int64, 0, len(roleRows))
	for _, row := range roleRows {
		roleIds = append(roleIds, row.RoleId)
	}
	userIds := make([]int64, 0, len(userRows))
	for _, row := range userRows {
		userIds = append(userIds, row.UserId)
	}
	return uniqueInt64(roleIds), uniqueInt64(userIds), nil
}

func (s *apiPermissionService) syncAffectedPolicies(ctx context.Context, roleIds []int64, userIds []int64) error {
	for _, roleId := range uniqueInt64(roleIds) {
		if err := s.syncRolePolicy(ctx, roleId); err != nil {
			return err
		}
	}
	for _, userId := range uniqueInt64(userIds) {
		if err := s.syncUserPolicy(ctx, userId); err != nil {
			return err
		}
	}
	return nil
}

func (s *apiPermissionService) syncRolePolicy(ctx context.Context, roleId int64) error {
	var role model.Role
	if err := s.db.WithContext(ctx).Where("id = ?", roleId).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("查询角色失败: %w", err)
	}
	permissions, err := s.listRolePolicyPermissions(ctx, roleId)
	if err != nil {
		return err
	}
	if err := s.replaceSubjectPolicies(fmt.Sprintf("role::%s", role.RoleKey), normalizeCoveredPermissions(permissions)); err != nil {
		return fmt.Errorf("同步角色 Casbin 权限失败: %w", err)
	}
	return nil
}

func (s *apiPermissionService) syncUserPolicy(ctx context.Context, userId int64) error {
	permissions, err := s.listUserPolicyPermissions(ctx, userId)
	if err != nil {
		return err
	}
	if err := s.replaceSubjectPolicies(fmt.Sprintf("user::%d", userId), normalizeCoveredPermissions(permissions)); err != nil {
		return fmt.Errorf("同步用户 Casbin 权限失败: %w", err)
	}
	return nil
}

func (s *apiPermissionService) listRolePolicyPermissions(ctx context.Context, roleId int64) ([]model.ApiPermission, error) {
	var permissions []model.ApiPermission
	err := s.db.WithContext(ctx).
		Model(&model.ApiPermission{}).
		Joins("INNER JOIN m_role_api_permission rp ON s_api_permission.id = rp.permission_id").
		Where("rp.role_id = ? AND s_api_permission.status = 0", roleId).
		Order("s_api_permission.sort ASC, s_api_permission.created_time ASC").
		Find(&permissions).Error
	if err != nil {
		return nil, fmt.Errorf("查询角色 API 权限失败: %w", err)
	}
	return permissions, nil
}

func (s *apiPermissionService) listUserPolicyPermissions(ctx context.Context, userId int64) ([]model.ApiPermission, error) {
	var permissions []model.ApiPermission
	err := s.db.WithContext(ctx).
		Model(&model.ApiPermission{}).
		Joins("INNER JOIN m_user_api_permission up ON s_api_permission.id = up.permission_id").
		Where("up.user_id = ? AND s_api_permission.status = 0", userId).
		Order("s_api_permission.sort ASC, s_api_permission.created_time ASC").
		Find(&permissions).Error
	if err != nil {
		return nil, fmt.Errorf("查询用户 API 权限失败: %w", err)
	}
	return permissions, nil
}

func uniqueInt64(values []int64) []int64 {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func normalizeCoveredPermissions(permissions []model.ApiPermission) []model.ApiPermission {
	sort.Slice(permissions, func(i, j int) bool {
		return len(permissions[i].Code) < len(permissions[j].Code)
	})
	selected := make([]model.ApiPermission, 0, len(permissions))
	wildcards := make([]string, 0)
	for _, permission := range permissions {
		covered := false
		for _, wildcard := range wildcards {
			prefix := strings.TrimSuffix(wildcard, "*")
			if permission.Code != wildcard && strings.HasPrefix(permission.Code, prefix) {
				covered = true
				break
			}
		}
		if covered {
			continue
		}
		selected = append(selected, permission)
		if strings.HasSuffix(permission.Code, ".*") || permission.Code == "*" {
			wildcards = append(wildcards, permission.Code)
		}
	}
	return selected
}

func normalizeAction(code, action string) string {
	if code == "*" || strings.HasSuffix(code, ".*") {
		return "*"
	}
	if action == "" {
		if strings.HasSuffix(code, ".read") {
			return "read"
		}
		return "write"
	}
	return action
}

func buildApiPermissionTree(permissions []model.ApiPermission, parentId int64) []*ApiPermissionTree {
	tree := make([]*ApiPermissionTree, 0)
	for _, permission := range permissions {
		if permission.ParentId != parentId {
			continue
		}
		node := &ApiPermissionTree{
			ApiPermission: permission,
			Children:      buildApiPermissionTree(permissions, permission.ID),
		}
		tree = append(tree, node)
	}
	return tree
}
