package svc

import (
	"github.com/force-c/nai-tizi/application/sys-rpc/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	_ "github.com/lib/pq"
)

type ServiceContext struct {
	Config config.Config
	DB     sqlx.SqlConn
	Redis  *redis.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := sqlx.NewSqlConn("postgres", c.Postgres.Dsn)
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.CacheRedis.Addr,
		Password: c.CacheRedis.Password,
		DB:       c.CacheRedis.Db,
	})

	return &ServiceContext{
		Config: c,
		DB:     db,
		Redis:  rdb,
	}
}
