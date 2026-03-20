package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/force-c/nai-tizi/internal/domain/model"
	"github.com/force-c/nai-tizi/internal/domain/request"
	"github.com/force-c/nai-tizi/internal/infrastructure/storage"
	"github.com/force-c/nai-tizi/internal/logger"
	"github.com/force-c/nai-tizi/internal/utils/pagination"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AttachmentService 附件管理服务接口
type AttachmentService interface {
	// UploadFile 上传文件（步骤1：只上传文件，返回附件ID）
	UploadFile(ctx context.Context, req *request.UploadFileRequest) (*model.Attachment, error)

	// BindToBusiness 绑定附件到业务（步骤2：绑定业务信息）
	BindToBusiness(ctx context.Context, attachmentId int64, req *request.BindAttachmentToBusinessRequest) error

	// Download 下载附件
	Download(ctx context.Context, attachmentId int64) (io.ReadCloser, string, error)

	// Delete 删除附件
	Delete(ctx context.Context, attachmentId int64) error

	// GetURL 获取附件 URL
	GetURL(ctx context.Context, attachmentId int64, expires time.Duration) (string, error)

	// GetById 根据 ID 查询附件
	GetById(ctx context.Context, attachmentId int64) (*model.Attachment, error)

	// ListByBusiness 根据业务查询附件列表
	ListByBusiness(ctx context.Context, businessType, businessId string) ([]*model.Attachment, error)

	// Page 分页查询附件列表
	Page(ctx context.Context, pageNum, pageSize int, fileName, fileType, businessType string) (*pagination.Page[model.Attachment], error)

	// CleanExpired 清理过期附件
	CleanExpired(ctx context.Context) error
}

type attachmentService struct {
	db                *gorm.DB
	storageManager    storage.StorageManager
	storageEnvService StorageEnvService
	logger            logger.Logger
}

// NewAttachmentService 创建附件服务实例
func NewAttachmentService(
	db *gorm.DB,
	storageManager storage.StorageManager,
	storageEnvService StorageEnvService,
	logger logger.Logger,
) AttachmentService {
	return &attachmentService{
		db:                db,
		storageManager:    storageManager,
		storageEnvService: storageEnvService,
		logger:            logger,
	}
}

// UploadFile 上传文件（步骤1：只上传文件）
func (s *attachmentService) UploadFile(ctx context.Context, req *request.UploadFileRequest) (*model.Attachment, error) {
	// 1. 根据 envCode 获取存储环境（如果未指定，使用默认环境）
	var env *model.StorageEnv
	var err error

	if req.EnvCode != "" {
		env, err = s.storageEnvService.GetByCode(ctx, req.EnvCode)
	} else {
		// 使用默认环境
		env, err = s.storageEnvService.GetDefault(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("获取存储环境失败: %w", err)
	}

	// 2. 验证文件类型（基于扩展名）
	fileExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(req.File.Filename), "."))

	// 3. 获取 Storage 实例
	stor, err := s.storageManager.GetStorage(env.ID)
	if err != nil {
		return nil, fmt.Errorf("获取存储实例失败: %w", err)
	}

	// 5. 生成文件 Key（使用临时业务类型）
	fileKey := s.generateFileKey(req.File.Filename, "temp")

	// 6. 打开文件
	file, err := req.File.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 7. 上传文件
	if err := stor.Upload(ctx, fileKey, file, req.File.Size); err != nil {
		return nil, fmt.Errorf("上传文件失败: %w", err)
	}

	// 8. 保存附件记录（暂不绑定业务信息）
	attachment := &model.Attachment{
		EnvId:    env.ID,
		FileName: req.File.Filename,
		FileKey:  fileKey,
		FileSize: req.File.Size,
		FileType: req.File.Header.Get("Content-Type"),
		FileExt:  fileExt,
		Status:   0, // 0 = 正常
	}

	if err := s.db.Create(attachment).Error; err != nil {
		// 回滚：删除已上传的文件
		if delErr := stor.Delete(ctx, fileKey); delErr != nil {
			s.logger.Error("删除文件失败", zap.Error(delErr))
		}
		return nil, fmt.Errorf("保存附件记录失败: %w", err)
	}

	s.logger.Info("上传文件成功",
		zap.Int64("attachmentId", attachment.ID),
		zap.String("fileName", attachment.FileName),
		zap.Int64("fileSize", attachment.FileSize))

	return attachment, nil
}

