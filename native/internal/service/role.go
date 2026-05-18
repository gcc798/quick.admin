package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/gcc798/nai-tizi/internal/domain/model"
	"github.com/gcc798/nai-tizi/internal/logger"
	"github.com/gcc798/nai-tizi/internal/utils/pagination"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RoleService 角色管理服务接口
type RoleService interface {
	// Create 创建角色
	Create(ctx context.Context, role *model.Role) error

	// Update 更新角色
	Update(ctx context.Context, role *model.Role) error

	// Delete 删除角色
	Delete(ctx context.Context, roleId int64) error

	// GetById 根据ID查询角色
	GetById(ctx context.Context, roleId int64) (*model.Role, error)

	// GetByRoleKey 根据角色标识查询角色
	GetByRoleKey(ctx context.Context, roleKey string) (*model.Role, error)

	// Page 分页查询角色列表
	Page(ctx context.Context, pageNum, pageSize int, roleName string, status int32) (*pagination.Page[model.Role], error)

	// AssignRoleToUser 为用户分配角色（包含 Casbin 同步）
	AssignRoleToUser(ctx context.Context, userId, roleId int64) error

	// RemoveRoleFromUser 移除用户的角色（包含 Casbin 同步）
	RemoveRoleFromUser(ctx context.Context, userId, roleId int64) error

	// GetUserRoles 获取用户的所有角色
	GetUserRoles(ctx context.Context, userId int64) ([]model.Role, error)

	// GetRoleUsers 获取拥有该角色的用户
	GetRoleUsers(ctx context.Context, roleId int64) ([]model.User, error)

	// AssignUsersToRole 批量为角色添加用户
	AssignUsersToRole(ctx context.Context, roleId int64, userIds []int64, operatorId int64) error

	// RemoveUsersFromRole 批量移除角色下的用户
	RemoveUsersFromRole(ctx context.Context, roleId int64, userIds []int64) error

	// AssignMenusToRole 为角色分配菜单权限
	AssignMenusToRole(ctx context.Context, roleId int64, menuIds []int64) error

	// GetRoleMenus 获取角色的所有菜单
	GetRoleMenus(ctx context.Context, roleId int64) ([]model.Menu, error)

	// AddRolePermission 为角色添加权限
	AddRolePermission(ctx context.Context, roleKey string, resource, action string) error

	// DeleteRolePermission 删除角色权限
	DeleteRolePermission(ctx context.Context, roleKey string, resource, action string) error

	// GetRolePermissions 获取角色的所有权限
	GetRolePermissions(ctx context.Context, roleKey string) ([][]string, error)
}

type roleService struct {
	db            *gorm.DB
	casbinService CasbinServiceV2
	logger        logger.Logger
}

// NewRoleService 创建角色服务实例
func NewRoleService(db *gorm.DB, casbinService CasbinServiceV2, logger logger.Logger) RoleService {
	return &roleService{
		db:            db,
		casbinService: casbinService,
		logger:        logger,
	}
}

