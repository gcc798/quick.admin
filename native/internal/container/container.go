package container

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/force-c/nai-tizi/internal/config"
	"github.com/force-c/nai-tizi/internal/domain/model"
	"github.com/force-c/nai-tizi/internal/infrastructure/captcha"
	"github.com/force-c/nai-tizi/internal/infrastructure/database"
	// "github.com/force-c/nai-tizi/internal/infrastructure/idempotent" // TODO: 待实现
	"github.com/force-c/nai-tizi/internal/infrastructure/jwt"
	"github.com/force-c/nai-tizi/internal/infrastructure/mqtt"
	mqtthandler "github.com/force-c/nai-tizi/internal/infrastructure/mqtt/handler"
	"github.com/force-c/nai-tizi/internal/infrastructure/mqtt/retry"
	"github.com/force-c/nai-tizi/internal/infrastructure/rabbitmq"
	"github.com/force-c/nai-tizi/internal/infrastructure/redis"
	"github.com/force-c/nai-tizi/internal/infrastructure/s3"
	"github.com/force-c/nai-tizi/internal/infrastructure/scheduler"
	"github.com/force-c/nai-tizi/internal/infrastructure/scheduler/jobs"
	"github.com/force-c/nai-tizi/internal/infrastructure/storage"
	"github.com/force-c/nai-tizi/internal/infrastructure/thirdparty/email"
	"github.com/force-c/nai-tizi/internal/infrastructure/thirdparty/sms"
	"github.com/force-c/nai-tizi/internal/infrastructure/thirdparty/wechat"
	"github.com/force-c/nai-tizi/internal/infrastructure/websocket"
	"github.com/force-c/nai-tizi/internal/logger"
	goredis "github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Component 统一组件接口
type Component interface {
	Name() string
	Start() error
	Stop() error
}

// Container 依赖注入容器接口
type Container interface {
	GetConfig() *config.Config
	GetViper() *viper.Viper
	GetDB() *gorm.DB
	GetRedis() *goredis.Client
	GetJWT() *jwt.Jwt
	GetLogger() logger.Logger
	GetCasbin() *casbin.Enforcer
	GetMQTT() *mqtt.Client
	GetRabbitMQProducer() *rabbitmq.ProducerService
	GetRetryManager() *retry.Manager
	GetWeChat() *wechat.Manager
	GetSMS() *sms.Manager
	GetEmail() *email.Manager
	GetS3() *s3.Manager
	GetStorageManager() storage.StorageManager
	GetWebSocketHub() *websocket.Hub
	GetScheduler() *scheduler.Scheduler
	// GetIdempotent() *idempotent.Idempotent // TODO: 待实现
	GetCaptchaManager() *captcha.CaptchaManager
	Start() error
	Stop()
}

type container struct {
	config         *config.Config
	viper          *viper.Viper
	db             *gorm.DB
	redis          *goredis.Client
	jwt            *jwt.Jwt
	logger         logger.Logger
	casbin         *casbin.Enforcer
	mqttClient     *mqtt.Client
	rabbitMQ       *rabbitmq.Manager
	retryManager   *retry.Manager
	wechatManager  *wechat.Manager
	smsManager     *sms.Manager
	emailManager   *email.Manager
	s3Manager      *s3.Manager
	storageManager storage.StorageManager
	wsHub          *websocket.Hub
	sched          *scheduler.Scheduler
	// idempotent     *idempotent.Idempotent // TODO: 待实现
	captchaManager *captcha.CaptchaManager

	components []Component
}

