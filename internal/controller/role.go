package controller

import (
	"strconv"

	"github.com/force-c/nai-tizi/internal/container"
	"github.com/force-c/nai-tizi/internal/domain/model"
	"github.com/force-c/nai-tizi/internal/domain/request"
	"github.com/force-c/nai-tizi/internal/domain/response"
	"github.com/force-c/nai-tizi/internal/logger"
	"github.com/force-c/nai-tizi/internal/service"
	_ "github.com/force-c/nai-tizi/internal/utils/pagination"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RoleController 角色控制器接口
type RoleController interface {
	CreateRole(ctx *gin.Context)           // 创建角色
	UpdateRole(ctx *gin.Context)           // 更新角色
	DeleteRole(ctx *gin.Context)           // 删除角色
	GetRole(ctx *gin.Context)              // 获取角色详情
	PageRole(ctx *gin.Context)             // 分页查询角色列表
	AssignRoleToUser(ctx *gin.Context)     // 为用户分配角色
	RemoveRoleFromUser(ctx *gin.Context)   // 移除用户的角色
	GetUserRoles(ctx *gin.Context)         // 获取用户的所有角色
	AddRolePermission(ctx *gin.Context)    // 为角色添加权限
	DeleteRolePermission(ctx *gin.Context) // 删除角色权限
	GetRolePermissions(ctx *gin.Context)   // 获取角色的所有权限
}

type roleController struct {
	roleService service.RoleService
	logger      logger.Logger
}

func NewRoleController(c container.Container) RoleController {
	casbinService := service.NewCasbinServiceV2(c.GetCasbin(), c.GetDB(), c.GetLogger())
	return &roleController{
		roleService: service.NewRoleService(c.GetDB(), casbinService, c.GetLogger()),
		logger:      c.GetLogger(),
	}
}

// CreateRole 创建角色
//
//	@Summary		创建角色
//	@Description	创建新角色
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string						true	"Bearer {token}"
//	@Param			body			body		request.CreateRoleRequest	true	"角色信息"
//	@Success		200				{object}	response.Response{data=response.RoleResponse}
//	@Failure		400				{object}	response.Response	"参数错误"
//	@Router			/api/v1/role [post]
func (c *roleController) CreateRole(ctx *gin.Context) {
	var req request.CreateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误: "+err.Error())
		return
	}

	userId, _ := ctx.Get("userId")

	role := &model.Role{
		RoleKey:   req.RoleKey,
		RoleName:  req.RoleName,
		Sort:      req.Sort,
		Status:    req.Status,
		DataScope: req.DataScope,
		IsSystem:  false,
		Remark:    req.Remark,
	}
	role.CreateBy = userId.(int64)

	if err := c.roleService.Create(ctx.Request.Context(), role); err != nil {
		c.logger.Error("创建角色失败", zap.Error(err))
		response.InternalServerError(ctx, "创建角色失败: "+err.Error())
		return
	}

	response.Success(ctx, role)
}

// UpdateRole 更新角色
//
//	@Summary		更新角色
//	@Description	更新角色信息
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string						true	"Bearer {token}"
//	@Param			body			body		request.UpdateRoleRequest	true	"角色信息"
//	@Success		200				{object}	response.Response
//	@Failure		400				{object}	response.Response	"参数错误"
//	@Router			/api/v1/role [put]
func (c *roleController) UpdateRole(ctx *gin.Context) {
	var req request.UpdateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误: "+err.Error())
		return
	}

	userId, _ := ctx.Get("userId")

	role := &model.Role{
		ID:        req.RoleId,
		RoleName:  req.RoleName,
		Sort:      req.Sort,
		Status:    req.Status,
		DataScope: req.DataScope,
		Remark:    req.Remark,
	}
	role.UpdateBy = userId.(int64)

	if err := c.roleService.Update(ctx.Request.Context(), role); err != nil {
		c.logger.Error("更新角色失败", zap.Error(err))
		response.InternalServerError(ctx, "更新角色失败: "+err.Error())
		return
	}

	response.Success(ctx, nil)
}

