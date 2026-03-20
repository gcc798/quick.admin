package database

import (
	"context"
	"fmt"
	"time"

	"github.com/force-c/nai-tizi/internal/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SlowQueryPlugin 慢 SQL 监控插件
type SlowQueryPlugin struct {
	logger    logger.Logger
	threshold time.Duration // 慢 SQL 阈值
}

// NewSlowQueryPlugin 创建慢 SQL 监控插件
func NewSlowQueryPlugin(log logger.Logger, threshold time.Duration) *SlowQueryPlugin {
	if threshold <= 0 {
		//threshold = 500 * time.Millisecond
		threshold = 10 * time.Millisecond
	}
	return &SlowQueryPlugin{
		logger:    log,
		threshold: threshold,
	}
}

// Name 插件名称
func (p *SlowQueryPlugin) Name() string {
	return "slow_query_plugin"
}

// Initialize 初始化插件
func (p *SlowQueryPlugin) Initialize(db *gorm.DB) error {
	// 为查询操作注册回调
	if err := db.Callback().Query().Before("gorm:query").Register("slow_query:before_query", p.before); err != nil {
		return fmt.Errorf("failed to register before query callback: %w", err)
	}
	if err := db.Callback().Query().After("gorm:after_query").Register("slow_query:after_query", p.after); err != nil {
		return fmt.Errorf("failed to register after query callback: %w", err)
	}

	// 为创建操作注册回调
	if err := db.Callback().Create().Before("gorm:create").Register("slow_query:before_create", p.before); err != nil {
		return fmt.Errorf("failed to register before create callback: %w", err)
	}
	if err := db.Callback().Create().After("gorm:after_create").Register("slow_query:after_create", p.after); err != nil {
		return fmt.Errorf("failed to register after create callback: %w", err)
	}

	// 为更新操作注册回调
	if err := db.Callback().Update().Before("gorm:update").Register("slow_query:before_update", p.before); err != nil {
		return fmt.Errorf("failed to register before update callback: %w", err)
	}
	if err := db.Callback().Update().After("gorm:after_update").Register("slow_query:after_update", p.after); err != nil {
		return fmt.Errorf("failed to register after update callback: %w", err)
	}

	// 为删除操作注册回调
	if err := db.Callback().Delete().Before("gorm:delete").Register("slow_query:before_delete", p.before); err != nil {
		return fmt.Errorf("failed to register before delete callback: %w", err)
	}
	if err := db.Callback().Delete().After("gorm:after_delete").Register("slow_query:after_delete", p.after); err != nil {
		return fmt.Errorf("failed to register after delete callback: %w", err)
	}

	// 为 Row 操作注册回调
	if err := db.Callback().Row().Before("gorm:row").Register("slow_query:before_row", p.before); err != nil {
		return fmt.Errorf("failed to register before row callback: %w", err)
	}
	if err := db.Callback().Row().After("gorm:row").Register("slow_query:after_row", p.after); err != nil {
		return fmt.Errorf("failed to register after row callback: %w", err)
	}

	// 为 Raw 操作注册回调
	if err := db.Callback().Raw().Before("gorm:raw").Register("slow_query:before_raw", p.before); err != nil {
		return fmt.Errorf("failed to register before raw callback: %w", err)
	}
	if err := db.Callback().Raw().After("gorm:raw").Register("slow_query:after_raw", p.after); err != nil {
		return fmt.Errorf("failed to register after raw callback: %w", err)
	}

	p.logger.Info("slow query plugin initialized", zap.Duration("threshold", p.threshold))
	return nil
}

// before 执行前回调
func (p *SlowQueryPlugin) before(db *gorm.DB) {
	db.InstanceSet("slow_query:start_time", time.Now())
}

// after 执行后回调
func (p *SlowQueryPlugin) after(db *gorm.DB) {
	// 获取开始时间
	startTime, ok := db.InstanceGet("slow_query:start_time")
	if !ok {
		return
	}

	// 计算执行时间
	duration := time.Since(startTime.(time.Time))

	// 如果超过阈值，记录慢 SQL
	if duration > p.threshold {
		p.logSlowQuery(db, duration)
	}
}

// logSlowQuery 记录慢 SQL
func (p *SlowQueryPlugin) logSlowQuery(db *gorm.DB, duration time.Duration) {
	ctx := db.Statement.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// 获取 SQL 语句
	sql := db.Statement.SQL.String()
	if sql == "" && db.Statement.SQL.Len() == 0 {
		sql = db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
	}

	// 获取影响行数
	rowsAffected := db.Statement.RowsAffected

	// 获取错误信息
	var errMsg string
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		errMsg = db.Error.Error()
	}

	// 构建日志字段
	fields := []zap.Field{
		zap.String("sql", sql),
		zap.Duration("duration", duration),
		zap.Int64("rows_affected", rowsAffected),
		zap.String("table", db.Statement.Table),
	}

	// 添加 SQL 参数（如果有）
	if len(db.Statement.Vars) > 0 {
		fields = append(fields, zap.Any("vars", db.Statement.Vars))
	}

	// 添加错误信息（如果有）
	if errMsg != "" {
		fields = append(fields, zap.String("error", errMsg))
	}

	// 记录日志
	p.logger.Warn("slow query detected", fields...)

	// 可选：发送到监控系统
	// 例如：Prometheus、Grafana、Sentry 等
	// metrics.SlowQueryCounter.Inc()
	// metrics.SlowQueryDuration.Observe(duration.Seconds())
}