// Create 创建角色
func (s *roleService) Create(ctx context.Context, role *model.Role) error {
	// 检查角色标识是否已存在
	count, err := gorm.G[model.Role](s.db).Where("role_key = ?", role.RoleKey).Count(ctx, "id")
	if err != nil {
		s.logger.Error("检查角色标识失败", zap.Error(err))
		return fmt.Errorf("检查角色标识失败: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("角色标识已存在: %s", role.RoleKey)
	}

	// 创建角色
	if err := gorm.G[model.Role](s.db).Create(ctx, role); err != nil {
		s.logger.Error("创建角色失败", zap.Error(err))
		return fmt.Errorf("创建角色失败: %w", err)
	}

	s.logger.Info("创建角色成功", zap.Int64("roleId", role.ID), zap.String("roleKey", role.RoleKey))
	return nil
}

// Update 更新角色
func (s *roleService) Update(ctx context.Context, role *model.Role) error {
	// 检查角色是否存在
	existingRole, err := gorm.G[model.Role](s.db).Where("id = ?", role.ID).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("角色不存在")
		}
		s.logger.Error("查询角色失败", zap.Error(err))
		return fmt.Errorf("查询角色失败: %w", err)
	}

	// 系统内置角色不允许修改角色标识
	if existingRole.IsSystem && existingRole.RoleKey != role.RoleKey {
		return fmt.Errorf("系统内置角色不允许修改角色标识")
	}

	// 更新角色
	updates := map[string]any{
		"role_name":  role.RoleName,
		"sort":       role.Sort,
		"status":     role.Status,
		"data_scope": role.DataScope,
		"remark":     role.Remark,
		"update_by":  role.UpdateBy,
	}

	if err := s.db.Model(&model.Role{}).Where("id = ?", role.ID).Updates(updates).Error; err != nil {
		s.logger.Error("更新角色失败", zap.Error(err))
		return fmt.Errorf("更新角色失败: %w", err)
	}

	s.logger.Info("更新角色成功", zap.Int64("roleId", role.ID))
	return nil
}

// Delete 删除角色
func (s *roleService) Delete(ctx context.Context, roleId int64) error {
	// 检查角色是否存在
	role, err := gorm.G[model.Role](s.db).Where("id = ?", roleId).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("角色不存在")
		}
		s.logger.Error("查询角色失败", zap.Error(err))
		return fmt.Errorf("查询角色失败: %w", err)
	}

	// 系统内置角色不允许删除
	if role.IsSystem {
		return fmt.Errorf("系统内置角色不允许删除")
	}

	// 检查是否有用户使用该角色
	userRoleCount, err := gorm.G[model.MUserRole](s.db).Where("role_id = ?", roleId).Count(ctx, "id")
	if err != nil {
		return fmt.Errorf("检查角色使用情况失败: %w", err)
	}
	if userRoleCount > 0 {
		return fmt.Errorf("该角色正在被使用，无法删除")
	}

	// 开启事务删除角色及相关数据
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除角色菜单关联
		if _, err := gorm.G[model.MRoleMenu](tx).Where("role_id = ?", roleId).Delete(ctx); err != nil {
			return fmt.Errorf("删除角色菜单关联失败: %w", err)
		}

		// 删除角色
		if _, err := gorm.G[model.Role](tx).Where("id = ?", roleId).Delete(ctx); err != nil {
			return fmt.Errorf("删除角色失败: %w", err)
		}

		s.logger.Info("删除角色成功", zap.Int64("roleId", roleId))
		return nil
	})
}

// GetById 根据ID查询角色
func (s *roleService) GetById(ctx context.Context, roleId int64) (*model.Role, error) {
	role, err := gorm.G[model.Role](s.db).Where("id = ?", roleId).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("角色不存在")
		}
		s.logger.Error("查询角色失败", zap.Error(err))
		return nil, fmt.Errorf("查询角色失败: %w", err)
	}

	return &role, nil
}

// GetByRoleKey 根据角色标识查询角色
func (s *roleService) GetByRoleKey(ctx context.Context, roleKey string) (*model.Role, error) {
	role, err := gorm.G[model.Role](s.db).Where("role_key = ?", roleKey).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("角色不存在")
		}
		s.logger.Error("查询角色失败", zap.Error(err))
		return nil, fmt.Errorf("查询角色失败: %w", err)
	}

	return &role, nil
}

