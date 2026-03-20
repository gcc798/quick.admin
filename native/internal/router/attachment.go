package router

import (
	"github.com/force-c/nai-tizi/internal/constants"
	"github.com/force-c/nai-tizi/internal/controller"
	"github.com/force-c/nai-tizi/internal/middleware"
	"github.com/gin-gonic/gin"
)

// registerAttachmentRoutes 注册附件管理路由
func registerAttachmentRoutes(r *gin.Engine, ctx *RouterContext) {
	// 初始化 controller
	attachmentController := controller.NewAttachmentController(ctx.Container)

	// 附件管理路由组（需要认证和权限）
	attachments := r.Group("/api/v1/attachment")
	attachments.Use(ctx.AuthMiddleware)
	{
		// 分两步上传（符合参数传递规范）
		// 步骤1：上传文件 - 需要 attachment.upload 权限
		attachments.POST("/upload-file", middleware.Permission(ctx.CasbinService, constants.ResourceAttachmentUpload), attachmentController.UploadFile)
		// 步骤2：绑定业务信息 - 需要 attachment.bind 权限
		attachments.POST("/:attachmentId/bind", middleware.Permission(ctx.CasbinService, constants.ResourceAttachmentBind), attachmentController.BindAttachmentToBusiness)

		// 附件查询 - 需要 attachment.read 权限
		attachments.GET("/:attachmentId", middleware.Permission(ctx.CasbinService, constants.ResourceAttachmentRead), attachmentController.GetAttachment)
		attachments.GET("/business", middleware.Permission(ctx.CasbinService, constants.ResourceAttachmentRead), attachmentController.ListAttachmentsByBusiness)
		attachments.POST("/page", middleware.Permission(ctx.CasbinService, constants.ResourceAttachmentRead), attachmentController.PageAttachments)

		// 附件下载 - 需要 attachment.download 权限
		attachments.GET("/:attachmentId/download", middleware.Permission(ctx.CasbinService, constants.ResourceAttachmentDownload), attachmentController.DownloadAttachment)

		// 获取附件URL - 需要 attachment.read 权限
		attachments.GET("/:attachmentId/url", middleware.Permission(ctx.CasbinService, constants.ResourceAttachmentRead), attachmentController.GetAttachmentURL)

		// 附件删除 - 需要 attachment.delete 权限
		attachments.DELETE("/:attachmentId", middleware.Permission(ctx.CasbinService, constants.ResourceAttachmentDelete), attachmentController.DeleteAttachment)
	}
}
