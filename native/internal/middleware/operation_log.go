package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/force-c/nai-tizi/internal/domain/model"
	logging "github.com/force-c/nai-tizi/internal/logger"
	"github.com/force-c/nai-tizi/internal/utils"
	"github.com/force-c/nai-tizi/internal/utils/idgen"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 操作日志上下文键
const (
	OperLogTitleKey        = "oper_log_title"
	OperLogBusinessTypeKey = "oper_log_business_type"
)

// 操作日志批量写入配置
const (
	operLogBatchSize     = 100              // 批量写入大小
	operLogFlushInterval = 10 * time.Second // 定时刷新间隔
	operLogChannelSize   = 1000             // 通道缓冲大小
)

// OperLogWriter 操作日志写入器
type OperLogWriter struct {
	db      *gorm.DB
	logger  logging.Logger
	logChan chan *model.OperLog
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

var (
	operLogWriter     *OperLogWriter
	operLogWriterOnce sync.Once
)

// getOperLogWriter 获取操作日志写入器单例
func getOperLogWriter(db *gorm.DB, logger logging.Logger) *OperLogWriter {
	operLogWriterOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		operLogWriter = &OperLogWriter{
			db:      db,
			logger:  logger,
			logChan: make(chan *model.OperLog, operLogChannelSize),
			ctx:     ctx,
			cancel:  cancel,
		}
		operLogWriter.start()
	})
	return operLogWriter
}

// start 启动日志写入协程
func (w *OperLogWriter) start() {
	w.wg.Add(1)
	go w.batchWrite()
}

// batchWrite 批量写入日志
func (w *OperLogWriter) batchWrite() {
	defer w.wg.Done()

	buffer := make([]*model.OperLog, 0, operLogBatchSize)
	ticker := time.NewTicker(operLogFlushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(buffer) == 0 {
			return
		}

		// 批量插入
		if err := w.db.Create(&buffer).Error; err != nil {
			w.logger.Error("批量写入操作日志失败",
				zap.Error(err),
				zap.Int("count", len(buffer)))
		} else {
			w.logger.Debug("批量写入操作日志成功",
				zap.Int("count", len(buffer)))
		}

		// 清空缓冲区
		buffer = buffer[:0]
	}

	for {
		select {
		case <-w.ctx.Done():
			// 程序退出前刷新剩余日志
			flush()
			w.logger.Info("操作日志写入器已停止")
			return

		case log := <-w.logChan:
			buffer = append(buffer, log)
			// 达到批量大小，立即刷新
			if len(buffer) >= operLogBatchSize {
				flush()
			}

		case <-ticker.C:
			// 定时刷新
			flush()
		}
	}
}

// Write 写入日志到通道
func (w *OperLogWriter) Write(log *model.OperLog) {
	select {
	case w.logChan <- log:
		// 成功写入通道
	default:
		// 通道已满，记录警告
		w.logger.Warn("操作日志通道已满，丢弃日志",
			zap.String("title", log.Title),
			zap.String("operName", log.OperName))
	}
}

// Stop 停止日志写入器
func (w *OperLogWriter) Stop() {
	w.cancel()
	w.wg.Wait()
	close(w.logChan)
}

// OperationLog 操作日志中间件
func OperationLog(db *gorm.DB, logger logging.Logger) gin.HandlerFunc {
	// 获取日志写入器单例
	writer := getOperLogWriter(db, logger)

	return func(c *gin.Context) {
		// 跳过操作日志相关的请求，避免递归记录
		if shouldSkipLogging(c.Request.URL.Path) {
			c.Next()
			return
		}

		start := utils.Now()
		bodyBytes := readBody(c)

		c.Next()

		duration := time.Since(start.Time()).Milliseconds()
		operParam := extractParams(c, bodyBytes)
		status, errMsg := resolveStatus(c)

		// 获取用户信息：格式为 "用户ID-用户名称"
		userId := getStringValue(c, "userId")
		userName := getStringValue(c, "userName")
		operName := formatOperName(userId, userName)

		// 获取终端类型（从 token 中解析）
		deviceType := getStringValue(c, "deviceType")
		if deviceType == "" {
			deviceType = "unknown"
		}

		// 获取标题和业务类型（优先从上下文获取，否则自动推断）
		title := getOperLogTitle(c)
		businessType := getBusinessType(c)

		logEntry := &model.OperLog{
			ID:            idgen.MustNextID(),
			Title:         title,
			BusinessType:  businessType,
			Method:        c.HandlerName(),
			RequestMethod: c.Request.Method,
			DeviceType:    deviceType,
			OperName:      operName,
			OperUrl:       c.Request.URL.Path,
			OperIp:        utils.GetClientIP(c),
			OperParam:     truncate(operParam, 2000),
			Status:        status,
			ErrorMsg:      truncate(errMsg, 1000),
			OperTime:      start,
			CostTime:      duration,
			UserAgent:     c.Request.UserAgent(),
		}

		// 写入日志到通道（批量写入）
		writer.Write(logEntry)
	}
}

