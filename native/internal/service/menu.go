package service

import (
	"errors"
	"fmt"

	"github.com/force-c/nai-tizi/internal/domain/model"
	"gorm.io/gorm"
)

type MenuService struct {
	db *gorm.DB
}

func NewMenuService(db *gorm.DB) *MenuService {
	return &MenuService{
		db: db,
	}
}

// MenuTree 菜单树节点
type MenuTree struct {
	model.Menu
	Children []*MenuTree `json:"children,omitempty"`
}

// GetUserMenuTree 获取用户的菜单树（用于前端路由生成）
func (s *MenuService) GetUserMenuTree(userId int64) ([]*MenuTree, error) {
	// 1. 获取用户的角色列表
	roleModel := &model.Role{}
	roles, err := roleModel.FindByUserId(s.db, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	if len(roles) == 0 {
		return []*MenuTree{}, nil
	}

	// 2. 检查是否是超级管理员
	isSuperAdmin := false
	roleIds := make([]int64, len(roles))
	for i, role := range roles {
		roleIds[i] = role.ID
		// 超级管理员的 role_key 是 super_admin
		if role.RoleKey == "super_admin" {
			isSuperAdmin = true
		}
	}

	var menus []model.Menu
	menuModel := &model.Menu{}

	// 3. 如果是超级管理员，获取所有菜单
	if isSuperAdmin {
		menus, err = menuModel.FindAll(s.db)
		if err != nil {
			return nil, fmt.Errorf("failed to get all menus: %w", err)
		}
	} else {
		// 普通用户，根据角色获取菜单
		menus, err = menuModel.FindByRoleIds(s.db, roleIds)
		if err != nil {
			return nil, fmt.Errorf("failed to get menus by roles: %w", err)
		}
	}

	// 4. 过滤停用/隐藏的菜单，但保留按钮类型（前端需要按钮的权限标识）
	var filteredMenus []model.Menu
	for _, menu := range menus {
		// 保留所有状态正常的菜单（包括按钮类型）
		// 按钮类型的菜单虽然不生成路由，但其 perms 字段用于权限控制
		if menu.Status == 0 {
			filteredMenus = append(filteredMenus, menu)
		}
	}

	// 5. 构建菜单树
	return s.buildMenuTree(filteredMenus, 0), nil
}

// GetAllMenuTree 获取所有菜单树（用于菜单管理页面）
func (s *MenuService) GetAllMenuTree() ([]*MenuTree, error) {
	menuModel := &model.Menu{}
	menus, err := menuModel.FindAll(s.db)
	if err != nil {
		return nil, err
	}
	return s.buildMenuTree(menus, 0), nil
}

// GetMenuList 获取菜单列表
func (s *MenuService) GetMenuList() ([]model.Menu, error) {
	menuModel := &model.Menu{}
	return menuModel.FindAll(s.db)
}

// GetMenuById 根据ID获取菜单
func (s *MenuService) GetMenuById(menuId int64) (*model.Menu, error) {
	menu, err := (&model.Menu{}).FindByID(s.db, menuId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("菜单不存在")
		}
		return nil, fmt.Errorf("查询菜单失败: %w", err)
	}
	return menu, nil
}

// CreateMenu 创建菜单
func (s *MenuService) CreateMenu(menu *model.Menu) error {
	// 场景规则：创建时 ID 必须为空
	if menu.ID != 0 {
		return errors.New("创建时不能指定菜单ID")
	}

	// 场景规则：验证父菜单是否存在
	if menu.ParentId != 0 {
		parent, err := (&model.Menu{}).FindByID(s.db, menu.ParentId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("父菜单不存在")
			}
			return fmt.Errorf("查询父菜单失败: %w", err)
		}

		// 使用领域规则检查是否可以创建子菜单
		if !parent.CanHaveChild(menu.MenuType) {
			if parent.IsDirectory() && menu.IsButton() {
				return errors.New("目录下不能直接创建按钮")
			}
			if parent.IsMenu() && !menu.IsButton() {
				return errors.New("菜单下只能创建按钮")
			}
			return errors.New("不允许创建此类型的子菜单")
		}
	}

	// 场景规则：检查同级菜单名称唯一性
	exists, err := (&model.Menu{}).CheckMenuNameExists(s.db, menu.MenuName, menu.ParentId)
	if err != nil {
		return fmt.Errorf("检查菜单名称失败: %w", err)
	}
	if exists {
		return errors.New("同级菜单名称已存在")
	}

	// 调用模型层的创建方法
	if err := menu.Create(s.db, menu); err != nil {
		return fmt.Errorf("创建菜单失败: %w", err)
	}

	return nil
}

// UpdateMenu 更新菜单
func (s *MenuService) UpdateMenu(menu *model.Menu) error {
	// 场景规则：更新时 ID 必须不为空
	if menu.ID == 0 {
		return errors.New("更新时必须指定菜单ID")
	}

	// 场景规则：检查菜单是否存在
	existing, err := (&model.Menu{}).FindByID(s.db, menu.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("菜单不存在")
		}
		return fmt.Errorf("查询菜单失败: %w", err)
	}

	// 场景规则：不能将自己设置为父菜单
	if menu.ParentId == menu.ID {
		return errors.New("不能将自己设置为父菜单")
	}

	// 场景规则：验证父菜单是否存在
	if menu.ParentId != 0 {
		_, err := (&model.Menu{}).FindByID(s.db, menu.ParentId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("父菜单不存在")
			}
			return fmt.Errorf("查询父菜单失败: %w", err)
		}
	}

	// 场景规则：检查同级菜单名称唯一性（排除自己）
	exists, err := (&model.Menu{}).CheckMenuNameExistsExcludingSelf(s.db, menu.ID, menu.MenuName, menu.ParentId)
	if err != nil {
		return fmt.Errorf("检查菜单名称失败: %w", err)
	}
	if exists {
		return errors.New("同级菜单名称已存在")
	}

	// 保留原有的创建信息
	menu.CreateBy = existing.CreateBy
	menu.CreatedTime = existing.CreatedTime

	// 调用模型层的更新方法
	if err := menu.Update(s.db, menu); err != nil {
		return fmt.Errorf("更新菜单失败: %w", err)
	}

	return nil
}

// DeleteMenu 删除菜单
func (s *MenuService) DeleteMenu(menuId int64) error {
	// 场景规则：ID 必须不为空
	if menuId == 0 {
		return errors.New("菜单ID不能为空")
	}

	// 查询菜单
	menu, err := (&model.Menu{}).FindByID(s.db, menuId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("菜单不存在")
		}
		return fmt.Errorf("查询菜单失败: %w", err)
	}

	// 检查是否有子菜单
	hasChildren, err := menu.HasChildren(s.db)
	if err != nil {
		return fmt.Errorf("检查子菜单失败: %w", err)
	}
	if hasChildren {
		return errors.New("存在子菜单，无法删除")
	}

	// 调用模型层的删除方法
	if err := menu.Delete(s.db, menuId); err != nil {
		return fmt.Errorf("删除菜单失败: %w", err)
	}

	return nil
}

// buildMenuTree 构建菜单树
func (s *MenuService) buildMenuTree(menus []model.Menu, parentId int64) []*MenuTree {
	var tree []*MenuTree

	for _, menu := range menus {
		if menu.ParentId == parentId {
			node := &MenuTree{
				Menu:     menu,
				Children: s.buildMenuTree(menus, menu.ID),
			}
			tree = append(tree, node)
		}
	}

	return tree
}