// Page 分页查询角色列表
func (s *roleService) Page(ctx context.Context, pageNum, pageSize int, roleName string, status int32) (*pagination.Page[model.Role], error) {
	query := s.db.Model(&model.Role{})

	// 条件查询
	if roleName != "" {
		query = query.Where("role_name LIKE ?", "%"+roleName+"%")
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	// 构建 PageQuery
	pageQuery := &pagination.PageQuery{
		PageNum:  pageNum,
		PageSize: pageSize,
	}

	// 使用 Paginator 进行分页
	page, err := pagination.New[model.Role](query, pageQuery).Find()
	if err != nil {
		s.logger.Error("分页查询角色列表失败", zap.Error(err))
		return nil, fmt.Errorf("分页查询角色列表失败: %w", err)
	}

	return page, nil
}

// AssignRoleToUser 为用户分配角色（包含 Casbin 同步）
func (s *roleService) AssignRoleToUser(ctx context.Context, userId, roleId int64) error {
	// 使用事务确保数据一致性
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 检查用户是否存在
		if _, err := gorm.G[model.User](tx).Where("id = ?", userId).First(ctx); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("用户不存在")
			}
			return fmt.Errorf("查询用户失败: %w", err)
		}

		// 2. 检查角色是否存在
		role, err := gorm.G[model.Role](tx).Where("id = ?", roleId).First(ctx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("角色不存在")
			}
			return fmt.Errorf("查询角色失败: %w", err)
		}

		// 3. 检查是否已分配
		count, err := gorm.G[model.MUserRole](tx).
			Where("user_id = ? AND role_id = ?", userId, roleId).
			Count(ctx, "id")
		if err != nil {
			return fmt.Errorf("检查用户角色关系失败: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("用户已拥有该角色")
		}

		// 4. 创建用户角色关联
		userRole := &model.MUserRole{
			UserId: userId,
			RoleId: roleId,
		}
		if err := gorm.G[model.MUserRole](tx).Create(ctx, userRole); err != nil {
			s.logger.Error("分配用户角色失败", zap.Error(err))
			return fmt.Errorf("分配用户角色失败: %w", err)
		}

		// 5. 同步到 Casbin（在事务外执行，失败不影响数据库操作）
		// 注意：Casbin 操作失败只记录日志，不回滚事务
		if err := s.casbinService.AddRoleForUser(ctx, userId, role.RoleKey); err != nil {
			s.logger.Error("同步 Casbin 失败",
				zap.Int64("userId", userId),
				zap.String("roleKey", role.RoleKey),
				zap.Error(err))
			// 不返回错误，允许继续
		}

		s.logger.Info("为用户分配角色成功",
			zap.Int64("userId", userId),
			zap.Int64("roleId", roleId),
			zap.String("roleKey", role.RoleKey))

		return nil
	})
}

// RemoveRoleFromUser 移除用户的角色（包含 Casbin 同步）
func (s *roleService) RemoveRoleFromUser(ctx context.Context, userId, roleId int64) error {
	// 使用事务确保数据一致性
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 获取角色信息（用于 Casbin 同步）
		role, err := gorm.G[model.Role](tx).Where("id = ?", roleId).First(ctx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("角色不存在")
			}
			return fmt.Errorf("查询角色失败: %w", err)
		}

		// 2. 从数据库移除用户角色关联
		rowsAffected, err := gorm.G[model.MUserRole](tx).Where("user_id = ? AND role_id = ?", userId, roleId).Delete(ctx)

		if err != nil {
			s.logger.Error("移除用户角色失败", zap.Error(err))
			return fmt.Errorf("移除用户角色失败: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("用户角色关系不存在")
		}

		// 3. 从 Casbin 移除（在事务外执行，失败不影响数据库操作）
		if err := s.casbinService.DeleteRoleForUser(ctx, userId, role.RoleKey); err != nil {
			s.logger.Error("从 Casbin 移除角色失败",
				zap.Int64("userId", userId),
				zap.String("roleKey", role.RoleKey),
				zap.Error(err))
			// 不返回错误，允许继续
		}

		s.logger.Info("移除用户角色成功",
			zap.Int64("userId", userId),
			zap.Int64("roleId", roleId),
			zap.String("roleKey", role.RoleKey))

		return nil
	})
}

