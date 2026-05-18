package jobs

import (
	"context"

	logging "github.com/gcc798/nai-tizi/internal/logger"
	"github.com/gcc798/nai-tizi/internal/messaging/mqtt/retry"
	"go.uber.org/zap"
)

// MessageRetryJob 定义业务数据结构。
type MessageRetryJob struct {
	retryManager *retry.Manager
	logger       logging.Logger
}

// NewMessageRetryJob 创建组件实例。
func NewMessageRetryJob(retryManager *retry.Manager, logger logging.Logger) *MessageRetryJob {
	return &MessageRetryJob{retryManager: retryManager, logger: logger}
}

// Run 执行业务任务。
func (j *MessageRetryJob) Run() {
	if j.retryManager == nil {
		j.logger.Debug("message retry manager not initialized, skip job")
		return
	}
	ctx := context.Background()
	if err := j.retryManager.ProcessPending(ctx, 0); err != nil {
		j.logger.Error("message retry job failed", zap.Error(err))
	} else {
		j.logger.Debug("message retry job completed")
	}
}

// Schedule 返回任务调度表达式。
func (j *MessageRetryJob) Schedule() string { return "0 */1 * * * *" }
