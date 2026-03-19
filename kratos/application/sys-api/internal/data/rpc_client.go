package data

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1 "github.com/force-c/nai-tizi/kratos/api/system/v1"
	"github.com/force-c/nai-tizi/kratos/application/sys-api/internal/conf"
	"github.com/force-c/nai-tizi/kratos/pkg/configx"
	"github.com/force-c/nai-tizi/kratos/pkg/grpcx"
	"github.com/force-c/nai-tizi/kratos/pkg/registryx"
	grpc "google.golang.org/grpc"
)

type RPCClientSet struct {
	conn       *grpc.ClientConn
	cleanup    func()
	Health     v1.HealthServiceClient
	Auth       v1.AuthServiceClient
	Captcha    v1.CaptchaServiceClient
	Menu       v1.MenuServiceClient
	User       v1.UserServiceClient
	Role       v1.RoleServiceClient
	Org        v1.OrgServiceClient
	Config     v1.ConfigServiceClient
	Dict       v1.DictServiceClient
	LoginLog   v1.LoginLogServiceClient
	OperLog    v1.OperLogServiceClient
	StorageEnv v1.StorageEnvServiceClient
	Attachment v1.AttachmentServiceClient
}

func NewRPCClientSet(cfg conf.RPC) (*RPCClientSet, error) {
	conn, cleanup, err := newRPCConn(cfg)
	if err != nil {
		return nil, err
	}
	return &RPCClientSet{
		conn:       conn,
		cleanup:    cleanup,
		Health:     v1.NewHealthServiceClient(conn),
		Auth:       v1.NewAuthServiceClient(conn),
		Captcha:    v1.NewCaptchaServiceClient(conn),
		Menu:       v1.NewMenuServiceClient(conn),
		User:       v1.NewUserServiceClient(conn),
		Role:       v1.NewRoleServiceClient(conn),
		Org:        v1.NewOrgServiceClient(conn),
		Config:     v1.NewConfigServiceClient(conn),
		Dict:       v1.NewDictServiceClient(conn),
		LoginLog:   v1.NewLoginLogServiceClient(conn),
		OperLog:    v1.NewOperLogServiceClient(conn),
		StorageEnv: v1.NewStorageEnvServiceClient(conn),
		Attachment: v1.NewAttachmentServiceClient(conn),
	}, nil
}

func (c *RPCClientSet) Close() error {
	if c == nil {
		return nil
	}
	if c.cleanup != nil {
		c.cleanup()
	}
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func newRPCConn(cfg conf.RPC) (*grpc.ClientConn, func(), error) {
	timeout := 2 * time.Second
	if strings.TrimSpace(cfg.Timeout) != "" {
		parsed, err := time.ParseDuration(strings.TrimSpace(cfg.Timeout))
		if err != nil {
			return nil, nil, fmt.Errorf("parse rpc timeout: %w", err)
		}
		timeout = parsed
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return grpcx.DialInsecure(ctx, grpcx.ClientConfig{
		Mode:     cfg.Mode,
		Endpoint: cfg.Endpoint,
		Timeout:  timeout,
		Discovery: registryx.DiscoveryConfig{
			Driver:    cfg.Discovery.Driver,
			Service:   cfg.Discovery.Service,
			Namespace: cfg.Discovery.Etcd.Namespace,
			Etcd: registryx.EtcdConfig{
				Endpoints: cfg.Discovery.Etcd.Endpoints,
				Username:  cfg.Discovery.Etcd.Username,
				Password:  cfg.Discovery.Etcd.Password,
				Timeout:   configx.ParseDurationOrDefault(cfg.Discovery.Etcd.Timeout, 5*time.Second),
			},
		},
	}, outgoingUserInterceptor)
}