// DeleteRole 删除角色
//
//	@Summary		删除角色
//	@Description	删除角色
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer {token}"
//	@Param			roleId			path		int		true	"角色ID"
//	@Success		200				{object}	response.Response
//	@Failure		400				{object}	response.Response	"参数错误"
//	@Router			/api/v1/role/{roleId} [delete]
func (c *roleController) DeleteRole(ctx *gin.Context) {
	roleIdStr := ctx.Param("roleId")
	roleId, err := strconv.ParseInt(roleIdStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "角色ID格式错误")
		return
	}

	if err := c.roleService.Delete(ctx.Request.Context(), roleId); err != nil {
		c.logger.Error("删除角色失败", zap.Error(err))
		response.InternalServerError(ctx, "删除角色失败: "+err.Error())
		return
	}

	response.Success(ctx, nil)
}

// GetRole 获取角色详情
//
//	@Summary		获取角色详情
//	@Description	根据角色ID获取角色详情
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer {token}"
//	@Param			roleId			path		int		true	"角色ID"
//	@Success		200				{object}	response.Response{data=response.RoleResponse}
//	@Failure		400				{object}	response.Response	"参数错误"
//	@Router			/api/v1/role/{roleId} [get]
func (c *roleController) GetRole(ctx *gin.Context) {
	roleIdStr := ctx.Param("roleId")
	roleId, err := strconv.ParseInt(roleIdStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "角色ID格式错误")
		return
	}

	role, err := c.roleService.GetById(ctx.Request.Context(), roleId)
	if err != nil {
		c.logger.Error("获取角色详情失败", zap.Error(err))
		response.InternalServerError(ctx, "获取角色详情失败: "+err.Error())
		return
	}

	response.Success(ctx, role)
}

// PageRole 分页查询角色列表
//
//	@Summary		分页查询角色列表
//	@Description	使用 Paginator 分页查询角色列表
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string					true	"Bearer {token}"
//	@Param			body			body		request.PageRoleRequest	true	"查询参数"
//	@Success		200				{object}	response.Response{data=object}
//	@Router			/api/v1/role/page [post]
func (c *roleController) PageRole(ctx *gin.Context) {
	var req request.PageRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误: "+err.Error())
		return
	}

	page, err := c.roleService.Page(ctx.Request.Context(), req.PageNum, req.PageSize, req.RoleName, req.Status)
	if err != nil {
		c.logger.Error("分页查询角色列表失败", zap.Error(err))
		response.InternalServerError(ctx, "分页查询角色列表失败: "+err.Error())
		return
	}

	response.Success(ctx, page)
}

// AssignRoleToUser 为用户分配角色
//
//	@Summary		为用户分配角色
//	@Description	为用户分配角色
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string							true	"Bearer {token}"
//	@Param			body			body		request.AssignRoleToUserRequest	true	"分配信息"
//	@Success		200				{object}	response.Response
//	@Failure		400				{object}	response.Response	"参数错误"
//	@Router			/api/v1/role/assign [post]
func (c *roleController) AssignRoleToUser(ctx *gin.Context) {
	var req request.AssignRoleToUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误: "+err.Error())
		return
	}

	if err := c.roleService.AssignRoleToUser(ctx.Request.Context(), req.UserId, req.RoleId); err != nil {
		c.logger.Error("为用户分配角色失败", zap.Error(err))
		response.InternalServerError(ctx, "为用户分配角色失败: "+err.Error())
		return
	}

	response.Success(ctx, nil)
}

// RemoveRoleFromUser 移除用户的角色
//
//	@Summary		移除用户的角色
//	@Description	移除指定用户的角色
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer {token}"
//	@Param			userId			query		int		true	"用户ID"
//	@Param			roleId			query		int		true	"角色ID"
//	@Success		200				{object}	response.Response
//	@Failure		400				{object}	response.Response	"参数错误"
//	@Router			/api/v1/role/remove [delete]
func (c *roleController) RemoveRoleFromUser(ctx *gin.Context) {
	userIdStr := ctx.Query("userId")
	roleIdStr := ctx.Query("roleId")

	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "用户ID格式错误")
		return
	}
	roleId, err := strconv.ParseInt(roleIdStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "角色ID格式错误")
		return
	}

	if err := c.roleService.RemoveRoleFromUser(ctx.Request.Context(), userId, roleId); err != nil {
		c.logger.Error("移除用户角色失败", zap.Error(err))
		response.InternalServerError(ctx, "移除用户角色失败: "+err.Error())
		return
	}

	response.Success(ctx, nil)
}