// New 创建新的容器实例
func New(cfg *config.Config, v *viper.Viper) (Container, error) {
	c := &container{
		config:     cfg,
		viper:      v,
		components: make([]Component, 0),
	}

	// 1. 初始化基础组件
	if err := c.initLogger(); err != nil {
		return nil, err
	}
	if err := c.initDB(); err != nil {
		return nil, err
	}
	if err := c.initRedis(); err != nil {
		return nil, err
	}
	c.initJWT()
	// c.initIdempotent() // TODO: 待实现
	if err := c.initCasbin(); err != nil {
		return nil, err
	}

	// 2. 初始化业务组件
	if err := c.initMQTT(); err != nil {
		return nil, err
	}
	if err := c.initRabbitMQ(); err != nil {
		return nil, err
	}
	c.initRetryManager()
	c.initThirdParty()
	c.initStorageManager()
	c.initWebSocket()
	c.initCaptchaManager()

	// 3. 初始化调度器（依赖其他组件）
	c.initScheduler()
	if err := c.registerJobs(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *container) RegisterComponent(comp Component) {
	c.components = append(c.components, comp)
}

// initLogger 初始化日志
func (c *container) initLogger() error {
	log, err := logger.NewLogger(c.config.Env)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	c.logger = log
	return nil
}

// initDB 初始化数据库
func (c *container) initDB() error {
	dsn := c.config.Database.DSN
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(c.config.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(c.config.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(c.config.Database.ConnMaxLifetimeMinutes) * time.Minute)

	// 注册 ID 生成插件
	idGenPlugin := &database.IDGenPlugin{}
	if err := db.Use(idGenPlugin); err != nil {
		c.logger.Warn("failed to register ID generation plugin", zap.Error(err))
	}

	// 注册慢 SQL 监控插件
	slowThreshold := 100 * time.Millisecond // 默认阈值 100ms
	if c.config.Database.SlowThreshold > 0 {
		slowThreshold = time.Duration(c.config.Database.SlowThreshold) * time.Millisecond
	}
	slowQueryPlugin := database.NewSlowQueryPlugin(c.logger, slowThreshold)
	if err := db.Use(slowQueryPlugin); err != nil {
		c.logger.Warn("failed to register slow query plugin", zap.Error(err))
	}

	// GORM AutoMigrate 配置
	if c.config.Database.AutoMigrate {
		c.logger.Info("starting database auto migration...")
		if err := db.AutoMigrate(
			&model.User{},
			&model.DictData{},
			&model.LoginLog{},
			&model.OperLog{},
			&model.AuthClient{},
			&model.BuMessageRetry{},
			&model.BuMessageRetryLog{},
			&model.Role{},
			&model.Menu{},
			&model.Org{},
			&model.Config{},
			&model.StorageEnv{},
			&model.Attachment{},
			&model.CasbinRule{},
			&model.MUserRole{},
			&model.MRoleMenu{},
		); err != nil {
			return fmt.Errorf("failed to auto migrate: %w", err)
		}
		c.logger.Info("database auto migration completed")
	}
	c.db = db
	return nil
}

// initRedis 初始化Redis
func (c *container) initRedis() error {
	redisClient := redis.NewRedis(c.config.Redis.Addr, c.config.Redis.Password, c.config.Redis.DB)
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}
	c.redis = redisClient
	return nil
}

// initJWT 初始化JWT
func (c *container) initJWT() {
	c.jwt = jwt.New(c.config.JWT.Secret, int64(c.config.JWT.Expire))
}

// initIdempotent 初始化幂等处理器 (TODO: 待实现)
// func (c *container) initIdempotent() {
// 	c.idempotent = idempotent.New(c.db)
// }

// initCasbin 初始化 Casbin 权限管理
func (c *container) initCasbin() error {
	// 使用 GORM Adapter 连接数据库
	adapter, err := gormadapter.NewAdapterByDB(c.db)
	if err != nil {
		return fmt.Errorf("failed to create casbin adapter: %w", err)
	}

	// 查找 Casbin 模型配置文件的多个可能位置
	var modelPath string
	modelFileName := "casbin_model.conf"

	// 1. 优先尝试当前工作目录（支持 IDE 调试）
	workDir, _ := os.Getwd()
	tryPath := filepath.Join(workDir, modelFileName)
	if _, err := os.Stat(tryPath); err == nil {
		modelPath = tryPath
	}

	// 2. 如果当前目录没找到，尝试可执行文件目录（生产环境）
	if modelPath == "" {
		tryPath = filepath.Join(c.config.AppDir, modelFileName)
		if _, err := os.Stat(tryPath); err == nil {
			modelPath = tryPath
		}
	}

	// 3. 如果都没找到，返回错误
	if modelPath == "" {
		return fmt.Errorf("casbin model file not found: tried %s and %s",
			filepath.Join(workDir, modelFileName),
			filepath.Join(c.config.AppDir, modelFileName))
	}

	c.logger.Info("loading casbin model", zap.String("path", modelPath))

	// 加载 Casbin 模型配置文件
	enforcer, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		return fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	// 添加通配符匹配函数（支持 * 通配符）
	enforcer.AddFunction("keyMatch2", func(args ...interface{}) (interface{}, error) {
		name1 := args[0].(string)
		name2 := args[1].(string)

		// 如果策略是 *，匹配所有
		if name2 == "*" {
			return true, nil
		}

		// 如果策略以 * 结尾，例如 "user.*"
		if len(name2) > 0 && name2[len(name2)-1] == '*' {
			prefix := name2[:len(name2)-1]
			return len(name1) >= len(prefix) && name1[:len(prefix)] == prefix, nil
		}

		// 如果策略以 * 开头，例如 "*.read"
		if len(name2) > 0 && name2[0] == '*' {
			suffix := name2[1:]
			return len(name1) >= len(suffix) && name1[len(name1)-len(suffix):] == suffix, nil
		}

		// 精确匹配
		return name1 == name2, nil
	})

	// 启用日志（开发环境）
	if c.config.Env == "development" || c.config.Env == "dev" {
		enforcer.EnableLog(true)
	}

	// 加载策略
	if err := enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("failed to load casbin policy: %w", err)
	}

	c.casbin = enforcer
	c.logger.Info("casbin enforcer initialized successfully")
	return nil
}

