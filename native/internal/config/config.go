package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	AppEnvVar       = "quick_admin_APP_ENV"
	LegacyAppEnvVar = "ADCS_APP_ENV"
)

// Server 定义业务数据结构。
type Server struct {
	Port int `mapstructure:"port"`
}

// Database 定义业务数据结构。
type Database struct {
	DSN                    string `mapstructure:"dsn"`
	MaxOpenConns           int    `mapstructure:"maxOpenConns"`
	MaxIdleConns           int    `mapstructure:"maxIdleConns"`
	ConnMaxLifetimeMinutes int    `mapstructure:"connMaxLifetimeMinutes"`
	SlowThreshold          int    `mapstructure:"slowThreshold"` // 慢 SQL 阈值（毫秒），默认 100ms
	AutoMigrate            bool   `mapstructure:"autoMigrate"`   // 是否开启 GORM 自动生成表结构
}

// Redis 定义业务数据结构。
type Redis struct {
	Addr, Password string
	DB             int
}

// JWT 定义业务数据结构。
type JWT struct {
	Secret string
	Expire int64
}

// WeChat 定义业务数据结构。
type WeChat struct {
	Enabled    bool   `mapstructure:"enabled" json:"enabled"`
	AppID      string `mapstructure:"appId" json:"appid"`
	Secret     string `mapstructure:"secret" json:"secret"`
	TemplateID string `mapstructure:"templateId" json:"templateId"`
}

// MQTT 定义业务数据结构。
type MQTT struct {
	Enabled                              bool `mapstructure:"enabled"`
	Broker, ClientID, Username, Password string
	QoS                                  byte
}

// RabbitMQ 定义业务数据结构。
type RabbitMQ struct {
	Enabled bool `mapstructure:"enabled"`
	URL     string
}

// SMS 定义业务数据结构。
type SMS struct {
	Enabled                                              bool `mapstructure:"enabled"`
	AccessKeyId, AccessKeySecret, SignName, TemplateCode string
}

