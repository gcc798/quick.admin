package main

import (
	"flag"
	"os"

	"github.com/force-c/nai-tizi/kratos/app/sys-rpc/internal/biz"
	"github.com/force-c/nai-tizi/kratos/app/sys-rpc/internal/conf"
	"github.com/force-c/nai-tizi/kratos/app/sys-rpc/internal/data"
	"github.com/force-c/nai-tizi/kratos/app/sys-rpc/internal/server"
	"github.com/force-c/nai-tizi/kratos/app/sys-rpc/internal/service"
	kratos "github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"gopkg.in/yaml.v3"
)

func main() {
	configPath := flag.String("conf", "app/sys-rpc/configs/config.yaml", "config path")
	flag.Parse()
	cfg := mustLoadConfig(*configPath)
	logger := log.NewStdLogger(os.Stdout)
	resources := mustResources(data.NewResources(cfg.Data, cfg.Auth))
	defer func() {
		_ = resources.Close()
	}()

	healthSvc := service.NewHealthServiceService(biz.NewHealthUsecase())
	authSvc := service.NewAuthServiceService(biz.NewAuthUsecase(resources))
	captchaSvc := service.NewCaptchaServiceService(biz.NewCaptchaUsecase(resources))
	menuSvc := service.NewMenuServiceService(biz.NewMenuUsecase(resources))
	userSvc := service.NewUserServiceService(biz.NewUserUsecase(resources))
	roleSvc := service.NewRoleServiceService(biz.NewRoleUsecase(resources))
	orgSvc := service.NewOrgServiceService(biz.NewOrgUsecase(resources))
	configSvc := service.NewConfigServiceService(biz.NewConfigUsecase(resources))
	dictSvc := service.NewDictServiceService(biz.NewDictUsecase(resources))
	loginLogSvc := service.NewLoginLogServiceService(biz.NewLogUsecase(resources, "login"))
	operLogSvc := service.NewOperLogServiceService(biz.NewLogUsecase(resources, "oper"))
	storageSvc := service.NewStorageEnvServiceService(biz.NewStorageEnvUsecase(resources))
	attachmentSvc := service.NewAttachmentServiceService(biz.NewAttachmentUsecase(resources))
	grpcSrv := server.NewGRPCServer(cfg.Server.GRPC, healthSvc, authSvc, captchaSvc, menuSvc, userSvc, roleSvc, orgSvc, configSvc, dictSvc, loginLogSvc, operLogSvc, storageSvc, attachmentSvc)

	app := kratos.New(kratos.Name("sys-rpc"), kratos.Logger(logger), kratos.Server(grpcSrv))
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

func mustResources(resources *data.Resources, err error) *data.Resources {
	if err != nil {
		panic(err)
	}
	return resources
}
