package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/force-c/nai-tizi/internal/domain/model"
	"github.com/force-c/nai-tizi/internal/domain/request"
	logging "github.com/force-c/nai-tizi/internal/logger"
	"github.com/force-c/nai-tizi/internal/utils/pagination"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ConfigService 配置服务接口
type ConfigService interface {
	// Create 创建配置
	Create(ctx context.Context, req *request.CreateConfigRequest) error

	// Update 更新配置
	Update(ctx context.Context, req *request.UpdateConfigRequest) error

	// Delete 删除配置
	Delete(ctx context.Context, id int64) error

	// BatchDelete 批量删除配置
	BatchDelete(ctx context.Context, ids []int64) error

	// GetById 根据ID查询配置
	GetById(ctx context.Context, id int64) (*model.Config, error)

	// Page 分页查询配置列表
	Page(ctx context.Context, pageNum, pageSize int, configCode, name string) (*pagination.Page[model.Config], error)

	// GetByCode 根据配置编码获取配置列表
	GetByCode(ctx context.Context, configCode string) ([]model.Config, error)

	// GetDataByCode 根据配置编码获取配置数据（返回第一个匹配的配置的data字段）
	GetDataByCode(ctx context.Context, configCode string) (json.RawMessage, error)
}

type configService struct {
	db     *gorm.DB
	logger logging.Logger
}

// NewConfigService 创建配置服务实例
func NewConfigService(db *gorm.DB, logger logging.Logger) ConfigService {
	return &configService{
		db:     db,
		logger: logger,
	}
}

// Create 创建配置
func (s *configService) Create(ctx context.Context, req *request.CreateConfigRequest) error {
	// 检查配置名称是否已存在
	exists, err := (&model.Config{}).CheckNameExists(s.db, req.Name)
	if err != nil {
		s.logger.Error("检查配置名称失败", zap.Error(err))
		return fmt.Errorf("检查配置名称失败: %w", err)
	}
	if exists {
		return fmt.Errorf("配置名称已存在: %s", req.Name)
	}

	// 创建配置
	config := &model.Config{
		Name:     req.Name,
		Code:     req.Code,
		Data:     req.Data,
		Remark:   req.Remark,
		CreateBy: req.CreateBy,
		UpdateBy: req.UpdateBy,
	}

	if err := config.Create(s.db); err != nil {
		s.logger.Error("创建配置失败", zap.Error(err))
		return fmt.Errorf("创建配置失败: %w", err)
	}

	s.logger.Info("创建配置成功",
		zap.Int64("id", config.ID),
		zap.String("name", config.Name),
		zap.String("code", config.Code))

	return nil
}

// Update 更新配置
func (s *configService) Update(ctx context.Context, req *request.UpdateConfigRequest) error {
	// 检查配置是否存在
	existingConfig, err := (&model.Config{}).FindByID(s.db, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("配置不存在")
		}
		s.logger.Error("查询配置失败", zap.Error(err))
		return fmt.Errorf("查询配置失败: %w", err)
	}

	// 检查配置名称是否被其他配置占用
	if req.Name != existingConfig.Name {
		exists, err := (&model.Config{}).CheckNameExistsExcludingSelf(s.db, req.ID, req.Name)
		if err != nil {
			s.logger.Error("检查配置名称失败", zap.Error(err))
			return fmt.Errorf("检查配置名称失败: %w", err)
		}
		if exists {
			return fmt.Errorf("配置名称已被占用: %s", req.Name)
		}
	}

	// 更新配置
	updates := map[string]interface{}{
		"name":      req.Name,
		"code":      req.Code,
		"data":      req.Data,
		"remark":    req.Remark,
		"update_by": req.UpdateBy,
	}

	if err := existingConfig.Update(s.db, req.ID, updates); err != nil {
		s.logger.Error("更新配置失败", zap.Error(err))
		return fmt.Errorf("更新配置失败: %w", err)
	}

	s.logger.Info("更新配置成功", zap.Int64("id", req.ID))
	return nil
}

// Delete 删除配置
func (s *configService) Delete(ctx context.Context, id int64) error {
	// 检查配置是否存在
	_, err := (&model.Config{}).FindByID(s.db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("配置不存在")
		}
		s.logger.Error("查询配置失败", zap.Error(err))
		return fmt.Errorf("查询配置失败: %w", err)
	}

	// 删除配置
	if err := (&model.Config{}).Delete(s.db, id); err != nil {
		s.logger.Error("删除配置失败", zap.Error(err))
		return fmt.Errorf("删除配置失败: %w", err)
	}

	s.logger.Info("删除配置成功", zap.Int64("id", id))
	return nil
}

// BatchDelete 批量删除配置
func (s *configService) BatchDelete(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return fmt.Errorf("配置ID列表不能为空")
	}

	// 批量删除
	rowsAffected, err := (&model.Config{}).BatchDelete(s.db, ids)
	if err != nil {
		s.logger.Error("批量删除配置失败", zap.Error(err))
		return fmt.Errorf("批量删除配置失败: %w", err)
	}

	s.logger.Info("批量删除配置成功", zap.Int64("count", rowsAffected))
	return nil
}

// GetById 根据ID查询配置
func (s *configService) GetById(ctx context.Context, id int64) (*model.Config, error) {
	config, err := (&model.Config{}).FindByID(s.db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("配置不存在")
		}
		s.logger.Error("查询配置失败", zap.Error(err))
		return nil, fmt.Errorf("查询配置失败: %w", err)
	}
	return config, nil
}

// Page 分页查询配置列表
func (s *configService) Page(ctx context.Context, pageNum, pageSize int, configCode, name string) (*pagination.Page[model.Config], error) {
	query := s.db.Model(&model.Config{})

	// 条件查询
	if configCode != "" {
		query = query.Where("code = ?", configCode)
	}
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	// 构建 PageQuery
	pageQuery := &pagination.PageQuery{
		PageNum:  pageNum,
		PageSize: pageSize,
	}

	// 使用 Paginator 进行分页
	page, err := pagination.New[model.Config](query, pageQuery).Find()
	if err != nil {
		s.logger.Error("分页查询配置列表失败", zap.Error(err))
		return nil, fmt.Errorf("分页查询配置列表失败: %w", err)
	}

	return page, nil
}

// GetByCode 根据配置编码获取配置列表
func (s *configService) GetByCode(ctx context.Context, configCode string) ([]model.Config, error) {
	configs, err := (&model.Config{}).FindByCode(s.db, configCode)
	if err != nil {
		s.logger.Error("查询配置列表失败", zap.Error(err))
		return nil, fmt.Errorf("查询配置列表失败: %w", err)
	}
	return configs, nil
}

// GetDataByCode 根据配置编码获取配置数据
func (s *configService) GetDataByCode(ctx context.Context, configCode string) (json.RawMessage, error) {
	data, err := (&model.Config{}).GetDataByCode(s.db, configCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("配置不存在")
		}
		s.logger.Error("查询配置数据失败", zap.Error(err))
		return nil, fmt.Errorf("查询配置数据失败: %w", err)
	}
	return data, nil
}
