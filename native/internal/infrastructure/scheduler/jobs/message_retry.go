package jobs

import (
	"context"

	"github.com/force-c/nai-tizi/internal/infrastructure/mqtt/retry"
	logging "github.com/force-c/nai-tizi/internal/logger"
	"go.uber.org/zap"
)

type MessageRetryJob struct {
	retryManager *retry.Manager
	logger       logging.Logger
}

func NewMessageRetryJob(retryManager *retry.Manager, logger logging.Logger) *MessageRetryJob {
	return &MessageRetryJob{retryManager: retryManager, logger: logger}
}
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
func (j *MessageRetryJob) Schedule() string { return "0 */1 * * * *" }
