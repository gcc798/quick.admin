package main

import (
	"flag"
	"os"

	"github.com/force-c/nai-tizi/kratos/app/sys-api/internal/biz"
	"github.com/force-c/nai-tizi/kratos/app/sys-api/internal/conf"
	"github.com/force-c/nai-tizi/kratos/app/sys-api/internal/data"
	"github.com/force-c/nai-tizi/kratos/app/sys-api/internal/server"
	"github.com/force-c/nai-tizi/kratos/app/sys-api/internal/service"
	kratos "github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"gopkg.in/yaml.v3"
)

func main() {
	configPath := flag.String("conf", "app/sys-api/configs/config.yaml", "config path")
	flag.Parse()
	cfg := mustLoadConfig(*configPath)
	logger := log.NewStdLogger(os.Stdout)

	healthRepo := mustData(data.NewHealthRepo(cfg.Server.RPC.Endpoint))
	defer healthRepo.Close()
	authRepo := mustData(data.NewAuthRepo(cfg.Server.RPC.Endpoint))
	defer authRepo.Close()
	captchaRepo := mustData(data.NewCaptchaRepo(cfg.Server.RPC.Endpoint))
	defer captchaRepo.Close()
	menuRepo := mustData(data.NewMenuRepo(cfg.Server.RPC.Endpoint))
	defer menuRepo.Close()
	userRepo := mustData(data.NewUserRepo(cfg.Server.RPC.Endpoint))
	defer userRepo.Close()
	roleRepo := mustData(data.NewRoleRepo(cfg.Server.RPC.Endpoint))
	defer roleRepo.Close()
	orgRepo := mustData(data.NewOrgRepo(cfg.Server.RPC.Endpoint))
	defer orgRepo.Close()
	configRepo := mustData(data.NewConfigRepo(cfg.Server.RPC.Endpoint))
	defer configRepo.Close()
	dictRepo := mustData(data.NewDictRepo(cfg.Server.RPC.Endpoint))
	defer dictRepo.Close()
	loginLogRepo := mustData(data.NewLoginLogRepo(cfg.Server.RPC.Endpoint))
	defer loginLogRepo.Close()
	operLogRepo := mustData(data.NewOperLogRepo(cfg.Server.RPC.Endpoint))
	defer operLogRepo.Close()
	storageRepo := mustData(data.NewStorageEnvRepo(cfg.Server.RPC.Endpoint))
	defer storageRepo.Close()
	attachmentRepo := mustData(data.NewAttachmentRepo(cfg.Server.RPC.Endpoint))
	defer attachmentRepo.Close()

	healthUC := biz.NewHealthUsecase(healthRepo)
	authUC := biz.NewAuthUsecase(authRepo)
	captchaUC := biz.NewCaptchaUsecase(captchaRepo)
	menuUC := biz.NewMenuUsecase(menuRepo)
	userUC := biz.NewUserUsecase(userRepo)
	roleUC := biz.NewRoleUsecase(roleRepo)
	orgUC := biz.NewOrgUsecase(orgRepo)
	configUC := biz.NewConfigUsecase(configRepo)
	dictUC := biz.NewDictUsecase(dictRepo)
	loginLogUC := biz.NewLoginLogUsecase(loginLogRepo)
	operLogUC := biz.NewOperLogUsecase(operLogRepo)
	storageUC := biz.NewStorageEnvUsecase(storageRepo)
	attachmentUC := biz.NewAttachmentUsecase(attachmentRepo)

	healthSvc := service.NewHealthServiceService(healthUC)
	authSvc := service.NewAuthServiceService(authUC)
	captchaSvc := service.NewCaptchaServiceService(captchaUC)
	menuSvc := service.NewMenuServiceService(menuUC)
	userSvc := service.NewUserServiceService(userUC)
	roleSvc := service.NewRoleServiceService(roleUC)
	orgSvc := service.NewOrgServiceService(orgUC)
	configSvc := service.NewConfigServiceService(configUC)
	dictSvc := service.NewDictServiceService(dictUC)
	loginLogSvc := service.NewLoginLogServiceService(loginLogUC)
	operLogSvc := service.NewOperLogServiceService(operLogUC)
	storageSvc := service.NewStorageEnvServiceService(storageUC)
	attachmentSvc := service.NewAttachmentServiceService(attachmentUC)

	httpSrv := server.NewHTTPServer(
		cfg.Server.HTTP,
		&server.GatewayDeps{Auth: authRepo, Attachment: attachmentUC, OperLog: operLogRepo, WebSocket: server.NewWebSocketHub()},
		healthSvc,
		authSvc,
		captchaSvc,
		menuSvc,
		userSvc,
		roleSvc,
		orgSvc,
		configSvc,
		dictSvc,
		loginLogSvc,
		operLogSvc,
		storageSvc,
		attachmentSvc,
	)

	app := kratos.New(kratos.Name("sys-api"), kratos.Logger(logger), kratos.Server(httpSrv))
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func mustLoadConfig(path string) *conf.Bootstrap {
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var cfg conf.Bootstrap
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		panic(err)
	}
	return &cfg
}

func mustData[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}
