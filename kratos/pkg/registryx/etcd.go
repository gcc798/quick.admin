package registryx

import (
	"fmt"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdConfig struct {
	Endpoints []string
	Username  string
	Password  string
	Timeout   time.Duration
}

func NewEtcdClient(cfg EtcdConfig) (*clientv3.Client, error) {
	endpoints := make([]string, 0, len(cfg.Endpoints))
	for _, endpoint := range cfg.Endpoints {
		if value := strings.TrimSpace(endpoint); value != "" {
			endpoints = append(endpoints, value)
		}
	}
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("etcd endpoints are empty")
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		Username:    strings.TrimSpace(cfg.Username),
		Password:    strings.TrimSpace(cfg.Password),
		DialTimeout: timeout,
	})
}
