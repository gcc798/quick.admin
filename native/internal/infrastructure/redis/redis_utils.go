package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisUtils Redis工具类
type RedisUtils struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisUtils 创建Redis工具类实例
func NewRedisUtils(client *redis.Client) *RedisUtils {
	return &RedisUtils{
		client: client,
		ctx:    context.Background(),
	}
}

// SetCacheMap 设置Hash
func (r *RedisUtils) SetCacheMap(key string, data map[string]interface{}) error {
	// 将map转换为Redis Hash格式
	fields := make(map[string]interface{})
	for k, v := range data {
		// 将值转换为字符串
		switch val := v.(type) {
		case string:
			fields[k] = val
		default:
			// 其他类型转为JSON字符串
			jsonBytes, err := json.Marshal(val)
			if err != nil {
				return err
			}
			fields[k] = string(jsonBytes)
		}
	}
	return r.client.HSet(r.ctx, key, fields).Err()
}

// GetCacheMap 获取Hash
func (r *RedisUtils) GetCacheMap(key string) (map[string]interface{}, error) {
	result, err := r.client.HGetAll(r.ctx, key).Result()
	if err != nil {
		return nil, err
	}

	// 转换为map[string]interface{}
	data := make(map[string]interface{})
	for k, v := range result {
		data[k] = v
	}
	return data, nil
}

// Expire 设置过期时间
func (r *RedisUtils) Expire(key string, duration time.Duration) error {
	return r.client.Expire(r.ctx, key, duration).Err()
}

// ZAdd 添加到有序集合
func (r *RedisUtils) ZAdd(key string, score float64, member string) error {
	return r.client.ZAdd(r.ctx, key, redis.Z{
		Score:  score,
		Member: member,
	}).Err()
}

// ZRem 从有序集合中移除
func (r *RedisUtils) ZRem(key string, member string) (bool, error) {
	result, err := r.client.ZRem(r.ctx, key, member).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// ZRangeByScore 根据分数范围查询有序集合
func (r *RedisUtils) ZRangeByScore(key string, min, max float64, offset, count int64) ([]string, error) {
	return r.client.ZRangeByScore(r.ctx, key, &redis.ZRangeBy{
		Min:    fmt.Sprintf("%f", min),
		Max:    fmt.Sprintf("%f", max),
		Offset: offset,
		Count:  count,
	}).Result()
}

// TryLock 尝试获取分布式锁
func (r *RedisUtils) TryLock(key string, waitTime, leaseTime int64) (bool, error) {
	// 使用SET NX EX实现分布式锁
	result, err := r.client.SetNX(r.ctx, key, "1", time.Duration(leaseTime)*time.Millisecond).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}

// Unlock 释放分布式锁
func (r *RedisUtils) Unlock(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// DeleteObject 删除key
func (r *RedisUtils) DeleteObject(key string) error {
	return r.client.Del(r.ctx, key).Err()
}