// GetUserRoles 获取用户的所有角色
func (s *roleService) GetUserRoles(ctx context.Context, userId int64) ([]model.Role, error) {
	var roles []model.Role

	err := s.db.Table("s_role r").
		Joins("INNER JOIN m_user_role ur ON r.id = ur.role_id").
		Where("ur.user_id = ? AND r.status = 0", userId).
		Order("r.sort ASC").
		Find(&roles).Error

	if err != nil {
		s.logger.Error("查询用户角色失败", zap.Error(err))
		return nil, fmt.Errorf("查询用户角色失败: %w", err)
	}

	return roles, nil
}

// GetRoleUsers 获取拥有该角色的用户
func (s *roleService) GetRoleUsers(ctx context.Context, roleId int64) ([]model.User, error) {
	if _, err := gorm.G[model.Role](s.db).Where("id = ?", roleId).First(ctx); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("角色不存在")
		}
		return nil, fmt.Errorf("查询角色失败: %w", err)
	}

	var users []model.User
	err := s.db.Table("s_user u").
		Joins("INNER JOIN m_user_role ur ON u.id = ur.user_id").
		Where("ur.role_id = ?", roleId).
		Order("u.created_time DESC").
		Find(&users).Error
	if err != nil {
		s.logger.Error("查询角色用户失败", zap.Error(err))
		return nil, fmt.Errorf("查询角色用户失败: %w", err)
	}

	return users, nil
}

// AssignUsersToRole 批量为角色添加用户
func (s *roleService) AssignUsersToRole(ctx context.Context, roleId int64, userIds []int64, operatorId int64) error {
	userIds = uniqueInt64(userIds)
	if len(userIds) == 0 {
		return nil
	}

	role, err := gorm.G[model.Role](s.db).Where("id = ?", roleId).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("角色不存在")
		}
		return fmt.Errorf("查询角色失败: %w", err)
	}

	var existingUsers []model.User
	if err := s.db.WithContext(ctx).
		Where("id IN ?", userIds).
		Find(&existingUsers).Error; err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}
	if len(existingUsers) != len(userIds) {
		return fmt.Errorf("部分用户不存在")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var rows []model.MUserRole
		if err := tx.WithContext(ctx).
			Where("role_id = ? AND user_id IN ?", roleId, userIds).
			Find(&rows).Error; err != nil {
			return fmt.Errorf("查询用户角色关系失败: %w", err)
		}

		existing := make(map[int64]struct{}, len(rows))
		for _, row := range rows {
			existing[row.UserId] = struct{}{}
		}

		toCreate := make([]model.MUserRole, 0, len(userIds)-len(existing))
		for _, userId := range userIds {
			if _, ok := existing[userId]; ok {
				continue
			}
			toCreate = append(toCreate, model.MUserRole{
				UserId:   userId,
				RoleId:   roleId,
				CreateBy: operatorId,
				UpdateBy: operatorId,
			})
		}

		if len(toCreate) == 0 {
			return nil
		}

		if err := tx.Create(&toCreate).Error; err != nil {
			return fmt.Errorf("批量分配角色用户失败: %w", err)
		}

		for _, row := range toCreate {
			if err := s.casbinService.AddRoleForUser(ctx, row.UserId, role.RoleKey); err != nil {
				s.logger.Error("同步 Casbin 失败",
					zap.Int64("userId", row.UserId),
					zap.String("roleKey", role.RoleKey),
					zap.Error(err))
			}
		}

		s.logger.Info("批量为角色添加用户成功",
			zap.Int64("roleId", roleId),
			zap.Int("userCount", len(toCreate)))

		return nil
	})
}

