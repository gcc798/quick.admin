package controller

import (
	"github.com/force-c/nai-tizi/internal/container"
	"github.com/force-c/nai-tizi/internal/domain/model"
	"github.com/force-c/nai-tizi/internal/domain/request"
	"github.com/force-c/nai-tizi/internal/domain/response"
	"github.com/force-c/nai-tizi/internal/logger"
	"github.com/force-c/nai-tizi/internal/service"
	"github.com/force-c/nai-tizi/internal/utils"
	_ "github.com/force-c/nai-tizi/internal/utils/pagination"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type StorageEnvController interface {
	CreateStorageEnv(ctx *gin.Context)         // 创建存储环境
	UpdateStorageEnv(ctx *gin.Context)         // 更新存储环境
	DeleteStorageEnv(ctx *gin.Context)         // 删除存储环境
	GetStorageEnv(ctx *gin.Context)            // 获取存储环境详情
	SetDefaultStorageEnv(ctx *gin.Context)     // 设置默认存储环境
	GetDefaultStorageEnv(ctx *gin.Context)     // 获取默认存储环境
	PageStorageEnv(ctx *gin.Context)           // 分页查询存储环境列表
	TestStorageEnvConnection(ctx *gin.Context) // 测试存储环境连接
}

type storageEnvController struct {
	storageEnvService service.StorageEnvService
	logger            logger.Logger
}

func NewStorageEnvController(c container.Container) StorageEnvController {
	return &storageEnvController{
		storageEnvService: service.NewStorageEnvService(c.GetDB(), c.GetLogger()),
		logger:            c.GetLogger(),
	}
}

// CreateStorageEnv 创建存储环境
//
//	@Summary		创建存储环境
//	@Description	创建新的存储环境配置
//	@Tags			存储环境管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string							true	"Bearer {token}"
//	@Param			body			body		request.CreateStorageEnvRequest	true	"存储环境信息"
//	@Success		200				{object}	response.Response{data=response.StorageEnvResponse}
//	@Router			/api/v1/storage-env [post]
func (c *storageEnvController) CreateStorageEnv(ctx *gin.Context) {
	var req request.CreateStorageEnvRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误: "+err.Error())
		return
	}

	userId, _ := ctx.Get("userId")

	env := &model.StorageEnv{
		EnvName:     req.EnvName,
		EnvCode:     req.EnvCode,
		StorageType: req.StorageType,
		IsDefault:   req.IsDefault,
		Status:      req.Status,
		Config:      req.Config,
		Remark:      req.Remark,
	}
	env.CreateBy = userId.(int64)

	if err := c.storageEnvService.Create(ctx.Request.Context(), env); err != nil {
		c.logger.Error("创建存储环境失败", zap.Error(err))
		response.InternalServerError(ctx, "创建存储环境失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(ctx, "创建存储环境成功", env)
}

// UpdateStorageEnv 更新存储环境
//
//	@Summary		更新存储环境
//	@Description	更新存储环境配置
//	@Tags			存储环境管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string							true	"Bearer {token}"
//	@Param			body			body		request.UpdateStorageEnvRequest	true	"存储环境信息"
//	@Success		200				{object}	response.Response
//	@Router			/api/v1/storage-env [put]
func (c *storageEnvController) UpdateStorageEnv(ctx *gin.Context) {
	var req request.UpdateStorageEnvRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误: "+err.Error())
		return
	}

	userId, _ := ctx.Get("userId")

	env := &model.StorageEnv{
		ID:          req.ID,
		EnvName:     req.EnvName,
		EnvCode:     req.EnvCode,
		StorageType: req.StorageType,
		IsDefault:   req.IsDefault,
		Status:      req.Status,
		Config:      req.Config,
		Remark:      req.Remark,
	}
	env.UpdateBy = userId.(int64)

	if err := c.storageEnvService.Update(ctx.Request.Context(), env); err != nil {
		c.logger.Error("更新存储环境失败", zap.Error(err))
		response.InternalServerError(ctx, "更新存储环境失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(ctx, "更新存储环境成功", nil)
}

// DeleteStorageEnv 删除存储环境
//
//	@Summary		删除存储环境
//	@Description	删除存储环境（不能删除默认环境）
//	@Tags			存储环境管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer {token}"
//	@Param			id			path		int		true	"环境ID"
//	@Success		200				{object}	response.Response
//	@Router			/api/v1/storage-env/{id} [delete]
func (c *storageEnvController) DeleteStorageEnv(ctx *gin.Context) {
	envId, err := utils.ParseInt64Param(ctx, "id", "required")
	if err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	if err := c.storageEnvService.Delete(ctx.Request.Context(), envId); err != nil {
		c.logger.Error("删除存储环境失败", zap.Error(err))
		response.InternalServerError(ctx, "删除存储环境失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(ctx, "删除存储环境成功", nil)
}

// GetStorageEnv 获取存储环境详情
//
//	@Summary		获取存储环境详情
//	@Description	根据环境ID获取存储环境详情
//	@Tags			存储环境管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer {token}"
//	@Param			id			path		int		true	"环境ID"
//	@Success		200				{object}	response.Response{data=response.StorageEnvResponse}
//	@Router			/api/v1/storage-env/{id} [get]
func (c *storageEnvController) GetStorageEnv(ctx *gin.Context) {
	envId, err := utils.ParseInt64Param(ctx, "id", "required")
	if err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	env, err := c.storageEnvService.GetById(ctx.Request.Context(), envId)
	if err != nil {
		c.logger.Error("获取存储环境详情失败", zap.Error(err))
		response.InternalServerError(ctx, "获取存储环境详情失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(ctx, "获取存储环境详情成功", env)
}

// SetDefaultStorageEnv 设置默认存储环境
//
//	@Summary		设置默认存储环境
//	@Description	设置指定环境为默认存储环境
//	@Tags			存储环境管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string								true	"Bearer {token}"
//	@Param			body			body		request.SetDefaultStorageEnvRequest	true	"环境ID"
//	@Success		200				{object}	response.Response
//	@Router			/api/v1/storage-env/default [post]
func (c *storageEnvController) SetDefaultStorageEnv(ctx *gin.Context) {
	var req request.SetDefaultStorageEnvRequest
	if err := utils.BindJSONWithTypeCasting(ctx, &req); err != nil {
		response.BadRequest(ctx, "参数错误: "+err.Error())
		return
	}

	if err := c.storageEnvService.SetDefault(ctx.Request.Context(), req.ID); err != nil {
		c.logger.Error("设置默认存储环境失败", zap.Error(err))
		response.InternalServerError(ctx, "设置默认存储环境失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(ctx, "设置默认存储环境成功", nil)
}

// GetDefaultStorageEnv 获取默认存储环境
//
//	@Summary		获取默认存储环境
//	@Description	获取当前默认的存储环境
//	@Tags			存储环境管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer {token}"
//	@Success		200				{object}	response.Response{data=response.StorageEnvResponse}
//	@Router			/api/v1/storage-env/default [get]
func (c *storageEnvController) GetDefaultStorageEnv(ctx *gin.Context) {
	env, err := c.storageEnvService.GetDefault(ctx.Request.Context())
	if err != nil {
		c.logger.Error("获取默认存储环境失败", zap.Error(err))
		response.InternalServerError(ctx, "获取默认存储环境失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(ctx, "获取默认存储环境成功", env)
}

// PageStorageEnv 分页查询存储环境列表
//
//	@Summary		分页查询存储环境列表
//	@Description	使用 Paginator 分页查询存储环境列表
//	@Tags			存储环境管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string							true	"Bearer {token}"
//	@Param			body			body		request.PageStorageEnvsRequest	true	"查询参数"
//	@Success		200				{object}	response.Response{data=object}
//	@Router			/api/v1/storage-env/page [post]
func (c *storageEnvController) PageStorageEnv(ctx *gin.Context) {
	var req request.PageStorageEnvsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误: "+err.Error())
		return
	}

	page, err := c.storageEnvService.Page(ctx.Request.Context(), req.PageNum, req.PageSize, req.Name, req.StorageType)
	if err != nil {
		c.logger.Error("分页查询存储环境列表失败", zap.Error(err))
		response.InternalServerError(ctx, "分页查询存储环境列表失败: "+err.Error())
		return
	}

	response.Success(ctx, page)
}

// TestStorageEnvConnection 测试存储环境连接
//
//	@Summary		测试存储环境连接
//	@Description	测试指定存储环境的连接是否正常
//	@Tags			存储环境管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string									true	"Bearer {token}"
//	@Param			id			path		int										true	"环境ID"
//	@Param			body			body		request.TestStorageEnvConnectionRequest	false	"测试参数"
//	@Success		200				{object}	response.Response
//	@Router			/api/v1/storage-env/{id}/test [post]
func (c *storageEnvController) TestStorageEnvConnection(ctx *gin.Context) {
	envId, err := utils.ParseInt64Param(ctx, "id", "required")
	if err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	if err := c.storageEnvService.TestConnection(ctx.Request.Context(), envId); err != nil {
		c.logger.Error("测试存储环境连接失败", zap.Error(err))
		response.InternalServerError(ctx, "测试存储环境连接失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(ctx, "存储环境连接测试成功", nil)
}
