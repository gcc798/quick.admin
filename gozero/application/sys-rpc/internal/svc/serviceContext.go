package svc

import (
	"github.com/force-c/nai-tizi/application/sys-rpc/internal/config"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	_ "github.com/lib/pq"
)

type ServiceContext struct {
	Config config.Config
	DB     sqlx.SqlConn
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := sqlx.NewSqlConn("postgres", c.Postgres.Dsn)

	return &ServiceContext{
		Config: c,
		DB:     db,
	}
}
