package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HttpRequestsTotal HTTP 请求总数
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HttpRequestDuration HTTP 请求延迟
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// DbQueryDuration 数据库查询延迟
	DbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// RedisOpDuration Redis 操作延迟
	RedisOpDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_operation_duration_seconds",
			Help:    "Redis operation latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// ActiveConnections 活跃连接数
	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active connections",
		},
	)

	// DbConnectionPoolSize 数据库连接池大小
	DbConnectionPoolSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connection_pool_size",
			Help: "Current database connection pool size",
		},
	)

	// DbConnectionPoolIdle 数据库空闲连接数
	DbConnectionPoolIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connection_pool_idle",
			Help: "Number of idle database connections",
		},
	)

	// DbConnectionPoolInUse 数据库使用中连接数
	DbConnectionPoolInUse = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connection_pool_in_use",
			Help: "Number of database connections in use",
		},
	)
)
