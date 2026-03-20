package model

import (
	"github.com/force-c/nai-tizi/internal/utils"

	"gorm.io/gorm"
)

// 状态常量
const (
	RetryStatusPending   = 1 // PENDING - 待处理
	RetryStatusSuccess   = 2 // SUCCESS - 成功
	RetryStatusFailed    = 3 // FAILED - 失败
	RetryStatusAbandoned = 4 // ABANDONED - 已废弃
)

// 结果常量
const (
	RetryResultSuccess = 1 // SUCCESS - 成功
	RetryResultTimeout = 2 // TIMEOUT - 超时
)

// BuMessageRetry 消息重试主表
type BuMessageRetry struct {
	Id                   int64            `gorm:"column:id;primaryKey" autogen:"int64" json:"id"`                      // 重试记录ID
	MessageId            string           `gorm:"column:message_id;type:varchar(100);not null;index" json:"messageId"` // 消息ID（业务记录ID）
	DeviceId             int64            `gorm:"column:device_id;not null" json:"deviceId"`                           // 设备ID
	DeviceMac            string           `gorm:"column:device_mac;type:varchar(50);not null" json:"deviceMac"`        // 设备MAC地址
	DeviceSnNum          string           `gorm:"column:device_sn_num;type:varchar(50)" json:"deviceSnNum"`            // 设备序列号
	MessageType          int              `gorm:"column:message_type;not null" json:"messageType"`                     // 消息类型（对应OptCodeEnum）
	MessageContent       string           `gorm:"column:message_content;type:text" json:"messageContent"`              // MQTT消息内容
	MaxRetryCount        int              `gorm:"column:max_retry_count;default:3" json:"maxRetryCount"`               // 最大重试次数
	CurrentRetryCount    int              `gorm:"column:current_retry_count;default:0" json:"currentRetryCount"`       // 当前已重试次数
	Status               int              `gorm:"column:status;default:1" json:"status"`                               // 状态：1=PENDING, 2=SUCCESS, 3=FAILED, 4=ABANDONED
	SuccessRetrySequence *int             `gorm:"column:success_retry_sequence" json:"successRetrySequence"`           // 第几次成功（0=首次成功，1=第1次重试成功）
	SuccessTime          *utils.LocalTime `gorm:"column:success_time" json:"successTime"`                              // 最终成功时间
	AbandonReason        string           `gorm:"column:abandon_reason;type:varchar(200)" json:"abandonReason"`        // 废弃原因（如：SERVER_RESTART）
	CreateTime           utils.LocalTime  `gorm:"column:create_time;autoCreateTime" json:"createTime"`                 // 创建时间
	UpdateTime           utils.LocalTime  `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`                 // 更新时间
	DeletedAt            gorm.DeletedAt   `gorm:"column:deleted_at;index" json:"-"`                                    // 删除时间
}

func (*BuMessageRetry) TableName() string {
	return "biz_message_retry"
}

// BuMessageRetryLog 消息重试明细表
type BuMessageRetryLog struct {
	LogId           int64            `gorm:"column:log_id;primaryKey;autoIncrement" json:"logId"`                 // 日志ID
	RetryId         int64            `gorm:"column:retry_id;not null;index" json:"retryId"`                       // 关联主表外键
	MessageId       string           `gorm:"column:message_id;type:varchar(100);not null;index" json:"messageId"` // 消息ID（冗余字段，方便直接查询）
	RetrySequence   int              `gorm:"column:retry_sequence;not null" json:"retrySequence"`                 // 重试序号（0=首次，1=第1次重试...）
	SendTime        utils.LocalTime  `gorm:"column:send_time;not null" json:"sendTime"`                           // 发送时间
	ResponseTime    *utils.LocalTime `gorm:"column:response_time" json:"responseTime"`                            // 响应时间（成功时记录）
	Result          int              `gorm:"column:result;not null" json:"result"`                                // 结果：1=SUCCESS, 2=TIMEOUT
	ResponseContent string           `gorm:"column:response_content;type:text" json:"responseContent"`            // 响应内容（成功时记录）
	CreateTime      utils.LocalTime  `gorm:"column:create_time;autoCreateTime" json:"createTime"`                 // 创建时间
	UpdateTime      utils.LocalTime  `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`                 // 更新时间
	DeletedAt       gorm.DeletedAt   `gorm:"column:deleted_at;index" json:"-"`                                    // 删除时间
}

func (*BuMessageRetryLog) TableName() string {
	return "biz_message_retry_log"
}

// ============ BuMessageRetry 主表方法 ============