func readBody(c *gin.Context) []byte {
	if c.Request == nil || c.Request.Body == nil {
		return nil
	}
	// 仅对可重复读场景处理
	if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodDelete {
		return nil
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	return body
}

func extractParams(c *gin.Context, body []byte) string {
	// 检查是否为文件上传请求
	contentType := c.ContentType()
	if strings.Contains(contentType, "multipart/form-data") {
		// 对于文件上传，记录表单字段而不是二进制内容
		return extractMultipartParams(c)
	}

	if len(body) > 0 {
		return string(body)
	}
	return c.Request.URL.RawQuery
}

// extractMultipartParams 提取 multipart/form-data 请求的参数摘要
func extractMultipartParams(c *gin.Context) string {
	// 解析 multipart 表单
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		return "文件上传请求（解析失败）"
	}

	var params []string

	// 记录普通表单字段
	if c.Request.MultipartForm != nil && c.Request.MultipartForm.Value != nil {
		for key, values := range c.Request.MultipartForm.Value {
			for _, value := range values {
				params = append(params, fmt.Sprintf("%s=%s", key, value))
			}
		}
	}

	// 记录文件信息（文件名和大小，不包含内容）
	if c.Request.MultipartForm != nil && c.Request.MultipartForm.File != nil {
		for fieldName, files := range c.Request.MultipartForm.File {
			for _, file := range files {
				params = append(params, fmt.Sprintf("%s=[文件: %s, 大小: %d bytes]",
					fieldName, file.Filename, file.Size))
			}
		}
	}

	if len(params) == 0 {
		return "文件上传请求"
	}

	return strings.Join(params, ", ")
}

func resolveStatus(c *gin.Context) (string, string) {
	if len(c.Errors) == 0 {
		return "0", ""
	}
	return "1", c.Errors.String()
}

func getStringValue(c *gin.Context, key string) string {
	if v, ok := c.Get(key); ok {
		return fmt.Sprint(v)
	}
	return ""
}

func truncate(val string, max int) string {
	if len(val) <= max {
		return val
	}
	return val[:max]
}

// getOperLogTitle 获取操作日志标题
// 优先从上下文获取，否则根据路径自动生成
func getOperLogTitle(c *gin.Context) string {
	// 优先从上下文获取自定义标题
	if title := getStringValue(c, OperLogTitleKey); title != "" {
		return title
	}

	// 根据路径自动生成标题
	path := c.Request.URL.Path
	return generateTitleFromPath(path)
}

// getBusinessType 获取业务类型
// 优先从上下文获取，否则根据请求方法和路径自动推断
func getBusinessType(c *gin.Context) string {
	// 优先从上下文获取自定义业务类型
	if businessType := getStringValue(c, OperLogBusinessTypeKey); businessType != "" {
		return businessType
	}

	// 根据请求方法和路径自动推断业务类型
	return inferBusinessType(c)
}

// generateTitleFromPath 根据路径生成标题
func generateTitleFromPath(path string) string {
	// 移除 /api/v1 前缀
	if len(path) > 7 && path[:7] == "/api/v1" {
		path = path[7:]
	}

	// 提取资源名称（第一个路径段）
	parts := splitPath(path)
	if len(parts) == 0 {
		return "未知操作"
	}

	resource := parts[0]

	// 资源名称映射
	titleMap := map[string]string{
		"user":       "用户",
		"role":       "角色",
		"menu":       "菜单",
		"org":        "组织",
		"dict":       "字典",
		"config":     "配置",
		"loginLog":   "登录日志",
		"operLog":    "操作日志",
		"storageEnv": "存储环境",
		"attachment": "附件",
	}

	if title, ok := titleMap[resource]; ok {
		return title
	}

	return resource
}

// inferBusinessType 推断业务类型
func inferBusinessType(c *gin.Context) string {
	method := c.Request.Method
	path := c.Request.URL.Path

	// 特殊路径判断
	if containsSegment(path, "/export") {
		return "EXPORT"
	}
	if containsSegment(path, "/import") {
		return "IMPORT"
	}
	if containsSegment(path, "/grant") || containsSegment(path, "/permission") {
		return "GRANT"
	}
	if containsSegment(path, "/clean") {
		return "CLEAN"
	}
	if containsSegment(path, "/page") || containsSegment(path, "/list") {
		return "QUERY"
	}

	// 根据 HTTP 方法判断
	switch method {
	case http.MethodGet:
		return "QUERY"
	case http.MethodPost:
		return "CREATE"
	case http.MethodPut, http.MethodPatch:
		return "UPDATE"
	case http.MethodDelete:
		return "DELETE"
	default:
		return "OTHER"
	}
}

// splitPath 分割路径
func splitPath(path string) []string {
	var parts []string
	current := ""

	for _, ch := range path {
		if ch == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// containsSegment 检查路径是否包含指定段
func containsSegment(path, segment string) bool {
	parts := splitPath(path)
	for _, part := range parts {
		if part == segment[1:] { // 移除前导斜杠
			return true
		}
	}
	return false
}

// formatOperName 格式化操作者名称为 "用户ID-用户名称"
func formatOperName(userId, userName string) string {
	if userId == "" && userName == "" {
		return "-"
	}
	if userId == "" {
		return userName
	}
	if userName == "" {
		return userId
	}
	return fmt.Sprintf("%s-%s", userId, userName)
}

// shouldSkipLogging 判断是否应该跳过日志记录
func shouldSkipLogging(path string) bool {
	// 跳过操作日志相关的接口
	skipPaths := []string{
		"/api/v1/operLog",
		"/api/v1/oper-log",
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	return false
}
