package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
)

// Logger 日志接口
type Logger interface {
	Get() *zap.Logger
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
}

// Config 日志配置
type Config struct {
	Level      string        `yaml:"level"`      // 日志级别: debug, info, warn, error, fatal
	Output     string        `yaml:"output"`     // 输出模式: console, file, both
	Encoding   string        `yaml:"encoding"`   // 日志格式: json, console
	File       FileConfig    `yaml:"file"`       // 文件输出配置
	Console    ConsoleConfig `yaml:"console"`    // 控制台输出配置
	Caller     bool          `yaml:"caller"`     // 是否显示调用者信息
	Stacktrace bool          `yaml:"stacktrace"` // 是否显示堆栈跟踪
}

// FileConfig 文件输出配置
type FileConfig struct {
	Path       string `yaml:"path"`       // 日志文件路径
	Filename   string `yaml:"filename"`   // 日志文件名
	MaxSize    int    `yaml:"maxSize"`    // 单个文件最大大小(MB)
	MaxBackups int    `yaml:"maxBackups"` // 最多保留的旧文件数量
	MaxAge     int    `yaml:"maxAge"`     // 文件最多保留天数
	Compress   bool   `yaml:"compress"`   // 是否压缩旧文件
}

// ConsoleConfig 控制台输出配置
type ConsoleConfig struct {
	Colorful bool `yaml:"colorful"` // 是否启用彩色输出
}

type zapLogger struct {
	logger *zap.Logger
}

// LoadConfig 从配置文件加载日志配置
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 设置默认值
	if config.Level == "" {
		config.Level = "info"
	}
	if config.Output == "" {
		config.Output = "console"
	}
	if config.Encoding == "" {
		config.Encoding = "console"
	}
	if config.File.Path == "" {
		config.File.Path = "./logs"
	}
	if config.File.Filename == "" {
		config.File.Filename = "app"
	}
	if config.File.MaxSize == 0 {
		config.File.MaxSize = 100
	}
	if config.File.MaxBackups == 0 {
		config.File.MaxBackups = 30
	}
	if config.File.MaxAge == 0 {
		config.File.MaxAge = 7
	}

	return &config, nil
}

// NewLogger 创建新的Logger实例
func NewLogger(env string) (Logger, error) {
	// 标准化环境名称：development -> dev, production -> prod
	normalizedEnv := normalizeEnv(env)

	// 根据环境选择配置文件
	configFile := fmt.Sprintf("cmd/api/zaplogger.%s.yaml", normalizedEnv)

	// 如果配置文件不存在，使用默认配置
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return newDefaultLogger(env)
	}

	config, err := LoadConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load logger config: %w", err)
	}

	return NewLoggerWithConfig(config)
}

// normalizeEnv 标准化环境名称
func normalizeEnv(env string) string {
	switch env {
	case "development", "dev":
		return "dev"
	case "production", "prod":
		return "prod"
	default:
		return env
	}
}

// NewLoggerWithConfig 使用配置创建Logger
func NewLoggerWithConfig(config *Config) (Logger, error) {
	// 解析日志级别
	level := parseLevel(config.Level)

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 如果是控制台格式且启用彩色，使用彩色级别编码器
	if config.Encoding == "console" && config.Console.Colorful {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// 创建编码器
	var encoder zapcore.Encoder
	if config.Encoding == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 创建输出
	var cores []zapcore.Core

	switch config.Output {
	case "console":
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level))
	case "file":
		fileWriter := getFileWriter(config)
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(fileWriter), level))
	case "all":
		// 控制台输出
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level))
		// 文件输出
		fileWriter := getFileWriter(config)
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(fileWriter), level))
	default:
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level))
	}

	// 合并所有core
	core := zapcore.NewTee(cores...)

	// 创建logger选项
	opts := []zap.Option{}

	if config.Caller {
		// AddCaller() 显示调用者信息
		// AddCallerSkip(1) 跳过一层包装，显示实际调用代码的位置而不是logger包装层
		opts = append(opts, zap.AddCaller(), zap.AddCallerSkip(1))
	}

	if config.Stacktrace {
		opts = append(opts, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	// 创建logger
	logger := zap.New(core, opts...)

	return &zapLogger{logger: logger}, nil
}

// getFileWriter 创建文件写入器（支持日志轮转）
func getFileWriter(config *Config) *lumberjack.Logger {
	// 确保日志目录存在
	if err := os.MkdirAll(config.File.Path, 0755); err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
	}

	// 生成带时间戳的文件名
	now := time.Now()
	filename := fmt.Sprintf("%s-%s.log",
		config.File.Filename,
		now.Format("2006-01-02"))
	fullPath := filepath.Join(config.File.Path, filename)

	return &lumberjack.Logger{
		Filename:   fullPath,
		MaxSize:    config.File.MaxSize,    // MB
		MaxBackups: config.File.MaxBackups, // 保留的旧文件数量
		MaxAge:     config.File.MaxAge,     // 天
		Compress:   config.File.Compress,   // 是否压缩
		LocalTime:  true,                   // 使用本地时间
	}
}

// customTimeEncoder 自定义时间格式
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// parseLevel 解析日志级别
func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// newDefaultLogger 创建默认logger（向后兼容）
func newDefaultLogger(env string) (Logger, error) {
	var l *zap.Logger
	var err error
	if env == "production" || env == "prod" {
		l, err = zap.NewProduction(zap.AddCallerSkip(1))
	} else {
		l, err = zap.NewDevelopment(zap.AddCallerSkip(1))
	}
	if err != nil {
		return nil, err
	}
	return &zapLogger{logger: l}, nil
}

// 实现Logger接口

func (l *zapLogger) Get() *zap.Logger { return l.logger }

func (l *zapLogger) Debug(msg string, fields ...zap.Field) { l.logger.Debug(msg, fields...) }
func (l *zapLogger) Info(msg string, fields ...zap.Field)  { l.logger.Info(msg, fields...) }
func (l *zapLogger) Warn(msg string, fields ...zap.Field)  { l.logger.Warn(msg, fields...) }
func (l *zapLogger) Error(msg string, fields ...zap.Field) { l.logger.Error(msg, fields...) }
func (l *zapLogger) Fatal(msg string, fields ...zap.Field) { l.logger.Fatal(msg, fields...) }
func (l *zapLogger) With(fields ...zap.Field) Logger {
	return &zapLogger{logger: l.logger.With(fields...)}
}