// GetUserRoles 获取用户的所有角色
//
//	@Summary		获取用户的所有角色
//	@Description	获取用户的所有角色
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer {token}"
//	@Param			userId			query		int		true	"用户ID"
//	@Success		200				{object}	response.Response{data=[]response.RoleResponse}
//	@Failure		400				{object}	response.Response	"参数错误"
//	@Router			/api/v1/role/user [get]
func (c *roleController) GetUserRoles(ctx *gin.Context) {
	userIdStr := ctx.Query("userId")

	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "用户ID格式错误")
		return
	}

	roles, err := c.roleService.GetUserRoles(ctx.Request.Context(), userId)
	if err != nil {
		c.logger.Error("获取用户角色失败", zap.Error(err))
		response.InternalServerError(ctx, "获取用户角色失败: "+err.Error())
		return
	}

	response.Success(ctx, roles)
}

// AddRolePermission 为角色添加权限
//
//	@Summary		为角色添加权限
//	@Description	为角色添加权限（支持通配符）
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string								true	"Bearer {token}"
//	@Param			body			body		request.AddRolePermissionRequest	true	"权限信息"
//	@Success		200				{object}	response.Response
//	@Failure		400				{object}	response.Response	"参数错误"
//	@Router			/api/v1/role/permission [post]
func (c *roleController) AddRolePermission(ctx *gin.Context) {
	var req request.AddRolePermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误: "+err.Error())
		return
	}

	if err := c.roleService.AddRolePermission(ctx.Request.Context(), req.RoleKey, req.Resource, req.Action); err != nil {
		c.logger.Error("为角色添加权限失败", zap.Error(err))
		response.InternalServerError(ctx, "为角色添加权限失败: "+err.Error())
		return
	}

	response.Success(ctx, nil)
}

// DeleteRolePermission 删除角色权限
//
//	@Summary		删除角色权限
//	@Description	删除角色的指定权限
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer {token}"
//	@Param			roleKey			query		string	true	"角色标识"
//	@Param			resource		query		string	true	"资源路径"
//	@Param			action			query		string	true	"操作类型"
//	@Success		200				{object}	response.Response
//	@Failure		400				{object}	response.Response	"参数错误"
//	@Router			/api/v1/role/permission [delete]
func (c *roleController) DeleteRolePermission(ctx *gin.Context) {
	roleKey := ctx.Query("roleKey")
	resource := ctx.Query("resource")
	action := ctx.Query("action")

	if err := c.roleService.DeleteRolePermission(ctx.Request.Context(), roleKey, resource, action); err != nil {
		c.logger.Error("删除角色权限失败", zap.Error(err))
		response.InternalServerError(ctx, "删除角色权限失败: "+err.Error())
		return
	}

	response.Success(ctx, nil)
}

// GetRolePermissions 获取角色的所有权限
//
//	@Summary		获取角色的所有权限
//	@Description	获取角色的所有权限
//	@Tags			角色管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer {token}"
//	@Param			roleKey			query		string	true	"角色标识"
//	@Success		200				{object}	response.Response{data=[]response.RolePermissionResponse}
//	@Failure		400				{object}	response.Response	"参数错误"
//	@Router			/api/v1/role/permissions [get]
func (c *roleController) GetRolePermissions(ctx *gin.Context) {
	roleKey := ctx.Query("roleKey")

	permissions, err := c.roleService.GetRolePermissions(ctx.Request.Context(), roleKey)
	if err != nil {
		c.logger.Error("获取角色权限失败", zap.Error(err))
		response.InternalServerError(ctx, "获取角色权限失败: "+err.Error())
		return
	}

	response.Success(ctx, permissions)
}