// initMQTT 初始化MQTT
func (c *container) initMQTT() error {
	if !c.config.MQTT.Enabled {
		return nil
	}
	client, err := mqtt.NewClient(&mqtt.Config{
		Broker:   c.config.MQTT.Broker,
		ClientID: c.config.MQTT.ClientID,
		Username: c.config.MQTT.Username,
		Password: c.config.MQTT.Password,
		QoS:      c.config.MQTT.QoS,
	}, c.logger)
	if err != nil {
		return fmt.Errorf("failed to create MQTT client: %w", err)
	}

	// 注册MQTT消息处理器
	handler := mqtthandler.NewMessageHandler(c.db, c.logger)
	client.AddSubscription("scaffold/#", handler.Handle)

	c.mqttClient = client
	c.RegisterComponent(client)
	return nil
}

// initRabbitMQ 初始化RabbitMQ
func (c *container) initRabbitMQ() error {
	if !c.config.RabbitMQ.Enabled {
		return nil
	}
	manager, err := rabbitmq.NewManager(&rabbitmq.Config{
		URL:      c.config.RabbitMQ.URL,
		Exchange: c.config.RabbitMQ.Exchange,
	}, c.db, c.logger)
	if err != nil {
		c.logger.Warn("failed to create RabbitMQ manager", zap.Error(err))
		return nil // 允许失败，不阻断启动
	}
	c.rabbitMQ = manager
	c.RegisterComponent(manager)
	return nil
}

// initRetryManager 初始化重试管理器
func (c *container) initRetryManager() {
	if !c.config.MQTT.Enabled {
		return
	}
	retryConfig := &retry.RetryConfig{
		Enabled:            true,
		MaxRetryCount:      3,
		RetryInterval:      2000,
		ScanInterval:       5000,
		MaxBatchSize:       100,
		LockWaitTime:       200,
		LockLeaseTime:      10000,
		RedisExpireMinutes: 30,
		AbandonTimeout:     300000, // 5 minutes
		CleanupInterval:    600000, // 10 minutes
	}
	c.retryManager = retry.NewManager(c.db, c.mqttClient, c.redis, c.logger, retryConfig)
	c.RegisterComponent(c.retryManager)
}