// RemoveUsersFromRole 批量移除角色下的用户
func (s *roleService) RemoveUsersFromRole(ctx context.Context, roleId int64, userIds []int64) error {
	userIds = uniqueInt64(userIds)
	if len(userIds) == 0 {
		return nil
	}

	role, err := gorm.G[model.Role](s.db).Where("id = ?", roleId).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("角色不存在")
		}
		return fmt.Errorf("查询角色失败: %w", err)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).
			Where("role_id = ? AND user_id IN ?", roleId, userIds).
			Delete(&model.MUserRole{}).Error; err != nil {
			return fmt.Errorf("批量移除角色用户失败: %w", err)
		}

		for _, userId := range userIds {
			if err := s.casbinService.DeleteRoleForUser(ctx, userId, role.RoleKey); err != nil {
				s.logger.Error("从 Casbin 移除角色失败",
					zap.Int64("userId", userId),
					zap.String("roleKey", role.RoleKey),
					zap.Error(err))
			}
		}

		s.logger.Info("批量移除角色用户成功",
			zap.Int64("roleId", roleId),
			zap.Int("userCount", len(userIds)))

		return nil
	})
}

// AssignMenusToRole 为角色分配菜单权限
func (s *roleService) AssignMenusToRole(ctx context.Context, roleId int64, menuIds []int64) error {
	// 检查角色是否存在
	role, err := gorm.G[model.Role](s.db).Where("id = ?", roleId).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("角色不存在")
		}
		return fmt.Errorf("查询角色失败: %w", err)
	}
	menuIds = uniqueInt64(menuIds)
	oldMenuPermCodes, err := s.menuPermCodesByRole(ctx, roleId)
	if err != nil {
		return err
	}
	newMenuPermCodes, err := s.menuPermCodesByIDs(ctx, menuIds)
	if err != nil {
		return err
	}
	desiredApiPermissions, removedApiPermissions, err := s.diffMenuApiPermissions(ctx, oldMenuPermCodes, newMenuPermCodes)
	if err != nil {
		return err
	}

	// 开启事务
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		// 删除旧的菜单权限
		if err := tx.WithContext(ctx).Where("role_id = ?", roleId).Delete(&model.MRoleMenu{}).Error; err != nil {
			return fmt.Errorf("删除旧菜单权限失败: %w", err)
		}

		// 添加新的菜单权限
		if len(menuIds) > 0 {
			roleMenus := make([]model.MRoleMenu, 0, len(menuIds))
			for _, menuId := range menuIds {
				roleMenus = append(roleMenus, model.MRoleMenu{
					RoleId: roleId,
					MenuId: menuId,
				})
			}
			if err := tx.Create(&roleMenus).Error; err != nil {
				return fmt.Errorf("添加新菜单权限失败: %w", err)
			}
		}

		if err := s.applyMenuApiPermissionChanges(ctx, tx, roleId, desiredApiPermissions, removedApiPermissions); err != nil {
			return err
		}

		s.logger.Info("为角色分配菜单权限成功",
			zap.Int64("roleId", roleId),
			zap.Int("menuCount", len(menuIds)))

		return nil
	}); err != nil {
		return err
	}

	if err := s.casbinService.DeletePermissionsForRole(ctx, role.RoleKey, removedApiPermissions); err != nil {
		s.logger.Error("批量删除角色菜单联动 API 权限 Casbin 策略失败",
			zap.Int64("roleId", roleId),
			zap.String("roleKey", role.RoleKey),
			zap.Int("permissionCount", len(removedApiPermissions)),
			zap.Error(err))
	}
	if err := s.casbinService.AddPermissionsForRole(ctx, role.RoleKey, desiredApiPermissions); err != nil {
		s.logger.Error("批量添加角色菜单联动 API 权限 Casbin 策略失败",
			zap.Int64("roleId", roleId),
			zap.String("roleKey", role.RoleKey),
			zap.Int("permissionCount", len(desiredApiPermissions)),
			zap.Error(err))
	}

	return nil
}

func (s *roleService) menuPermCodesByRole(ctx context.Context, roleId int64) (map[string]struct{}, error) {
	var menus []model.Menu
	if err := s.db.WithContext(ctx).
		Table("s_menu m").
		Select("m.perms").
		Joins("INNER JOIN m_role_menu rm ON m.id = rm.menu_id").
		Where("rm.role_id = ? AND m.status = 0 AND m.perms <> ''", roleId).
		Find(&menus).Error; err != nil {
		return nil, fmt.Errorf("查询角色菜单权限标识失败: %w", err)
	}
	return menuPermCodeSet(menus), nil
}