// Email 定义业务数据结构。
type Email struct {
	Enabled  bool   `mapstructure:"enabled"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

// S3 定义业务数据结构。
type S3 struct {
	Enabled         bool   `mapstructure:"enabled"`         // 是否启用
	Endpoint        string `mapstructure:"endpoint"`        // MinIO/S3服务器地址
	AccessKeyID     string `mapstructure:"accessKeyId"`     // 访问密钥ID
	SecretAccessKey string `mapstructure:"secretAccessKey"` // 访问密钥
	Region          string `mapstructure:"region"`          // 区域
	Bucket          string `mapstructure:"bucket"`          // 默认存储桶
	UseSSL          bool   `mapstructure:"useSSL"`          // 是否使用SSL
	ForcePathStyle  bool   `mapstructure:"forcePathStyle"`  // 强制路径样式（MinIO需要）
}

// Scheduler 定义业务数据结构。
type Scheduler struct {
	Enabled bool `mapstructure:"enabled"`
}

// WebSocket 定义业务数据结构。
type WebSocket struct {
	Enabled             bool `mapstructure:"enabled"`
	TimeoutEnabled      bool `mapstructure:"timeoutEnabled"`
	ReadTimeoutSeconds  int  `mapstructure:"readTimeoutSeconds"`
	WriteTimeoutSeconds int  `mapstructure:"writeTimeoutSeconds"`
	HeartbeatEnabled    bool `mapstructure:"heartbeatEnabled"`
	MaxReadTimeouts     int  `mapstructure:"maxReadTimeouts"`
}

// Auth 认证配置
type Auth struct {
	TokenHeader     string `mapstructure:"tokenHeader"`     // Token 请求头名称，默认 "Authorization"
	AllowConcurrent bool   `mapstructure:"allowConcurrent"` // 是否允许并发登录，默认 false
	ShareToken      bool   `mapstructure:"shareToken"`      // 并发登录时是否共享 Token，默认 false
}

// Captcha 验证码配置
type Captcha struct {
	Image ImageCaptcha `mapstructure:"image"`
	SMS   SMSCaptcha   `mapstructure:"sms"`
	Email EmailCaptcha `mapstructure:"email"`
}

// ImageCaptcha 图形验证码配置
type ImageCaptcha struct {
	Enabled bool `mapstructure:"enabled"` // 是否启用
	Length  int  `mapstructure:"length"`  // 验证码长度
	Width   int  `mapstructure:"width"`   // 图片宽度
	Height  int  `mapstructure:"height"`  // 图片高度
	Expire  int  `mapstructure:"expire"`  // 过期时间（秒）
}

// SMSCaptcha 短信验证码配置
type SMSCaptcha struct {
	Enabled  bool   `mapstructure:"enabled"`  // 是否启用
	Length   int    `mapstructure:"length"`   // 验证码长度
	Expire   int    `mapstructure:"expire"`   // 过期时间（秒）
	Template string `mapstructure:"template"` // 短信模板
	Provider string `mapstructure:"provider"` // 短信服务商（aliyun/tencent）
}

// EmailCaptcha 邮箱验证码配置
type EmailCaptcha struct {
	Enabled  bool   `mapstructure:"enabled"`  // 是否启用
	Length   int    `mapstructure:"length"`   // 验证码长度
	Expire   int    `mapstructure:"expire"`   // 过期时间（秒）
	Template string `mapstructure:"template"` // 邮件模板
}

// Config 定义业务数据结构。
type Config struct {
	AppDir    string // 应用程序所在目录（可执行文件目录）
	Server    Server
	Database  Database
	Redis     Redis
	JWT       JWT
	Auth      Auth
	Captcha   Captcha // 验证码配置
	WeChat    WeChat
	MQTT      MQTT
	RabbitMQ  RabbitMQ
	SMS       SMS
	Email     Email
	S3        S3
	Scheduler Scheduler
	WebSocket WebSocket
	Env       string
}

// Load 执行业务逻辑。
func Load(appDir string) (*Config, *viper.Viper, error) {
	v := viper.New()
	env := CurrentEnv()

	configFileName := fmt.Sprintf("conf.%s.yaml", env)
	foundPath, err := ResolveFilePath(appDir, configFileName)
	if err != nil {
		return nil, nil, err
	}

	v.SetConfigFile(foundPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, nil, fmt.Errorf("failed to read config from %s: %w", foundPath, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, nil, err
	}

	// 设置应用程序目录
	cfg.AppDir = appDir
	cfg.Env = env

	if cfg.Database.DSN == "" {
		return nil, nil, fmt.Errorf("database dsn is required")
	}
	if cfg.Redis.Addr == "" {
		return nil, nil, fmt.Errorf("redis addr is required")
	}
	if cfg.JWT.Secret == "" {
		return nil, nil, fmt.Errorf("jwt secret is required")
	}
	// MQTT validation
	if cfg.MQTT.Enabled {
		if cfg.MQTT.Broker == "" || cfg.MQTT.ClientID == "" {
			return nil, nil, fmt.Errorf("mqtt broker and clientId are required when enabled")
		}
	}
	// RabbitMQ validation
	if cfg.RabbitMQ.Enabled {
		if cfg.RabbitMQ.URL == "" {
			return nil, nil, fmt.Errorf("rabbitmq url is required when enabled")
		}
	}
	// Auth 默认值设置
	if cfg.Auth.TokenHeader == "" {
		cfg.Auth.TokenHeader = "Authorization"
	}

	// Captcha 默认值设置
	if !v.IsSet("captcha.image.length") {
		cfg.Captcha.Image.Length = 4
	}
	if !v.IsSet("captcha.image.width") {
		cfg.Captcha.Image.Width = 120
	}
	if !v.IsSet("captcha.image.height") {
		cfg.Captcha.Image.Height = 40
	}
	if !v.IsSet("captcha.image.expire") {
		cfg.Captcha.Image.Expire = 300
	}
	if !v.IsSet("captcha.sms.length") {
		cfg.Captcha.SMS.Length = 6
	}
	if !v.IsSet("captcha.sms.expire") {
		cfg.Captcha.SMS.Expire = 300
	}
	if !v.IsSet("captcha.email.length") {
		cfg.Captcha.Email.Length = 6
	}
	if !v.IsSet("captcha.email.expire") {
		cfg.Captcha.Email.Expire = 300
	}
	return &cfg, v, nil
}

func CurrentEnv() string {
	env := os.Getenv(AppEnvVar)
	if env == "" {
		env = os.Getenv(LegacyAppEnvVar)
	}
	if env == "" {
		return "dev"
	}
	return env
}

func ResolveFilePath(appDir, fileName string) (string, error) {
	workDir, _ := os.Getwd()
	candidates := []string{
		filepath.Join(workDir, fileName),
		filepath.Join(workDir, "cmd", "api", fileName),
		filepath.Join(appDir, fileName),
	}

	for _, path := range candidates {
		if path == "" {
			continue
		}
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("config file not found: tried %v", candidates)
}