// initThirdParty 初始化第三方服务
func (c *container) initThirdParty() {
	if c.config.WeChat.Enabled {
		c.wechatManager = wechat.NewManager(wechat.Config{
			AppID:  c.config.WeChat.AppID,
			Secret: c.config.WeChat.Secret,
		}, c.logger, c.redis)
	}

	if c.config.SMS.Enabled {
		smsManager, err := sms.NewManager(sms.Config{
			AccessKeyId:     c.config.SMS.AccessKeyId,
			AccessKeySecret: c.config.SMS.AccessKeySecret,
			SignName:        c.config.SMS.SignName,
			TemplateCode:    c.config.SMS.TemplateCode,
		}, c.redis, c.logger)
		if err != nil {
			c.logger.Warn("failed to create SMS service", zap.Error(err))
		} else {
			c.smsManager = smsManager
		}
	}

	if c.config.Email.Enabled {
		emailManager, err := email.NewManager(email.Config{
			Host:     c.config.Email.Host,
			Port:     c.config.Email.Port,
			Username: c.config.Email.Username,
			Password: c.config.Email.Password,
			From:     c.config.Email.From,
		}, c.redis, c.logger)
		if err != nil {
			c.logger.Warn("failed to create email service", zap.Error(err))
		} else {
			c.emailManager = emailManager
		}
	}

	if c.config.S3.Enabled {
		s3Manager, err := s3.NewManager(&s3.Config{
			Enabled:         c.config.S3.Enabled,
			Endpoint:        c.config.S3.Endpoint,
			AccessKeyID:     c.config.S3.AccessKeyID,
			SecretAccessKey: c.config.S3.SecretAccessKey,
			Region:          c.config.S3.Region,
			Bucket:          c.config.S3.Bucket,
			UseSSL:          c.config.S3.UseSSL,
			ForcePathStyle:  c.config.S3.ForcePathStyle,
		}, c.logger)
		if err != nil {
			c.logger.Warn("failed to create S3 manager", zap.Error(err))
		} else {
			c.s3Manager = s3Manager
		}
	}
}

// initStorageManager 初始化存储管理器
func (c *container) initStorageManager() {
	// 创建存储管理器
	c.storageManager = storage.NewStorageManager(c.db, c.logger)

	// 注册存储类型工厂
	c.storageManager.RegisterStorageType("s3", storage.NewS3StorageFactory())
	c.storageManager.RegisterStorageType("local", storage.NewLocalStorageFactory())

	c.logger.Info("storage manager initialized successfully")
}

// initCaptchaManager 初始化验证码管理器
func (c *container) initCaptchaManager() {
	manager := captcha.NewCaptchaManager()

	// 注册图形验证码提供者
	if c.config.Captcha.Image.Enabled {
		imageConfig := &captcha.ImageCaptchaConfig{
			Enabled: c.config.Captcha.Image.Enabled,
			Length:  c.config.Captcha.Image.Length,
			Width:   c.config.Captcha.Image.Width,
			Height:  c.config.Captcha.Image.Height,
			Expire:  c.config.Captcha.Image.Expire,
		}
		imageProvider := captcha.NewImageCaptchaProvider(imageConfig, c.redis)
		manager.RegisterProvider(imageProvider)
	}

	// 注册短信验证码提供者
	if c.config.Captcha.SMS.Enabled {
		if c.smsManager == nil {
			c.logger.Warn("SMS captcha enabled but SMS service not configured")
		} else {
			smsConfig := &captcha.SMSCaptchaConfig{
				Enabled:  c.config.Captcha.SMS.Enabled,
				Length:   c.config.Captcha.SMS.Length,
				Expire:   c.config.Captcha.SMS.Expire,
				Template: c.config.Captcha.SMS.Template,
				Provider: c.config.Captcha.SMS.Provider,
			}
			smsAdapter := captcha.NewSMSManagerAdapter(c.smsManager)
			smsProvider := captcha.NewSMSCaptchaProvider(smsConfig, c.redis, smsAdapter)
			manager.RegisterProvider(smsProvider)
		}
	}

	// 注册邮箱验证码提供者
	if c.config.Captcha.Email.Enabled {
		if c.emailManager == nil {
			c.logger.Warn("Email captcha enabled but email service not configured")
		} else {
			emailConfig := &captcha.EmailCaptchaConfig{
				Enabled:  c.config.Captcha.Email.Enabled,
				Length:   c.config.Captcha.Email.Length,
				Expire:   c.config.Captcha.Email.Expire,
				Template: c.config.Captcha.Email.Template,
			}
			emailAdapter := captcha.NewEmailManagerAdapter(c.emailManager)
			emailProvider := captcha.NewEmailCaptchaProvider(emailConfig, c.redis, emailAdapter)
			manager.RegisterProvider(emailProvider)
		}
	}

	c.captchaManager = manager
	c.logger.Info("captcha manager initialized successfully")
}

