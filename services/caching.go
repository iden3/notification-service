package services

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisStatus = "PONG"

// RedisCache for implementation of CacheService
type RedisCache struct {
	redisClient *redis.Client
}

// NewRedisCacheService for constructing *RedisCacheService
func NewRedisCacheService(client *redis.Client) *RedisCache {
	return &RedisCache{
		redisClient: client,
	}
}

// Ping check redis status.
func (r RedisCache) Ping(ctx context.Context) error {
	// be default for PING request Redis should response "PONG"
	// https://redis.io/commands/ping/
	res, err := r.redisClient.Ping(ctx).Result()
	if err != nil {
		return err
	}
	if res != redisStatus {
		return errors.New("unknown status from redis")
	}
	return nil
}

// Get for retrieving value from cache by key.
// If key doesn't exist in cache return nil, nil.
func (r RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	v, err := r.redisClient.Get(ctx, key).Result()

	// redis is working. But the value does not exist in cache.
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil { // redis return "real" error.
		return nil, err
	}

	return v, nil
}

// Set for put value to cache by specific key for some duration period
func (r RedisCache) Set(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	err := r.redisClient.Set(ctx, key, value, duration).Err()
	if err != nil {
		return err
	}
	return nil
}
