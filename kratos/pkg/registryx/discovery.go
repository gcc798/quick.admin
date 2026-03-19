package registryx

import (
	"fmt"
	"strings"
	"time"

	ketcd "github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/registry"
)

type DiscoveryConfig struct {
	Driver    string
	Service   string
	Namespace string
	Etcd      EtcdConfig
}

func NewDiscovery(cfg DiscoveryConfig) (registry.Discovery, func(), error) {
	driver := strings.ToLower(strings.TrimSpace(cfg.Driver))
	switch driver {
	case "", "etcd":
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
		return ketcd.New(client, ketcd.Namespace(namespace)), func() { _ = client.Close() }, nil
	default:
		return nil, nil, fmt.Errorf("unsupported registry discovery driver: %s", driver)
	}
}