// initWebSocket 初始化WebSocket
func (c *container) initWebSocket() {
	if !c.config.WebSocket.Enabled {
		return
	}
	c.wsHub = websocket.NewHub(c.logger)
	c.RegisterComponent(c.wsHub)
}

// initScheduler 初始化调度器
func (c *container) initScheduler() {
	if !c.config.Scheduler.Enabled {
		return
	}
	c.sched = scheduler.New(c.logger)
	c.RegisterComponent(c.sched)
}

func (c *container) GetConfig() *config.Config {
	return c.config
}

func (c *container) GetViper() *viper.Viper { return c.viper }

func (c *container) GetDB() *gorm.DB {
	return c.db
}

func (c *container) GetRedis() *goredis.Client {
	return c.redis
}

func (c *container) GetJWT() *jwt.Jwt {
	return c.jwt
}

func (c *container) GetLogger() logger.Logger {
	return c.logger
}

func (c *container) GetCasbin() *casbin.Enforcer {
	return c.casbin
}

func (c *container) GetMQTT() *mqtt.Client {
	return c.mqttClient
}

func (c *container) GetRabbitMQProducer() *rabbitmq.ProducerService {
	if c.rabbitMQ == nil {
		return nil
	}
	return c.rabbitMQ.GetProducer()
}

func (c *container) GetRetryManager() *retry.Manager {
	return c.retryManager
}

func (c *container) GetWeChat() *wechat.Manager {
	return c.wechatManager
}

func (c *container) GetSMS() *sms.Manager {
	return c.smsManager
}

func (c *container) GetEmail() *email.Manager {
	return c.emailManager
}

func (c *container) GetS3() *s3.Manager {
	return c.s3Manager
}

func (c *container) GetStorageManager() storage.StorageManager {
	return c.storageManager
}

func (c *container) GetWebSocketHub() *websocket.Hub {
	return c.wsHub
}

func (c *container) GetScheduler() *scheduler.Scheduler {
	return c.sched
}

// GetIdempotent TODO: 待实现
// func (c *container) GetIdempotent() *idempotent.Idempotent {
// 	return c.idempotent
// }

func (c *container) GetCaptchaManager() *captcha.CaptchaManager {
	return c.captchaManager
}

func (c *container) Start() error {
	for _, comp := range c.components {
		c.logger.Info("starting component", zap.String("name", comp.Name()))
		if err := comp.Start(); err != nil {
			c.logger.Error("failed to start component", zap.String("name", comp.Name()), zap.Error(err))
			return err
		}
	}
	c.logger.Info("all components started successfully")
	return nil
}

func (c *container) Stop() {
	// 反向停止
	for i := len(c.components) - 1; i >= 0; i-- {
		comp := c.components[i]
		c.logger.Info("stopping component", zap.String("name", comp.Name()))
		if err := comp.Stop(); err != nil {
			c.logger.Error("failed to stop component", zap.String("name", comp.Name()), zap.Error(err))
		}
	}
	c.logger.Info("all components stopped")
}

func (c *container) registerJobs() error {
	if c.sched == nil {
		return nil
	}

	return jobs.RegisterJobs(c.sched, c.db, c.redis, c.retryManager, c.logger)
}
