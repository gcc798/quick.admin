package server

import (
	"context"
	"time"

	v1 "github.com/force-c/nai-tizi/kratos/api/system/v1"
	"github.com/force-c/nai-tizi/kratos/app/sys-api/internal/conf"
	"github.com/force-c/nai-tizi/kratos/app/sys-api/internal/service"
	selector "github.com/go-kratos/kratos/v2/middleware/selector"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

func NewHTTPServer(c conf.HTTP,
	deps *GatewayDeps,
	healthSvc *service.HealthServiceService,
	authSvc *service.AuthServiceService,
	captchaSvc *service.CaptchaServiceService,
	menuSvc *service.MenuServiceService,
	userSvc *service.UserServiceService,
	roleSvc *service.RoleServiceService,
	orgSvc *service.OrgServiceService,
	configSvc *service.ConfigServiceService,
	dictSvc *service.DictServiceService,
	loginLogSvc *service.LoginLogServiceService,
	operLogSvc *service.OperLogServiceService,
	storageSvc *service.StorageEnvServiceService,
	attachmentSvc *service.AttachmentServiceService,
) *khttp.Server {
	timeout := time.Second
	if c.Timeout != "" {
		if parsed, err := time.ParseDuration(c.Timeout); err == nil {
			timeout = parsed
		}
	}
	srv := khttp.NewServer(
		khttp.Network(c.Network),
		khttp.Address(c.Addr),
		khttp.Timeout(timeout),
		khttp.ResponseEncoder(responseEncoder),
		khttp.ErrorEncoder(errorEncoder),
		khttp.Middleware(
			metricsMiddleware(),
			selector.Server(authMiddleware(deps)).Match(func(ctx context.Context, operation string) bool {
				return !isPublicOperation(operation)
			}).Build(),
			selector.Server(permissionMiddleware(deps)).Match(func(ctx context.Context, operation string) bool {
				_, ok := permissionForOperation(operation)
				return ok
			}).Build(),
			operLogMiddleware(deps),
		),
	)
	registerAttachmentHTTPRoutes(srv, deps, attachmentSvc)
	registerMetricsEndpoint(srv)
	v1.RegisterHealthServiceHTTPServer(srv, healthSvc)
	v1.RegisterAuthServiceHTTPServer(srv, authSvc)
	v1.RegisterCaptchaServiceHTTPServer(srv, captchaSvc)
	v1.RegisterMenuServiceHTTPServer(srv, menuSvc)
	v1.RegisterUserServiceHTTPServer(srv, userSvc)
	v1.RegisterRoleServiceHTTPServer(srv, roleSvc)
	v1.RegisterOrgServiceHTTPServer(srv, orgSvc)
	v1.RegisterConfigServiceHTTPServer(srv, configSvc)
	v1.RegisterDictServiceHTTPServer(srv, dictSvc)
	v1.RegisterLoginLogServiceHTTPServer(srv, loginLogSvc)
	v1.RegisterOperLogServiceHTTPServer(srv, operLogSvc)
	v1.RegisterStorageEnvServiceHTTPServer(srv, storageSvc)
	registerHealthEndpoints(srv)
	registerSwaggerEndpoints(srv)
	registerWebSocketEndpoint(srv, deps)
	return srv
}
