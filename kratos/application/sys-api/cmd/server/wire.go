//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/force-c/nai-tizi/kratos/application/sys-api/internal/biz"
	"github.com/force-c/nai-tizi/kratos/application/sys-api/internal/conf"
	"github.com/force-c/nai-tizi/kratos/application/sys-api/internal/data"
	apisrv "github.com/force-c/nai-tizi/kratos/application/sys-api/internal/server"
	"github.com/force-c/nai-tizi/kratos/application/sys-api/internal/service"
	kratos "github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

func newRepositories(cfg *conf.Bootstrap) (*data.Repositories, func(), error) {
	repos, err := data.NewRepositories(cfg.Server.RPC)
	if err != nil {
		return nil, nil, err
	}
	return repos, func() {
		_ = repos.Close()
	}, nil
}

func newWebSocketHub() (*apisrv.WebSocketHub, func(), error) {
	hub := apisrv.NewWebSocketHub()
	return hub, func() {
		hub.Stop()
	}, nil
}

func newGatewayDeps(repos *data.Repositories, usecases *biz.Usecases, wsHub *apisrv.WebSocketHub) *apisrv.GatewayDeps {
	return &apisrv.GatewayDeps{Auth: repos.Auth, Attachment: usecases.Attachment, OperLog: repos.OperLog, WebSocket: wsHub}
}

func newHTTPServer(cfg *conf.Bootstrap, deps *apisrv.GatewayDeps, services *service.Services) *khttp.Server {
	return apisrv.NewHTTPServer(
		cfg.Server.HTTP,
		deps,
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

func newApp(logger log.Logger, httpSrv *khttp.Server) *kratos.App {
	return kratos.New(kratos.Name("sys-api"), kratos.Logger(logger), kratos.Server(httpSrv))
}

func wireApp(cfg *conf.Bootstrap, logger log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(newRepositories, biz.NewUsecases, service.NewServices, newWebSocketHub, newGatewayDeps, newHTTPServer, newApp))
}
