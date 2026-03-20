package controller

import (
	"time"

	"github.com/force-c/nai-tizi/internal/container"
	"github.com/force-c/nai-tizi/internal/domain/request"
	"github.com/force-c/nai-tizi/internal/domain/response"
	"github.com/force-c/nai-tizi/internal/logger"
	"github.com/force-c/nai-tizi/internal/service"
	"github.com/force-c/nai-tizi/internal/utils"
	_ "github.com/force-c/nai-tizi/internal/utils/pagination"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AttachmentController interface {
	UploadFile(ctx *gin.Context)                // 上传文件
	BindAttachmentToBusiness(ctx *gin.Context)  // 绑定附件到业务
	DownloadAttachment(ctx *gin.Context)        // 下载附件
	DeleteAttachment(ctx *gin.Context)          // 删除附件
	GetAttachmentURL(ctx *gin.Context)          // 获取附件访问URL
	GetAttachment(ctx *gin.Context)             // 获取附件详情
	ListAttachmentsByBusiness(ctx *gin.Context) // 根据业务查询附件列表
	PageAttachments(ctx *gin.Context)           // 分页查询附件列表
}

type attachmentController struct {
	attachmentService service.AttachmentService
	logger            logger.Logger
}

func NewAttachmentController(c container.Container) AttachmentController {
	storageEnvService := service.NewStorageEnvService(c.GetDB(), c.GetLogger())
	return &attachmentController{
		attachmentService: service.NewAttachmentService(c.GetDB(), c.GetStorageManager(), storageEnvService, c.GetLogger()),
		logger:            c.GetLogger(),
	}
}

// UploadFile 上传文件（步骤1：只上传文件）
//
//	@Summary		上传文件
//	@Description	上传文件到指定存储环境（步骤1：只上传文件，返回附件ID）
//	@Tags			附件管理
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer {token}"
//	@Param			file			formData	file	true	"文件"
//	@Param			envCode			formData	string	false	"存储环境编码（不传则使用默认环境）"
//	@Success		200				{object}	response.Response{data=response.AttachmentResponse}
//	@Router			/api/v1/attachment/upload-file [post]
func (c *attachmentController) UploadFile(ctx *gin.Context) {
	var req request.UploadFileRequest
	if err := ctx.ShouldBind(&req); err != nil {
		response.BadRequest(ctx, "参数错误: "+err.Error())
		return
	}

	attachment, err := c.attachmentService.UploadFile(ctx.Request.Context(), &req)
	if err != nil {
		c.logger.Error("上传文件失败", zap.Error(err))
		response.InternalServerError(ctx, "上传文件失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(ctx, "上传文件成功", attachment)
}

// BindAttachmentToBusiness 绑定附件到业务（步骤2：绑定业务信息）
//
//	@Summary		绑定附件到业务
//	@Description	将附件绑定到指定业务（步骤2：绑定业务信息）
//	@Tags			附件管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string									true	"Bearer {token}"
//	@Param			attachmentId	path		int										true	"附件ID"
//	@Param			body			body		request.BindAttachmentToBusinessRequest	true	"业务信息"
//	@Success		200				{object}	response.Response
//	@Router			/api/v1/attachment/{attachmentId}/bind [post]
func (c *attachmentController) BindAttachmentToBusiness(ctx *gin.Context) {
	attachmentId, err := utils.ParseInt64Param(ctx, "attachmentId", "required")
	if err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	var req request.BindAttachmentToBusinessRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误: "+err.Error())
		return
	}

	if err := c.attachmentService.BindToBusiness(ctx.Request.Context(), attachmentId, &req); err != nil {
		c.logger.Error("绑定附件到业务失败", zap.Error(err))
		response.InternalServerError(ctx, "绑定附件到业务失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(ctx, "绑定附件到业务成功", nil)
}

// DownloadAttachment 下载附件
//
//	@Summary		下载附件
//	@Description	下载指定附件
//	@Tags			附件管理
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			Authorization	header	string	true	"Bearer {token}"
//	@Param			attachmentId	path	int		true	"附件ID"
//	@Success		200				{file}	binary
//	@Router			/api/v1/attachment/{attachmentId}/download [get]
func (c *attachmentController) DownloadAttachment(ctx *gin.Context) {
	attachmentId, err := utils.ParseInt64Param(ctx, "attachmentId", "required")
	if err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	reader, filename, err := c.attachmentService.Download(ctx.Request.Context(), attachmentId)
	if err != nil {
		c.logger.Error("下载附件失败", zap.Error(err))
		response.InternalServerError(ctx, "下载附件失败: "+err.Error())
		return
	}
	defer reader.Close()

	ctx.Header("Content-Disposition", "attachment; filename="+filename)
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.DataFromReader(200, -1, "application/octet-stream", reader, nil)
}

// DeleteAttachment 删除附件
//
//	@Summary		删除附件
//	@Description	删除指定附件
//	@Tags			附件管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer {token}"
//	@Param			attachmentId	path		int		true	"附件ID"
//	@Success		200				{object}	response.Response
//	@Router			/api/v1/attachment/{attachmentId} [delete]
func (c *attachmentController) DeleteAttachment(ctx *gin.Context) {
	attachmentId, err := utils.ParseInt64Param(ctx, "attachmentId", "required")
	if err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	if err := c.attachmentService.Delete(ctx.Request.Context(), attachmentId); err != nil {
		c.logger.Error("删除附件失败", zap.Error(err))
		response.InternalServerError(ctx, "删除附件失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(ctx, "删除附件成功", nil)
}

// GetAttachmentURL 获取附件访问URL
//
//	@Summary		获取附件访问URL
//	@Description	获取附件的访问URL（支持临时URL）
//	@Tags			附件管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer {token}"
//	@Param			attachmentId	path		int		true	"附件ID"
//	@Param			expires			query		int		false	"过期时间（秒）"	default(3600)
//	@Success		200				{object}	response.Response{data=response.AttachmentURLResponse}
//	@Router			/api/v1/attachment/{attachmentId}/url [get]
func (c *attachmentController) GetAttachmentURL(ctx *gin.Context) {
	attachmentId, err := utils.ParseInt64Param(ctx, "attachmentId", "required")
	if err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	var req request.GetAttachmentURLRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数错误: "+err.Error())
		return
	}

	expires := time.Duration(req.Expires) * time.Second
	url, err := c.attachmentService.GetURL(ctx.Request.Context(), attachmentId, expires)
	if err != nil {
		c.logger.Error("获取附件URL失败", zap.Error(err))
		response.InternalServerError(ctx, "获取附件URL失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(ctx, "获取附件URL成功", gin.H{
		"url":     url,
		"expires": req.Expires,
	})
}

// GetAttachment 获取附件详情
//
//	@Summary		获取附件详情
//	@Description	根据附件ID获取附件详情
//	@Tags			附件管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer {token}"
//	@Param			attachmentId	path		int		true	"附件ID"
//	@Success		200				{object}	response.Response{data=response.AttachmentResponse}
//	@Router			/api/v1/attachment/{attachmentId} [get]
func (c *attachmentController) GetAttachment(ctx *gin.Context) {
	attachmentId, err := utils.ParseInt64Param(ctx, "attachmentId", "required")
	if err != nil {
		response.BadRequest(ctx, err.Error())
		return
	}

	attachment, err := c.attachmentService.GetById(ctx.Request.Context(), attachmentId)
	if err != nil {
		c.logger.Error("获取附件详情失败", zap.Error(err))
		response.InternalServerError(ctx, "获取附件详情失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(ctx, "获取附件详情成功", attachment)
}

// ListAttachmentsByBusiness 根据业务查询附件列表
//
//	@Summary		根据业务查询附件列表
//	@Description	根据业务类型和业务ID查询附件列表
//	@Tags			附件管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer {token}"
//	@Param			businessType	query		string	true	"业务类型"
//	@Param			businessId		query		string	true	"业务ID"
//	@Success		200				{object}	response.Response{data=[]response.AttachmentResponse}
//	@Router			/api/v1/attachment/business [get]
func (c *attachmentController) ListAttachmentsByBusiness(ctx *gin.Context) {
	var req request.ListAttachmentsByBusinessRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数错误: "+err.Error())
		return
	}

	attachments, err := c.attachmentService.ListByBusiness(ctx.Request.Context(), req.BusinessType, req.BusinessId)
	if err != nil {
		c.logger.Error("查询业务附件列表失败", zap.Error(err))
		response.InternalServerError(ctx, "查询业务附件列表失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(ctx, "查询业务附件列表成功", attachments)
}

// PageAttachments 分页查询附件列表
//
//	@Summary		分页查询附件列表
//	@Description	分页查询附件列表，支持按文件名、文件类型、业务类型筛选
//	@Tags			附件管理
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string							true	"Bearer {token}"
//	@Param			body			body		request.PageAttachmentsRequest	true	"查询参数"
//	@Success		200				{object}	response.Response{data=response.PageResponse}
//	@Router			/api/v1/attachment/page [post]
func (c *attachmentController) PageAttachments(ctx *gin.Context) {
	var req request.PageAttachmentsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误: "+err.Error())
		return
	}

	page, err := c.attachmentService.Page(
		ctx.Request.Context(),
		req.PageNum,
		req.PageSize,
		req.FileName,
		req.FileType,
		req.BusinessType,
	)
	if err != nil {
		c.logger.Error("分页查询附件列表失败", zap.Error(err))
		response.InternalServerError(ctx, "分页查询附件列表失败: "+err.Error())
		return
	}

	response.Success(ctx, page)
}
