package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "status"},
	)

	DBQueryTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_query_total",
			Help: "Total number of database operations",
		},
		[]string{"operation", "status"},
	)

	RedisOpDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_operation_duration_seconds",
			Help:    "Redis operation latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "status"},
	)

	RedisOpTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_operation_total",
			Help: "Total number of Redis operations",
		},
		[]string{"operation", "status"},
	)

	DBPoolOpenConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_pool_open_connections",
			Help: "Current number of open database connections",
		},
	)

	DBPoolInUseConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_pool_in_use_connections",
			Help: "Current number of in-use database connections",
		},
	)

	DBPoolIdleConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_pool_idle_connections",
			Help: "Current number of idle database connections",
		},
	)

	DBPoolWaitCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_pool_wait_count",
			Help: "Total wait count for a database connection",
		},
	)

	DBPoolWaitDurationSeconds = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_pool_wait_duration_seconds",
			Help: "Total time blocked waiting for a new database connection",
		},
	)
)
