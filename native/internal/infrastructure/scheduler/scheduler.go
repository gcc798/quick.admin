package scheduler

import (
	"context"
	"fmt"
	"sync"

	logging "github.com/force-c/nai-tizi/internal/logger"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Scheduler 调度器，管理所有定时任务
type Scheduler struct {
	cron    *cron.Cron
	logger  logging.Logger
	ctx     context.Context
	cancel  context.CancelFunc
	jobs    map[string]cron.EntryID // jobName -> entryID映射
	jobsMux sync.RWMutex            // 保护jobs map的并发访问
}

// New 创建新的调度器实例
func New(logger logging.Logger) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	c := cron.New(cron.WithSeconds())
	return &Scheduler{
		cron:   c,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
		jobs:   make(map[string]cron.EntryID),
	}
}

// AddJob 添加定时任务
// spec: cron表达式，如 "0 */5 * * * *" (每5分钟)
// name: 任务名称，用于标识和管理任务
// job: 要执行的任务函数
func (s *Scheduler) AddJob(spec string, name string, job func()) error {
	s.jobsMux.Lock()
	defer s.jobsMux.Unlock()

	// 检查任务是否已存在
	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("job %s already exists", name)
	}

	// 添加任务到cron
	entryID, err := s.cron.AddFunc(spec, func() {
		s.logger.Debug("running scheduled job", zap.String("job", name))
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("job panic", zap.String("job", name), zap.Any("panic", r))
			}
		}()
		job()
	})

	if err != nil {
		s.logger.Error("failed to add job", zap.String("job", name), zap.Error(err))
		return err
	}

	// 记录任务ID
	s.jobs[name] = entryID
	s.logger.Info("job added successfully", zap.String("job", name), zap.String("spec", spec))
	return nil
}

// RemoveJob 移除指定的任务
func (s *Scheduler) RemoveJob(name string) error {
	s.jobsMux.Lock()
	defer s.jobsMux.Unlock()

	entryID, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("job %s not found", name)
	}

	// 从cron中移除任务
	s.cron.Remove(entryID)

	// 从映射中删除
	delete(s.jobs, name)

	s.logger.Info("job removed successfully", zap.String("job", name))
	return nil
}

// UpdateJob 更新任务的cron表达式和执行函数
func (s *Scheduler) UpdateJob(spec string, name string, job func()) error {
	// 先移除旧任务
	if err := s.RemoveJob(name); err != nil {
		// 如果任务不存在，直接添加新任务
		if err.Error() != fmt.Sprintf("job %s not found", name) {
			return err
		}
	}

	// 添加新任务
	return s.AddJob(spec, name, job)
}

// JobExists 检查任务是否存在
func (s *Scheduler) JobExists(name string) bool {
	s.jobsMux.RLock()
	defer s.jobsMux.RUnlock()
	_, exists := s.jobs[name]
	return exists
}

// GetJobCount 获取当前任务数量
func (s *Scheduler) GetJobCount() int {
	s.jobsMux.RLock()
	defer s.jobsMux.RUnlock()
	return len(s.jobs)
}

func (s *Scheduler) Name() string {
	return "Scheduler"
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	s.logger.Info("starting scheduler")
	s.cron.Start()
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() error {
	s.logger.Info("stopping scheduler")
	s.cancel()
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Info("scheduler stopped")
	return nil
}

// GetContext 获取调度器上下文
func (s *Scheduler) GetContext() context.Context {
	return s.ctx
}
