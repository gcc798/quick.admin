//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/force-c/nai-tizi/kratos/application/sys-rpc/internal/biz"
	"github.com/force-c/nai-tizi/kratos/application/sys-rpc/internal/conf"
	"github.com/force-c/nai-tizi/kratos/application/sys-rpc/internal/data"
	rpcsrv "github.com/force-c/nai-tizi/kratos/application/sys-rpc/internal/server"
	"github.com/force-c/nai-tizi/kratos/application/sys-rpc/internal/service"
	kratos "github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
)

func newResources(cfg *conf.Bootstrap) (*data.Resources, func(), error) {
	resources, err := data.NewResources(cfg.Data, cfg.Auth, cfg.JWT, cfg.Observability)
	if err != nil {
		return nil, nil, err
	}
	return resources, func() {
		_ = resources.Close()
	}, nil
}

func newGRPCServer(cfg *conf.Bootstrap, services *service.Services) *kgrpc.Server {
	return rpcsrv.NewGRPCServer(
		cfg.Server.GRPC,
		services.Health,
		services.Auth,
		services.Captcha,
		services.Menu,
		services.User,
		services.Role,
		services.Org,
		services.Config,
		services.Dict,
		services.LoginLog,
		services.OperLog,
		services.StorageEnv,
		services.Attachment,
	)
}

func newApp(cfg *conf.Bootstrap, logger log.Logger, grpcSrv *kgrpc.Server, registrar registry.Registrar) *kratos.App {
	options := []kratos.Option{
		kratos.Name(serviceName(cfg)),
		kratos.Logger(logger),
		kratos.Server(grpcSrv),
	}
	if registrar != nil {
		options = append(options, kratos.Registrar(registrar))
	}
	return kratos.New(options...)
}

func wireApp(cfg *conf.Bootstrap, logger log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(newResources, newRegistrar, biz.NewUsecases, service.NewServices, newGRPCServer, newApp))
}
