package redis

import "github.com/redis/go-redis/v9"

func NewRedis(addr, password string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})
}