func (s *roleService) menuPermCodesByIDs(ctx context.Context, menuIds []int64) (map[string]struct{}, error) {
	result := make(map[string]struct{})
	if len(menuIds) == 0 {
		return result, nil
	}

	var menus []model.Menu
	if err := s.db.WithContext(ctx).
		Where("id IN ? AND status = 0 AND perms <> ''", menuIds).
		Find(&menus).Error; err != nil {
		return nil, fmt.Errorf("查询菜单权限标识失败: %w", err)
	}
	return menuPermCodeSet(menus), nil
}

func menuPermCodeSet(menus []model.Menu) map[string]struct{} {
	result := make(map[string]struct{}, len(menus))
	for _, menu := range menus {
		if menu.Perms == "" {
			continue
		}
		result[menu.Perms] = struct{}{}
	}
	return result
}

func (s *roleService) diffMenuApiPermissions(ctx context.Context, oldCodes, newCodes map[string]struct{}) ([]model.ApiPermission, []model.ApiPermission, error) {
	removedCodes := make([]string, 0)
	for code := range oldCodes {
		if _, ok := newCodes[code]; !ok {
			removedCodes = append(removedCodes, code)
		}
	}

	desiredCodes := stringSetKeys(newCodes)
	desiredPermissions, err := s.apiPermissionsByCodes(ctx, desiredCodes)
	if err != nil {
		return nil, nil, err
	}
	removedPermissions, err := s.apiPermissionsByCodes(ctx, removedCodes)
	if err != nil {
		return nil, nil, err
	}
	s.logMissingMenuApiPermissions("保存", desiredCodes, desiredPermissions)
	s.logMissingMenuApiPermissions("移除", removedCodes, removedPermissions)

	return desiredPermissions, removedPermissions, nil
}

func stringSetKeys(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	return result
}

func (s *roleService) apiPermissionsByCodes(ctx context.Context, codes []string) ([]model.ApiPermission, error) {
	if len(codes) == 0 {
		return nil, nil
	}

	var permissions []model.ApiPermission
	if err := s.db.WithContext(ctx).
		Where("code IN ? AND status = 0", codes).
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("查询菜单对应 API 权限失败: %w", err)
	}
	return permissions, nil
}

func (s *roleService) logMissingMenuApiPermissions(operation string, codes []string, permissions []model.ApiPermission) {
	if len(codes) == 0 {
		return
	}
	found := make(map[string]struct{}, len(permissions))
	for _, permission := range permissions {
		found[permission.Code] = struct{}{}
	}
	for _, code := range codes {
		if _, ok := found[code]; ok {
			continue
		}
		s.logger.Warn("菜单权限未找到同 code 的 API 权限，跳过自动同步",
			zap.String("operation", operation),
			zap.String("permissionCode", code))
	}
}