// Create 创建重试记录
func (*BuMessageRetry) Create(db *gorm.DB, retry *BuMessageRetry) error {
	return db.Create(retry).Error
}

// SelectByMessageId 根据消息ID查询重试记录
func (*BuMessageRetry) SelectByMessageId(db *gorm.DB, messageId string) (*BuMessageRetry, error) {
	var retry BuMessageRetry
	err := db.Where("message_id = ?", messageId).First(&retry).Error
	if err != nil {
		return nil, err
	}
	return &retry, nil
}

// UpdateRetryCount 更新重试次数
func (*BuMessageRetry) UpdateRetryCount(db *gorm.DB, messageId string, newRetryCount int) error {
	return db.Model(&BuMessageRetry{}).
		Where("message_id = ?", messageId).
		Update("current_retry_count", newRetryCount).Error
}

// UpdateStatus 更新状态
func (*BuMessageRetry) UpdateStatus(db *gorm.DB, messageId string, status int) error {
	return db.Model(&BuMessageRetry{}).
		Where("message_id = ?", messageId).
		Update("status", status).Error
}

// MarkSuccessWithTime 标记成功（使用指定时间）
func (*BuMessageRetry) MarkSuccessWithTime(db *gorm.DB, messageId string, status int, retrySequence int, successTime utils.LocalTime) error {
	updates := map[string]interface{}{
		"status":                 status,
		"success_retry_sequence": retrySequence,
		"success_time":           successTime,
		"update_time":            successTime,
	}
	return db.Model(&BuMessageRetry{}).
		Where("message_id = ?", messageId).
		Updates(updates).Error
}

// MarkAbandoned 批量标记为废弃
func (*BuMessageRetry) MarkAbandoned(db *gorm.DB, messageIds []string, status int, reason string) error {
	return db.Model(&BuMessageRetry{}).
		Where("message_id IN ?", messageIds).
		Updates(map[string]interface{}{
			"status":         status,
			"abandon_reason": reason,
		}).Error
}

// SelectAbandonedTasks 查询超时的PENDING任务
func (*BuMessageRetry) SelectAbandonedTasks(db *gorm.DB, status int, timeoutThreshold int64) ([]BuMessageRetry, error) {
	var retries []BuMessageRetry
	// PostgreSQL 使用 EXTRACT(EPOCH FROM create_time) 替代 MySQL 的 UNIX_TIMESTAMP()
	err := db.Where("status = ? AND EXTRACT(EPOCH FROM create_time) * 1000 < ?", status, timeoutThreshold).
		Find(&retries).Error
	return retries, err
}

// ============ BuMessageRetryLog 明细表方法 ============

// InsertRetryLog 插入重试日志
func (*BuMessageRetryLog) InsertRetryLog(db *gorm.DB, retryId int64, messageId string, retrySequence int, result int) error {
	log := &BuMessageRetryLog{
		RetryId:       retryId,
		MessageId:     messageId,
		RetrySequence: retrySequence,
		SendTime:      utils.Now(),
		Result:        result,
	}
	return db.Create(log).Error
}

// InsertRetryLogWithTime 插入重试日志（使用传入的发送时间）
func (*BuMessageRetryLog) InsertRetryLogWithTime(db *gorm.DB, retryId int64, messageId string, retrySequence int, sendTime utils.LocalTime, result int) error {
	log := &BuMessageRetryLog{
		RetryId:       retryId,
		MessageId:     messageId,
		RetrySequence: retrySequence,
		SendTime:      sendTime,
		Result:        result,
	}
	return db.Create(log).Error
}

// UpdateResponse 更新响应信息
func (*BuMessageRetryLog) UpdateResponse(db *gorm.DB, messageId string, retrySequence int, result int, responseContent string) error {
	now := utils.Now()
	return db.Model(&BuMessageRetryLog{}).
		Where("message_id = ? AND retry_sequence = ?", messageId, retrySequence).
		Updates(map[string]interface{}{
			"response_time":    now,
			"result":           result,
			"response_content": responseContent,
		}).Error
}

// UpdateResponseWithTime 更新响应信息（使用传入的响应时间）
func (*BuMessageRetryLog) UpdateResponseWithTime(db *gorm.DB, messageId string, retrySequence int, result int, responseContent string, responseTime utils.LocalTime) error {
	return db.Model(&BuMessageRetryLog{}).
		Where("message_id = ? AND retry_sequence = ?", messageId, retrySequence).
		Updates(map[string]interface{}{
			"response_time":    responseTime,
			"result":           result,
			"response_content": responseContent,
			"update_time":      responseTime,
		}).Error
}
