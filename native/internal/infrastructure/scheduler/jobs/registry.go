package jobs

import (
	"fmt"

	"github.com/force-c/nai-tizi/internal/infrastructure/mqtt/retry"
	"github.com/force-c/nai-tizi/internal/infrastructure/scheduler"
	logging "github.com/force-c/nai-tizi/internal/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RegisterJobs 注册所有定时任务
func RegisterJobs(
	sched *scheduler.Scheduler,
	db *gorm.DB,
	redis *redis.Client,
	retryManager *retry.Manager,
	logger logging.Logger,
) error {
	// 1. 数据清理任务
	cl := NewDataCleanupJob(db, redis, logger)
	if err := sched.AddJob(cl.Schedule(), "data-cleanup", cl.Run); err != nil {
		return fmt.Errorf("failed to add data-cleanup job: %w", err)
	}

	// 2. MQTT消息重试任务
	if retryManager != nil {
		mr := NewMessageRetryJob(retryManager, logger)
		if err := sched.AddJob(mr.Schedule(), "message-retry", mr.Run); err != nil {
			return fmt.Errorf("failed to add message-retry job: %w", err)
		}
	}

	logger.Info("all jobs registered successfully", zap.Int("count", sched.GetJobCount()))
	return nil
}
