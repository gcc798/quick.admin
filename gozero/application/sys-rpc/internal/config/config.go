package config

import "github.com/zeromicro/go-zero/zrpc"

type PostgresConf struct {
	Dsn string
}

type Config struct {
	zrpc.RpcServerConf
	Postgres PostgresConf
}
