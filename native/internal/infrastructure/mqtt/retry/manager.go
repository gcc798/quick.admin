package retry

import (
	"context"
	"fmt"
	"time"

	"github.com/force-c/nai-tizi/internal/constants"
	"github.com/force-c/nai-tizi/internal/domain/model"
	"github.com/force-c/nai-tizi/internal/infrastructure/mqtt"
	"github.com/force-c/nai-tizi/internal/infrastructure/redis"
	logging "github.com/force-c/nai-tizi/internal/logger"
	"github.com/force-c/nai-tizi/internal/utils"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	defaultMaxRetryCount = 2
	defaultRetryInterval = 1000 // 毫秒
	defaultMaxBatchSize  = 100
)

// Manager 消息重试管理器
type Manager struct {
	db         *gorm.DB
	mqttClient *mqtt.Client
	logger     logging.Logger
	redisUtils *redis.RedisUtils
	config     *RetryConfig
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// RetryConfig 重试配置
type RetryConfig struct {
	Enabled            bool  // 是否启用重试机制
	MaxRetryCount      int   // 默认最大重试次数
	RetryInterval      int64 // 默认重试间隔（毫秒）
	ScanInterval       int64 // 扫描超时任务的间隔（毫秒）
	MaxBatchSize       int   // 每次扫描处理的最大任务数
	LockWaitTime       int64 // 分布式锁等待时间（毫秒）
	LockLeaseTime      int64 // 分布式锁过期时间（毫秒）
	RedisExpireMinutes int   // Redis数据过期时间（分钟）
	AbandonTimeout     int64 // 废弃任务超时时间（毫秒）
	CleanupInterval    int64 // 清理废弃任务的间隔（毫秒）
}

// NewManager 创建消息重试管理器
func NewManager(db *gorm.DB, mqttClient *mqtt.Client, redisClient *goredis.Client, logger logging.Logger, config *RetryConfig) *Manager {
	if config == nil {
		config = &RetryConfig{
			Enabled:            true,
			MaxRetryCount:      defaultMaxRetryCount,
			RetryInterval:      defaultRetryInterval,
			ScanInterval:       1000,
			MaxBatchSize:       defaultMaxBatchSize,
			LockWaitTime:       100,
			LockLeaseTime:      2000,
			RedisExpireMinutes: 10,
			AbandonTimeout:     120000,
			CleanupInterval:    300000,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		db:         db,
		mqttClient: mqttClient,
		logger:     logger,
		redisUtils: redis.NewRedisUtils(redisClient),
		config:     config,
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

func (m *Manager) Name() string {
	return "Retry Manager"
}

// Start 启动定时任务
func (m *Manager) Start() error {
	if !m.config.Enabled {
		m.logger.Info("重试机制已禁用")
		return nil
	}

	// 启动扫描超时任务
	go m.scanTimeoutTasksLoop()

	// 启动清理废弃任务
	go m.cleanupAbandonedTasksLoop()

	m.logger.Info("消息重试管理器已启动")
	return nil
}

// Stop 停止定时任务
func (m *Manager) Stop() error {
	if m.cancelFunc != nil {
		m.cancelFunc()
	}
	m.logger.Info("消息重试管理器已停止")
	return nil
}

// SubmitRetryMessage 提交重试消息
func (m *Manager) SubmitRetryMessage(request *RetryMessageRequest) (int64, error) {
	if !m.config.Enabled {
		m.logger.Debug("重试机制已禁用，跳过提交重试任务")
		return 0, nil
	}

	messageId := request.MessageId
	retryInterval := m.config.RetryInterval
	if request.RetryInterval != nil {
		retryInterval = *request.RetryInterval
	}
	executeTime := time.Now().UnixMilli() + retryInterval

	// 1. 保存到数据库（持久化）
	retryRecord := m.buildRetryRecord(request)
	if err := (&model.BuMessageRetry{}).Create(m.db, retryRecord); err != nil {
		m.logger.Error("保存重试记录失败", zap.String("messageId", messageId), zap.Error(err))
		return 0, fmt.Errorf("保存重试记录失败: %w", err)
	}

	// 2. 保存请求数据到Redis Hash（临时存储）
	hashKey := constants.RetryDataPrefix + messageId
	retryData := m.buildRetryHashData(request)
	if err := m.redisUtils.SetCacheMap(hashKey, retryData); err != nil {
		m.logger.Error("保存Redis数据失败", zap.String("messageId", messageId), zap.Error(err))
		return 0, fmt.Errorf("保存Redis数据失败: %w", err)
	}
	if err := m.redisUtils.Expire(hashKey, time.Duration(m.config.RedisExpireMinutes)*time.Minute); err != nil {
		m.logger.Warn("设置Redis过期时间失败", zap.String("messageId", messageId), zap.Error(err))
	}

	// 3. 添加到延时队列（ZSET，score为执行时间）
	if err := m.redisUtils.ZAdd(constants.RetryDelayQueueKey, float64(executeTime), messageId); err != nil {
		m.logger.Error("添加到延时队列失败", zap.String("messageId", messageId), zap.Error(err))
		return 0, fmt.Errorf("添加到延时队列失败: %w", err)
	}

	// 4. 立即发送第一次消息并记录日志（记录MQTT发送的精确时间）
	sendTime := utils.Now()
	if err := m.sendMqttMessage(request); err != nil {
		m.logger.Error("发送MQTT消息失败", zap.String("messageId", messageId), zap.Error(err))
	}
	if err := (&model.BuMessageRetryLog{}).InsertRetryLogWithTime(m.db, retryRecord.Id, messageId, 0, sendTime, model.RetryResultTimeout); err != nil {
		m.logger.Error("插入重试日志失败", zap.String("messageId", messageId), zap.Error(err))
	}

	m.logger.Info("提交重试任务成功",
		zap.String("messageId", messageId),
		zap.String("deviceMac", request.DeviceMac),
		zap.Int64("executeTime", executeTime))

	return retryRecord.Id, nil
}

// MarkSuccess 标记消息成功（收到设备响应时调用）
func (m *Manager) MarkSuccess(messageId string, payload string) {
	// 1. 查询重试记录
	retryRecord, err := (&model.BuMessageRetry{}).SelectByMessageId(m.db, messageId)
	if err != nil {
		m.logger.Error("重试记录不存在", zap.String("messageId", messageId), zap.Error(err))
		return
	}

	if retryRecord.Status != model.RetryStatusPending {
		m.logger.Error("重试记录状态非PENDING",
			zap.String("messageId", messageId),
			zap.Int("status", retryRecord.Status))
		return
	}

	// 2. 获取统一的当前时间（支持毫秒精度）
	currentTime := utils.Now()

	// 3. 使用统一时间更新主表状态为成功
	if err := (&model.BuMessageRetry{}).MarkSuccessWithTime(m.db, messageId, model.RetryStatusSuccess, retryRecord.CurrentRetryCount, currentTime); err != nil {
		m.logger.Error("更新重试记录状态失败", zap.String("messageId", messageId), zap.Error(err))
		return
	}

	// 4. 使用相同时间更新明细表响应信息
	if err := (&model.BuMessageRetryLog{}).UpdateResponseWithTime(m.db, messageId, retryRecord.CurrentRetryCount, model.RetryResultSuccess, payload, currentTime); err != nil {
		m.logger.Error("更新重试日志失败", zap.String("messageId", messageId), zap.Error(err))
	}

	// 5. 从Redis延时队列中移除
	if _, err := m.redisUtils.ZRem(constants.RetryDelayQueueKey, messageId); err != nil {
		m.logger.Error("从延时队列移除失败", zap.String("messageId", messageId), zap.Error(err))
	}

	// 6. 清理Redis临时数据
	m.cleanupRedisData(messageId)

	m.logger.Info("重试任务成功完成",
		zap.String("messageId", messageId),
		zap.Int("retryCount", retryRecord.CurrentRetryCount))
}

// scanTimeoutTasksLoop 定时扫描超时任务
func (m *Manager) scanTimeoutTasksLoop() {
	ticker := time.NewTicker(time.Duration(m.config.ScanInterval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.scanTimeoutTasks()
		}
	}
}

// scanTimeoutTasks 扫描超时任务
func (m *Manager) scanTimeoutTasks() {
	// 只有一个实例能获取到锁，避免重复执行
	locked, err := m.redisUtils.TryLock(constants.RetryScanLockKey, m.config.LockWaitTime, m.config.LockLeaseTime)
	if err != nil {
		m.logger.Error("获取扫描锁失败", zap.Error(err))
		return
	}
	if !locked {
		return
	}

	// 快速获取到期任务
	expiredTasks, err := m.getExpiredTasksFromRedis()
	if err != nil {
		m.logger.Error("获取到期任务失败", zap.Error(err))
		m.redisUtils.Unlock(constants.RetryScanLockKey)
		return
	}

	// 立即释放锁
	m.redisUtils.Unlock(constants.RetryScanLockKey)

	// 异步处理任务
	if len(expiredTasks) > 0 {
		go m.processExpiredTasks(expiredTasks)
	}
}

// getExpiredTasksFromRedis 从Redis获取到期任务
func (m *Manager) getExpiredTasksFromRedis() ([]string, error) {
	currentTime := float64(time.Now().UnixMilli())
	return m.redisUtils.ZRangeByScore(constants.RetryDelayQueueKey, 0, currentTime, 0, int64(m.config.MaxBatchSize))
}

// processExpiredTasks 异步处理到期任务
func (m *Manager) processExpiredTasks(expiredTasks []string) {
	m.logger.Debug("异步处理超时重试任务", zap.Int("count", len(expiredTasks)))

	for _, messageId := range expiredTasks {
		// 从ZSET中移除（原子操作）
		removed, err := m.redisUtils.ZRem(constants.RetryDelayQueueKey, messageId)
		if err != nil {
			m.logger.Error("从延时队列移除失败", zap.String("messageId", messageId), zap.Error(err))
			continue
		}
		if removed {
			// 成功移除，处理该任务
			m.processSingleRetryTask(messageId)
		}
	}
}

// processSingleRetryTask 处理单个重试任务
func (m *Manager) processSingleRetryTask(messageId string) {
	// 1. 从数据库获取最新状态（确保数据一致性）
	dbRecord, err := (&model.BuMessageRetry{}).SelectByMessageId(m.db, messageId)
	if err != nil || dbRecord.Status != model.RetryStatusPending {
		// 任务已完成或失败，清理Redis数据
		m.cleanupRedisData(messageId)
		return
	}

	// 2. 检查重试次数
	if dbRecord.CurrentRetryCount >= dbRecord.MaxRetryCount {
		// 达到最大重试次数，标记失败
		if err := (&model.BuMessageRetry{}).UpdateStatus(m.db, messageId, model.RetryStatusFailed); err != nil {
			m.logger.Error("更新重试状态失败", zap.String("messageId", messageId), zap.Error(err))
		}
		m.logger.Error("模式控制最终失败",
			zap.String("messageId", messageId),
			zap.String("deviceMac", dbRecord.DeviceMac),
			zap.Int("retryCount", dbRecord.CurrentRetryCount),
			zap.Int("maxRetryCount", dbRecord.MaxRetryCount))
		m.cleanupRedisData(messageId)
		return
	}

	// 3. 从Redis获取请求数据
	hashKey := constants.RetryDataPrefix + messageId
	retryData, err := m.redisUtils.GetCacheMap(hashKey)
	if err != nil || len(retryData) == 0 {
		m.logger.Warn("Redis中找不到重试数据，跳过处理", zap.String("messageId", messageId))
		m.cleanupRedisData(messageId)
		return
	}

	// 4. 更新重试次数
	newRetryCount := dbRecord.CurrentRetryCount + 1
	if err := (&model.BuMessageRetry{}).UpdateRetryCount(m.db, messageId, newRetryCount); err != nil {
		m.logger.Error("更新重试次数失败", zap.String("messageId", messageId), zap.Error(err))
		return
	}

	// 5. 发送重试消息并记录日志（记录MQTT发送的精确时间）
	request, err := FromRedisCache(messageId, retryData)
	if err != nil {
		m.logger.Error("解析Redis数据失败", zap.String("messageId", messageId), zap.Error(err))
		return
	}

	sendTime := utils.Now()
	if err := m.sendMqttMessage(request); err != nil {
		m.logger.Error("发送MQTT消息失败", zap.String("messageId", messageId), zap.Error(err))
	}
	if err := (&model.BuMessageRetryLog{}).InsertRetryLogWithTime(m.db, dbRecord.Id, messageId, newRetryCount, sendTime, model.RetryResultTimeout); err != nil {
		m.logger.Error("插入重试日志失败", zap.String("messageId", messageId), zap.Error(err))
	}

	// 6. 安排下次重试（重新加入ZSET）
	retryInterval := m.config.RetryInterval
	if request.RetryInterval != nil {
		retryInterval = *request.RetryInterval
	}
	nextExecuteTime := time.Now().UnixMilli() + retryInterval
	if err := m.redisUtils.ZAdd(constants.RetryDelayQueueKey, float64(nextExecuteTime), messageId); err != nil {
		m.logger.Error("重新加入延时队列失败", zap.String("messageId", messageId), zap.Error(err))
	}

	m.logger.Info("执行重试成功",
		zap.String("messageId", messageId),
		zap.Int("retryCount", newRetryCount),
		zap.Int("maxRetryCount", dbRecord.MaxRetryCount))
}

// cleanupAbandonedTasksLoop 定时清理废弃任务
func (m *Manager) cleanupAbandonedTasksLoop() {
	ticker := time.NewTicker(time.Duration(m.config.CleanupInterval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.cleanupAbandonedTasks()
		}
	}
}

// cleanupAbandonedTasks 清理废弃任务
func (m *Manager) cleanupAbandonedTasks() {
	// 计算超时阈值
	timeoutThreshold := time.Now().UnixMilli() - m.config.AbandonTimeout

	// 查找超时的PENDING任务
	abandonedTasks, err := (&model.BuMessageRetry{}).SelectAbandonedTasks(m.db, model.RetryStatusPending, timeoutThreshold)
	if err != nil {
		m.logger.Error("查询废弃任务失败", zap.Error(err))
		return
	}

	if len(abandonedTasks) == 0 {
		return
	}

	// 提取messageId列表
	messageIds := make([]string, 0, len(abandonedTasks))
	for _, task := range abandonedTasks {
		messageIds = append(messageIds, task.MessageId)
	}

	// 批量标记为废弃状态
	if err := (&model.BuMessageRetry{}).MarkAbandoned(m.db, messageIds, model.RetryStatusAbandoned, constants.AbandonReasonTimeoutAbandoned); err != nil {
		m.logger.Error("标记废弃任务失败", zap.Error(err))
		return
	}

	// 清理Redis数据
	for _, messageId := range messageIds {
		m.cleanupRedisData(messageId)
	}

	m.logger.Warn("清理废弃重试任务",
		zap.Int("count", len(abandonedTasks)),
		zap.Int64("timeoutThreshold", m.config.AbandonTimeout))
}

// ProcessPending 供调度任务触发一次扫描
func (m *Manager) ProcessPending(ctx context.Context, _ int) error {
	if m == nil {
		return nil
	}
	m.scanTimeoutTasks()
	m.cleanupAbandonedTasks()
	return nil
}

// buildRetryRecord 构建重试记录
func (m *Manager) buildRetryRecord(request *RetryMessageRequest) *model.BuMessageRetry {
	maxRetryCount := m.config.MaxRetryCount
	if request.MaxRetryCount != nil {
		maxRetryCount = *request.MaxRetryCount
	}

	return &model.BuMessageRetry{
		MessageId:         request.MessageId,
		DeviceId:          request.DeviceId,
		DeviceMac:         request.DeviceMac,
		DeviceSnNum:       request.DeviceSnNum,
		MessageType:       request.MessageType,
		MessageContent:    request.MessageContent,
		MaxRetryCount:     maxRetryCount,
		CurrentRetryCount: 0,
		Status:            model.RetryStatusPending,
	}
}

// buildRetryHashData 构建Redis Hash数据
func (m *Manager) buildRetryHashData(request *RetryMessageRequest) map[string]interface{} {
	maxRetryCount := m.config.MaxRetryCount
	if request.MaxRetryCount != nil {
		maxRetryCount = *request.MaxRetryCount
	}

	retryInterval := m.config.RetryInterval
	if request.RetryInterval != nil {
		retryInterval = *request.RetryInterval
	}

	return map[string]interface{}{
		"deviceId":       fmt.Sprintf("%d", request.DeviceId),
		"deviceMac":      request.DeviceMac,
		"deviceSnNum":    request.DeviceSnNum,
		"messageType":    fmt.Sprintf("%d", request.MessageType),
		"messageContent": request.MessageContent,
		"maxRetryCount":  maxRetryCount,
		"retryInterval":  retryInterval,
	}
}

// sendMqttMessage 发送MQTT消息
func (m *Manager) sendMqttMessage(request *RetryMessageRequest) error {
	// 构建topic: NTZ/{netType}/{mac}/{sn}
	// 这里假设netType为"wifi"，如果需要动态获取，需要从request中传入
	topic := fmt.Sprintf("NTZ/wifi/%s/%s", request.DeviceMac, request.DeviceSnNum)
	return m.mqttClient.Publish(topic, request.MessageContent)
}

// cleanupRedisData 清理Redis数据
func (m *Manager) cleanupRedisData(messageId string) {
	hashKey := constants.RetryDataPrefix + messageId
	if err := m.redisUtils.DeleteObject(hashKey); err != nil {
		m.logger.Error("清理Redis数据失败", zap.String("messageId", messageId), zap.Error(err))
	}
}
