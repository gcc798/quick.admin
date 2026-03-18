package data

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/force-c/nai-tizi/kratos/app/sys-rpc/ent"
	"github.com/force-c/nai-tizi/kratos/app/sys-rpc/internal/conf"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

type Resources struct {
	Ent     *ent.Client
	Redis   *redis.Client
	Storage *StorageManager
	WeChat  *WeChatManager
	Auth    AuthPolicy
}

type AuthPolicy struct {
	AllowConcurrent bool
	ShareToken      bool
}

var idCounter atomic.Int64

func init() {
	idCounter.Store(time.Now().UnixNano())
}

func NewResources(dataCfg conf.Data, authCfg conf.Auth) (*Resources, error) {
	if dataCfg.Database.Driver == "" || dataCfg.Database.DSN == "" {
		return nil, errors.New("database config is required")
	}
	client, err := ent.Open(dataCfg.Database.Driver, dataCfg.Database.DSN)
	if err != nil {
		return nil, err
	}
	if dataCfg.Redis.Addr == "" {
		return nil, errors.New("redis config is required")
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     dataCfg.Redis.Addr,
		Password: dataCfg.Redis.Password,
		DB:       dataCfg.Redis.DB,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &Resources{
		Ent:     client,
		Redis:   redisClient,
		Storage: NewStorageManager(client),
		WeChat:  NewWeChatManager(dataCfg.WeChat),
		Auth: AuthPolicy{
			AllowConcurrent: authCfg.AllowConcurrent,
			ShareToken:      authCfg.ShareToken,
		},
	}, nil
}

func (r *Resources) Close() error {
	if r == nil {
		return nil
	}
	if r.Redis != nil {
		_ = r.Redis.Close()
	}
	if r.Ent == nil {
		return nil
	}
	return r.Ent.Close()
}

func (r *Resources) withTx(ctx context.Context, fn func(tx *ent.Tx) error) (err error) {
	tx, err := r.Ent.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	if err = fn(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("%w: rollback failed: %v", err, rollbackErr)
		}
		return err
	}
	return tx.Commit()
}

func nextID() int64 {
	return idCounter.Add(1)
}