func (s *roleService) applyMenuApiPermissionChanges(ctx context.Context, tx *gorm.DB, roleId int64, addedPermissions, removedPermissions []model.ApiPermission) error {
	if len(removedPermissions) > 0 {
		removedIDs := apiPermissionIDs(removedPermissions)
		if err := tx.WithContext(ctx).
			Where("role_id = ? AND permission_id IN ?", roleId, removedIDs).
			Delete(&model.MRoleApiPermission{}).Error; err != nil {
			return fmt.Errorf("删除菜单联动 API 权限失败: %w", err)
		}
	}

	if len(addedPermissions) == 0 {
		return nil
	}

	addedIDs := apiPermissionIDs(addedPermissions)
	var existingRows []model.MRoleApiPermission
	if err := tx.WithContext(ctx).
		Where("role_id = ? AND permission_id IN ?", roleId, addedIDs).
		Find(&existingRows).Error; err != nil {
		return fmt.Errorf("查询已有角色 API 权限失败: %w", err)
	}

	existing := make(map[int64]struct{}, len(existingRows))
	for _, row := range existingRows {
		existing[row.PermissionId] = struct{}{}
	}

	rows := make([]model.MRoleApiPermission, 0, len(addedPermissions))
	missingIDs := make([]int64, 0, len(addedPermissions))
	for _, permission := range addedPermissions {
		if _, ok := existing[permission.ID]; ok {
			continue
		}
		missingIDs = append(missingIDs, permission.ID)
		rows = append(rows, model.MRoleApiPermission{RoleId: roleId, PermissionId: permission.ID})
	}
	if len(rows) == 0 {
		return nil
	}

	if err := tx.WithContext(ctx).
		Where("role_id = ? AND permission_id IN ?", roleId, missingIDs).
		Delete(&model.MRoleApiPermission{}).Error; err != nil {
		return fmt.Errorf("清理历史角色 API 权限失败: %w", err)
	}
	if err := tx.WithContext(ctx).Create(&rows).Error; err != nil {
		return fmt.Errorf("保存菜单联动 API 权限失败: %w", err)
	}
	return nil
}

func apiPermissionIDs(permissions []model.ApiPermission) []int64 {
	ids := make([]int64, 0, len(permissions))
	for _, permission := range permissions {
		ids = append(ids, permission.ID)
	}
	return ids
}

// GetRoleMenus 获取角色的所有菜单
func (s *roleService) GetRoleMenus(ctx context.Context, roleId int64) ([]model.Menu, error) {
	var menus []model.Menu

	err := s.db.Table("s_menu m").
		Joins("INNER JOIN m_role_menu rm ON m.id = rm.menu_id").
		Where("rm.role_id = ? AND m.status = 0", roleId).
		Group("m.id").
		Order("m.sort ASC").
		Find(&menus).Error

	if err != nil {
		s.logger.Error("查询角色菜单失败", zap.Error(err))
		return nil, fmt.Errorf("查询角色菜单失败: %w", err)
	}

	return menus, nil
}

// AddRolePermission 为角色添加权限
func (s *roleService) AddRolePermission(ctx context.Context, roleKey string, resource, action string) error {
	if err := s.casbinService.AddPermissionForRole(ctx, roleKey, resource, action); err != nil {
		s.logger.Error("为角色添加权限失败",
			zap.String("roleKey", roleKey),
			zap.String("resource", resource),
			zap.String("action", action),
			zap.Error(err))
		return fmt.Errorf("为角色添加权限失败: %w", err)
	}

	s.logger.Info("为角色添加权限成功",
		zap.String("roleKey", roleKey),
		zap.String("resource", resource),
		zap.String("action", action))

	return nil
}

// DeleteRolePermission 删除角色权限
func (s *roleService) DeleteRolePermission(ctx context.Context, roleKey string, resource, action string) error {
	if err := s.casbinService.DeletePermissionForRole(ctx, roleKey, resource, action); err != nil {
		s.logger.Error("删除角色权限失败",
			zap.String("roleKey", roleKey),
			zap.String("resource", resource),
			zap.String("action", action),
			zap.Error(err))
		return fmt.Errorf("删除角色权限失败: %w", err)
	}

	s.logger.Info("删除角色权限成功",
		zap.String("roleKey", roleKey),
		zap.String("resource", resource),
		zap.String("action", action))

	return nil
}

// GetRolePermissions 获取角色的所有权限
func (s *roleService) GetRolePermissions(ctx context.Context, roleKey string) ([][]string, error) {
	permissions, err := s.casbinService.GetPermissionsForRole(ctx, roleKey)
	if err != nil {
		s.logger.Error("获取角色权限失败",
			zap.String("roleKey", roleKey),
			zap.Error(err))
		return nil, fmt.Errorf("获取角色权限失败: %w", err)
	}

	return permissions, nil
}
