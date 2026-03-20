package metrics

import (
	"database/sql"
	"time"
)

func ObserveDB(operation, status string, duration time.Duration) {
	DBQueryTotal.WithLabelValues(operation, status).Inc()
	DBQueryDuration.WithLabelValues(operation, status).Observe(duration.Seconds())
}

func ObserveRedis(operation, status string, duration time.Duration) {
	RedisOpTotal.WithLabelValues(operation, status).Inc()
	RedisOpDuration.WithLabelValues(operation, status).Observe(duration.Seconds())
}

func SetDBPoolStats(stats sql.DBStats) {
	DBPoolOpenConnections.Set(float64(stats.OpenConnections))
	DBPoolInUseConnections.Set(float64(stats.InUse))
	DBPoolIdleConnections.Set(float64(stats.Idle))
	DBPoolWaitCount.Set(float64(stats.WaitCount))
	DBPoolWaitDurationSeconds.Set(stats.WaitDuration.Seconds())
}