// BindToBusiness 绑定附件到业务（步骤2：绑定业务信息）
func (s *attachmentService) BindToBusiness(ctx context.Context, attachmentId int64, req *request.BindAttachmentToBusinessRequest) error {
	// 1. 查询附件记录
	attachment, err := s.GetById(ctx, attachmentId)
	if err != nil {
		return err
	}

	// 2. 如果需要公开访问，生成访问 URL
	var accessUrl string
	if req.IsPublic {
		stor, err := s.storageManager.GetStorage(attachment.ID)
		if err != nil {
			return fmt.Errorf("获取存储实例失败: %w", err)
		}
		accessUrl, err = stor.GetURL(ctx, attachment.FileKey, 0) // 0 表示永久 URL
		if err != nil {
			s.logger.Warn("获取访问 URL 失败", zap.Error(err))
		}
	}

	// 3. 将 Metadata map 转换为 json.RawMessage
	var metadata *json.RawMessage
	if req.Metadata != nil && len(req.Metadata) > 0 {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			return fmt.Errorf("元数据序列化失败: %w", err)
		}
		rawMsg := json.RawMessage(metadataBytes)
		metadata = &rawMsg
	}

	// 4. 更新附件记录
	updates := map[string]interface{}{
		"business_type":  req.BusinessType,
		"business_id":    req.BusinessId,
		"business_field": req.BusinessField,
		"is_public":      req.IsPublic,
		"access_url":     accessUrl,
		"metadata":       metadata,
	}

	if req.ExpireTime != nil {
		updates["expire_time"] = *req.ExpireTime
	}

	if err := s.db.Model(&model.Attachment{}).
		Where("attachment_id = ?", attachmentId).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("绑定附件到业务失败: %w", err)
	}

	s.logger.Info("绑定附件到业务成功",
		zap.Int64("attachmentId", attachmentId),
		zap.String("businessType", req.BusinessType),
		zap.String("businessId", req.BusinessId))

	return nil
}

// Download 下载附件
func (s *attachmentService) Download(ctx context.Context, attachmentId int64) (io.ReadCloser, string, error) {
	// 1. 查询附件记录
	attachment, err := s.GetById(ctx, attachmentId)
	if err != nil {
		return nil, "", err
	}

	// 2. 获取 Storage 实例
	stor, err := s.storageManager.GetStorage(attachment.ID)
	if err != nil {
		return nil, "", fmt.Errorf("获取存储实例失败: %w", err)
	}

	// 3. 下载文件
	reader, err := stor.Download(ctx, attachment.FileKey)
	if err != nil {
		return nil, "", fmt.Errorf("下载文件失败: %w", err)
	}

	return reader, attachment.FileName, nil
}

// Delete 删除附件
func (s *attachmentService) Delete(ctx context.Context, attachmentId int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 查询附件记录
		var attachment model.Attachment
		if err := tx.Where("attachment_id = ?", attachmentId).First(&attachment).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("附件不存在")
			}
			return fmt.Errorf("查询附件失败: %w", err)
		}

		// 2. 获取 Storage 实例
		stor, err := s.storageManager.GetStorage(attachment.ID)
		if err != nil {
			return fmt.Errorf("获取存储实例失败: %w", err)
		}

		// 3. 删除文件
		if err := stor.Delete(ctx, attachment.FileKey); err != nil {
			s.logger.Warn("删除存储文件失败", zap.Error(err))
			// 继续删除数据库记录
		}

		// 4. 删除数据库记录（软删除）
		if err := tx.Model(&model.Attachment{}).
			Where("attachment_id = ?", attachmentId).
			Update("status", "1").Error; err != nil {
			return fmt.Errorf("删除附件记录失败: %w", err)
		}

		s.logger.Info("删除附件成功", zap.Int64("attachmentId", attachmentId))
		return nil
	})
}

