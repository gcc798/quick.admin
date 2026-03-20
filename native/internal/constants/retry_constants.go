package constants

// Redis键前缀
const (
	RetryDelayQueueKey = "retry:delay_queue" // 延时队列
	RetryDataPrefix    = "retry:data:"       // 重试数据前缀
	RetryScanLockKey   = "retry:scan_lock"   // 扫描锁
)

// 废弃原因
const (
	AbandonReasonServerRestart    = "SERVER_RESTART"    // 服务重启
	AbandonReasonTimeoutAbandoned = "TIMEOUT_ABANDONED" // 超时废弃
)
