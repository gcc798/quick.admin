package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	_ "github.com/force-c/nai-tizi/docs/swagger" // 导入 Swagger 文档
	"github.com/force-c/nai-tizi/internal/bootstrap"
	"github.com/force-c/nai-tizi/internal/config"
	"github.com/force-c/nai-tizi/internal/container"
	ilogger "github.com/force-c/nai-tizi/internal/logger"
	"github.com/force-c/nai-tizi/internal/middleware"
	"github.com/force-c/nai-tizi/internal/router"
	"github.com/force-c/nai-tizi/internal/validator"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

//	@title			NTZ API 文档
//	@version		1.0
//	@description	Nai-Tizi RESTful API 接口文档，支持双 Token 认证机制
//	@termsOfService	https://example.com/terms/

//	@contact.name	技术支持
//	@contact.url	https://example.com/support
//	@contact.email	support@example.com

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

//	@host		localhost:9009
//	@BasePath	/

//	@securityDefinitions.apikey	Bearer
//	@in							header
//	@name						Authorization
//	@description				格式: "Bearer {access_token}"，AccessToken 从登录接口获取

//	@tag.name			认证
//	@tag.description	用户认证相关接口，包括登录、登出、Token 刷新等

//	@tag.name			用户管理
//	@tag.description	用户信息管理接口

func main() {
	// 获取可执行文件所在目录
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("failed to get executable path: %v\n", err)
		os.Exit(1)
	}
	execDir := filepath.Dir(execPath)

	// 加载配置（基于可执行文件目录）
	cfg, v, err := config.Load(execDir)
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	log, err := ilogger.NewLogger(config.CurrentEnv(), cfg.AppDir)
	if err != nil {
		fmt.Printf("failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// 初始化容器
	c, err := container.New(cfg, v, log)
	if err != nil {
		fmt.Printf("failed to initialize container: %v\n", err)
		os.Exit(1)
	}

	logger := c.GetLogger()
	logger.Info("container initialized successfully")

	b, err := bootstrap.New(c)
	if err != nil {
		logger.Fatal("failed to bootstrap business components", zap.Error(err))
	}

	if err := c.Start(); err != nil {
		logger.Fatal("failed to start background services", zap.Error(err))
	}
	defer c.Stop()

	// 调度器已由容器内部启动

	// 设置Gin模式
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化HTTP引擎
	r := gin.New()
	//binding.EnableDecoderUseNumber = true

	// 初始化中文验证错误翻译器
	validator.Init()

	// 使用自定义的 Recovery 中间件（支持结构化错误处理）
	// 注意：这里替换了 gin.Recovery()，提供更强大的错误处理能力
	r.Use(middleware.Recovery(logger.Get()))
	r.Use(gin.Logger())

	// 添加 CORS 中间件（必须在路由注册之前）
	r.Use(middleware.CORS())

	// 添加字符串ID转换中间件（处理前端传递的字符串ID）
	r.Use(middleware.StringIDConverter())

	// 添加操作日志中间件（全局记录所有接口访问）
	r.Use(middleware.OperationLog(c.GetDB(), c.GetLogger()))

	// 注册路由
	router.Setup(r, c, b)

	// 配置HTTP服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:           addr,
		Handler:        r,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// 在 goroutine 中启动服务器
	go func() {
		logger.Info("starting http server", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("failed to start http server", zap.Error(err))
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	// 默认终止命令发送 syscall.SIGTERM。
	// 中断命令发送 syscall.SIGINT。
	// 强制终止命令发送 syscall.SIGKILL，进程无法捕获，不需要监听。
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server...")

	// 设置 30 秒的超时时间用于优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 优雅关闭 HTTP 服务器
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server exited gracefully")
}
