package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Server struct {
	Port int `mapstructure:"port"`
}
type Database struct {
	DSN                    string `mapstructure:"dsn"`
	MaxOpenConns           int    `mapstructure:"maxOpenConns"`
	MaxIdleConns           int    `mapstructure:"maxIdleConns"`
	ConnMaxLifetimeMinutes int    `mapstructure:"connMaxLifetimeMinutes"`
	SlowThreshold          int    `mapstructure:"slowThreshold"` // 慢 SQL 阈值（毫秒），默认 100ms
	AutoMigrate            bool   `mapstructure:"autoMigrate"`   // 是否开启 GORM 自动生成表结构
}
type Redis struct {
	Addr, Password string
	DB             int
}
type JWT struct {
	Secret string
	Expire int64
}
type WeChat struct {
	Enabled    bool   `mapstructure:"enabled"`
	AppID      string `mapstructure:"appid"`
	Secret     string `mapstructure:"secret"`
	TemplateID string `mapstructure:"templateId"`
}
type MQTT struct {
	Enabled                              bool `mapstructure:"enabled"`
	Broker, ClientID, Username, Password string
	QoS                                  byte
}
type RabbitMQ struct {
	Enabled       bool `mapstructure:"enabled"`
	URL, Exchange string
}
type SMS struct {
	Enabled                                              bool `mapstructure:"enabled"`
	AccessKeyId, AccessKeySecret, SignName, TemplateCode string
}
type Email struct {
	Enabled  bool   `mapstructure:"enabled"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}
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

type Scheduler struct {
	Enabled bool `mapstructure:"enabled"`
}

type WebSocket struct {
	Enabled bool `mapstructure:"enabled"`
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

func Load(appDir string) (*Config, *viper.Viper, error) {
	v := viper.New()
	env := os.Getenv("NTZ_APP_ENV")
	if env == "" {
		env = "dev"
	}

	// 查找配置文件的多个可能位置
	configFileName := fmt.Sprintf("conf.%s.yaml", env)
	var configPath string
	var foundPath string

	// 1. 优先尝试当前工作目录（支持 IDE 调试）
	workDir, _ := os.Getwd()
	configPath = filepath.Join(workDir, configFileName)
	if _, err := os.Stat(configPath); err == nil {
		foundPath = configPath
	}

	// 2. 如果当前目录没找到，尝试可执行文件目录（生产环境）
	if foundPath == "" {
		configPath = filepath.Join(appDir, configFileName)
		if _, err := os.Stat(configPath); err == nil {
			foundPath = configPath
		}
	}

	// 3. 如果都没找到，返回错误
	if foundPath == "" {
		return nil, nil, fmt.Errorf("config file not found: tried %s and %s",
			filepath.Join(workDir, configFileName),
			filepath.Join(appDir, configFileName))
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
		if cfg.RabbitMQ.URL == "" || cfg.RabbitMQ.Exchange == "" {
			return nil, nil, fmt.Errorf("rabbitmq url and exchange are required when enabled")
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
