package jobs

import (
	"context"
	"time"

	"github.com/force-c/nai-tizi/internal/domain/model"
	logging "github.com/force-c/nai-tizi/internal/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DataCleanupJob 定义业务数据结构。
type DataCleanupJob struct {
	db     *gorm.DB
	redis  *redis.Client
	logger logging.Logger
}

// NewDataCleanupJob 创建组件实例。
func NewDataCleanupJob(db *gorm.DB, redis *redis.Client, logger logging.Logger) *DataCleanupJob {
	return &DataCleanupJob{db: db, redis: redis, logger: logger}
}

// Run 执行业务任务。
func (j *DataCleanupJob) Run() {
	j.logger.Info("starting data cleanup")

	cutoffDate := time.Now().AddDate(0, 0, -90)
	if result := j.db.WithContext(context.Background()).
		Where("login_time < ?", cutoffDate).
		Delete(&model.LoginLog{}); result.Error != nil {
		j.logger.Error("failed to cleanup login logs", zap.Error(result.Error))
	} else {
		j.logger.Info("cleaned up login logs", zap.Int64("rows", result.RowsAffected))
	}

	if result := j.db.WithContext(context.Background()).
		Where("oper_time < ?", cutoffDate).
		Delete(&model.OperLog{}); result.Error != nil {
		j.logger.Error("failed to cleanup oper logs", zap.Error(result.Error))
	} else {
		j.logger.Info("cleaned up oper logs", zap.Int64("rows", result.RowsAffected))
	}

	j.logger.Info("data cleanup completed")
}

// Schedule 返回任务调度表达式。
func (j *DataCleanupJob) Schedule() string { return "0 0 2 * * *" }
