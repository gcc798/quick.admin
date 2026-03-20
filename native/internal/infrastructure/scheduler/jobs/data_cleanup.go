package jobs

import (
	"time"

	logging "github.com/force-c/nai-tizi/internal/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DataCleanupJob struct {
	db     *gorm.DB
	redis  *redis.Client
	logger logging.Logger
}

func NewDataCleanupJob(db *gorm.DB, redis *redis.Client, logger logging.Logger) *DataCleanupJob {
	return &DataCleanupJob{db: db, redis: redis, logger: logger}
}
func (j *DataCleanupJob) Run() {
	j.logger.Info("starting data cleanup")

	// 清理系统登录日志（保留90天）
	cutoffDate90 := time.Now().AddDate(0, 0, -90)
	// 使用Raw SQL或模型进行删除，这里假设sys_logininfor表存在
	// 由于没有导入model包，这里使用Table方法
	result := j.db.Table("sys_logininfor").Where("login_time < ?", cutoffDate90).Delete(nil)
	if result.Error != nil {
		j.logger.Error("failed to cleanup login logs", zap.Error(result.Error))
	} else {
		j.logger.Info("cleaned up login logs", zap.Int64("rows", result.RowsAffected))
	}

	j.logger.Info("data cleanup completed")
}
func (j *DataCleanupJob) Schedule() string { return "0 0 2 * * *" }
