package controller

import (
	"strconv"

	"github.com/force-c/nai-tizi/internal/container"
	"github.com/force-c/nai-tizi/internal/domain/request"
	"github.com/force-c/nai-tizi/internal/domain/response"
	"github.com/force-c/nai-tizi/internal/service"
	"github.com/gin-gonic/gin"
)

// ApiPermissionController API 权限控制器。
type ApiPermissionController interface {
	Tree(ctx *gin.Context)
	List(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
	GetRolePermissions(ctx *gin.Context)
	AssignRolePermissions(ctx *gin.Context)
	GetUserPermissions(ctx *gin.Context)
	AssignUserPermissions(ctx *gin.Context)
}

type apiPermissionController struct {
	service service.ApiPermissionService
}

func NewApiPermissionController(c container.Container) ApiPermissionController {
	return &apiPermissionController{
		service: service.NewApiPermissionService(c.GetDB(), c.GetCasbin(), c.GetLogger()),
	}
}

func (c *apiPermissionController) Tree(ctx *gin.Context) {
	tree, err := c.service.Tree(ctx.Request.Context())
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Success(ctx, tree)
}

func (c *apiPermissionController) List(ctx *gin.Context) {
	list, err := c.service.List(ctx.Request.Context())
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Success(ctx, list)
}

func (c *apiPermissionController) Create(ctx *gin.Context) {
	var req request.ApiPermissionSaveRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误: "+err.Error())
		return
	}
	userId := currentUserId(ctx)
	permission, err := c.service.Create(ctx.Request.Context(), &req, userId)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Success(ctx, permission)
}

func (c *apiPermissionController) Update(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "权限ID格式错误")
		return
	}
	var req request.ApiPermissionSaveRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误: "+err.Error())
		return
	}
	if err := c.service.Update(ctx.Request.Context(), id, &req, currentUserId(ctx)); err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Success(ctx, nil)
}

func (c *apiPermissionController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "权限ID格式错误")
		return
	}
	if err := c.service.Delete(ctx.Request.Context(), id); err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Success(ctx, nil)
}

func (c *apiPermissionController) GetRolePermissions(ctx *gin.Context) {
	roleId, err := strconv.ParseInt(ctx.Param("roleId"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "角色ID格式错误")
		return
	}
	ids, err := c.service.GetRolePermissionIds(ctx.Request.Context(), roleId)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Success(ctx, ids)
}

func (c *apiPermissionController) AssignRolePermissions(ctx *gin.Context) {
	roleId, err := strconv.ParseInt(ctx.Param("roleId"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "角色ID格式错误")
		return
	}
	var req request.ApiPermissionAssignRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误: "+err.Error())
		return
	}
	if err := c.service.AssignRolePermissions(ctx.Request.Context(), roleId, req.PermissionIds, currentUserId(ctx)); err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Success(ctx, nil)
}

func (c *apiPermissionController) GetUserPermissions(ctx *gin.Context) {
	userId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "用户ID格式错误")
		return
	}
	ids, err := c.service.GetUserPermissionIds(ctx.Request.Context(), userId)
	if err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Success(ctx, ids)
}

func (c *apiPermissionController) AssignUserPermissions(ctx *gin.Context) {
	userId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "用户ID格式错误")
		return
	}
	var req request.ApiPermissionAssignRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误: "+err.Error())
		return
	}
	if err := c.service.AssignUserPermissions(ctx.Request.Context(), userId, req.PermissionIds, currentUserId(ctx)); err != nil {
		response.InternalServerError(ctx, err.Error())
		return
	}
	response.Success(ctx, nil)
}

func currentUserId(ctx *gin.Context) int64 {
	value, ok := ctx.Get("userId")
	if !ok {
		return 0
	}
	userId, ok := value.(int64)
	if !ok {
		return 0
	}
	return userId
}
