package main

import (
	"strings"
	"time"

	"github.com/force-c/nai-tizi/kratos/application/sys-rpc/internal/conf"
	"github.com/force-c/nai-tizi/kratos/pkg/configx"
	"github.com/force-c/nai-tizi/kratos/pkg/registryx"
	"github.com/go-kratos/kratos/v2/registry"
)

func newRegistrar(cfg *conf.Bootstrap) (registry.Registrar, func(), error) {
	mode := strings.ToLower(strings.TrimSpace(cfg.Registry.Mode))
	if mode == "" || mode == "direct" || mode == "disabled" || mode == "off" {
		return nil, func() {}, nil
	}
	driver := strings.ToLower(strings.TrimSpace(cfg.Registry.Driver))
	if driver == "" {
		return nil, func() {}, nil
	}
	return registryx.NewRegistrar(registryx.RegistrarConfig{
		Driver:      driver,
		Namespace:   cfg.Registry.Namespace,
		RegisterTTL: configx.ParseDurationOrDefault(cfg.Registry.RegisterTTL, 15*time.Second),
		Etcd: registryx.EtcdConfig{
			Endpoints: cfg.Registry.Etcd.Endpoints,
			Username:  cfg.Registry.Etcd.Username,
			Password:  cfg.Registry.Etcd.Password,
			Timeout:   configx.ParseDurationOrDefault(cfg.Registry.Etcd.Timeout, 5*time.Second),
		},
	})
}

func serviceName(cfg *conf.Bootstrap) string {
	if value := strings.TrimSpace(cfg.Registry.Service); value != "" {
		return value
	}
	return "sys-rpc"
}