// GetURL 获取附件 URL
func (s *attachmentService) GetURL(ctx context.Context, attachmentId int64, expires time.Duration) (string, error) {
	// 1. 查询附件记录
	attachment, err := s.GetById(ctx, attachmentId)
	if err != nil {
		return "", err
	}

	// 2. 如果是公开文件且有访问 URL，直接返回
	if attachment.IsPublic && attachment.AccessUrl != "" && expires == 0 {
		return attachment.AccessUrl, nil
	}

	// 3. 获取 Storage 实例
	stor, err := s.storageManager.GetStorage(attachment.ID)
	if err != nil {
		return "", fmt.Errorf("获取存储实例失败: %w", err)
	}

	// 4. 生成访问 URL
	url, err := stor.GetURL(ctx, attachment.FileKey, expires)
	if err != nil {
		return "", fmt.Errorf("生成访问 URL 失败: %w", err)
	}

	return url, nil
}

// GetById 根据 ID 查询附件
func (s *attachmentService) GetById(ctx context.Context, attachmentId int64) (*model.Attachment, error) {
	var attachment model.Attachment
	if err := s.db.Where("attachment_id = ? AND status = ?", attachmentId, "0").First(&attachment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("附件不存在")
		}
		return nil, fmt.Errorf("查询附件失败: %w", err)
	}
	return &attachment, nil
}

// ListByBusiness 根据业务查询附件列表
func (s *attachmentService) ListByBusiness(ctx context.Context, businessType, businessId string) ([]*model.Attachment, error) {
	var attachments []*model.Attachment

	query := s.db.Where("business_type = ? AND business_id = ? AND status = ?", businessType, businessId, "0")

	if err := query.Order("create_time DESC").Find(&attachments).Error; err != nil {
		return nil, fmt.Errorf("查询附件列表失败: %w", err)
	}

	return attachments, nil
}

// Page 分页查询附件列表
func (s *attachmentService) Page(ctx context.Context, pageNum, pageSize int, fileName, fileType, businessType string) (*pagination.Page[model.Attachment], error) {
	query := s.db.Model(&model.Attachment{}).Where("status = ?", "0")

	// 添加过滤条件
	if fileName != "" {
		query = query.Where("file_name LIKE ?", "%"+fileName+"%")
	}
	if fileType != "" {
		query = query.Where("file_type = ?", fileType)
	}
	if businessType != "" {
		query = query.Where("business_type = ?", businessType)
	}

	// 构建 PageQuery
	pageQuery := &pagination.PageQuery{
		PageNum:  pageNum,
		PageSize: pageSize,
	}

	// 使用 Paginator 进行分页
	page, err := pagination.New[model.Attachment](query, pageQuery).Find()
	if err != nil {
		s.logger.Error("分页查询附件列表失败", zap.Error(err))
		return nil, fmt.Errorf("分页查询附件列表失败: %w", err)
	}

	return page, nil
}

// CleanExpired 清理过期附件
func (s *attachmentService) CleanExpired(ctx context.Context) error {
	// 查询过期的附件
	var attachments []*model.Attachment
	if err := s.db.Where("expire_time IS NOT NULL AND expire_time < ? AND status = ?",
		time.Now(), "0").Find(&attachments).Error; err != nil {
		return fmt.Errorf("查询过期附件失败: %w", err)
	}

	s.logger.Info("开始清理过期附件", zap.Int("count", len(attachments)))

	// 删除过期附件
	for _, attachment := range attachments {
		if err := s.Delete(ctx, attachment.ID); err != nil {
			s.logger.Error("删除过期附件失败",
				zap.Int64("attachmentId", attachment.ID),
				zap.Error(err))
			continue
		}
	}

	s.logger.Info("清理过期附件完成", zap.Int("count", len(attachments)))
	return nil
}

// generateFileKey 生成文件 Key
func (s *attachmentService) generateFileKey(filename, businessType string) string {
	// 格式：{businessType}/{date}/{timestamp}_{filename}
	now := time.Now()
	date := now.Format("20060102")
	timestamp := now.UnixNano()

	// 清理文件名中的特殊字符
	cleanFilename := filepath.Base(filename)

	return fmt.Sprintf("%s/%s/%d_%s", businessType, date, timestamp, cleanFilename)
}
