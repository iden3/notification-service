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

func (r RedisCache) Delete(ctx context.Context, keys ...string) error {
	_, err := r.redisClient.Del(ctx, keys...).Result()
	if err != nil {
		return err
	}
	return nil
}

// GetAllByPrefix get all values by prefix key.
func (r RedisCache) GetAllByPrefix(ctx context.Context, prefix string) (values []interface{}, keys []string, err error) {
	searchKey := buildSearchKey(prefix)

	keys, err = r.Scan(ctx, searchKey)
	if err != nil {
		return nil, nil, err
	}
	if len(keys) == 0 {
		return nil, nil, nil
	}

	values, err = r.redisClient.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, nil, err
	}

	return values, keys, nil
}

// Scan get all keys by prefix.
func (r RedisCache) Scan(ctx context.Context, prefix string) ([]string, error) {
	var (
		cursor uint64
		values []string
	)

	for {
		var keys []string
		var err error
		keys, cursor, err = r.redisClient.Scan(ctx, cursor, prefix, 10).Result()
		if err != nil {
			return nil, err
		}
		values = append(values, keys...)
		if cursor == 0 {
			break
		}
	}

	return values, nil
}

// Set for put value to cache by specific key for some duration period
func (r RedisCache) Set(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	err := r.redisClient.Set(ctx, key, value, duration).Err()
	if err != nil {
		return err
	}
	return nil
}
