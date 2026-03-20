package registryx

import (
	"fmt"
	"strings"
	"time"

	ketcd "github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/registry"
)

type RegistrarConfig struct {
	Driver      string
	Namespace   string
	RegisterTTL time.Duration
	Etcd        EtcdConfig
}

func NewRegistrar(cfg RegistrarConfig) (registry.Registrar, func(), error) {
	driver := strings.ToLower(strings.TrimSpace(cfg.Driver))
	if driver == "" {
		return nil, func() {}, nil
	}
	switch driver {
	case "etcd":
		timeout := cfg.Etcd.Timeout
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		client, err := NewEtcdClient(EtcdConfig{
			Endpoints: cfg.Etcd.Endpoints,
			Username:  cfg.Etcd.Username,
			Password:  cfg.Etcd.Password,
			Timeout:   timeout,
		})
		if err != nil {
			return nil, nil, err
		}
		namespace := strings.TrimSpace(cfg.Namespace)
		if namespace == "" {
			namespace = "/kratos"
		}
		registerTTL := cfg.RegisterTTL
		if registerTTL <= 0 {
			registerTTL = 15 * time.Second
		}
		return ketcd.New(client, ketcd.Namespace(namespace), ketcd.RegisterTTL(registerTTL)), func() {
			_ = client.Close()
		}, nil
	default:
		return nil, nil, fmt.Errorf("unsupported registry driver: %s", driver)
	}
}
