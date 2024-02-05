package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr string, readTimeout, writeTimeout int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &RedisCache{
		client: client,
	}, nil
}

func (c *RedisCache) Get(key string) (string, error) {
	val, err := c.client.Get(context.Background(), key).Result()
	if err != nil && err != redis.Nil {
		return "", fmt.Errorf("failed to get value from Redis: %v", err)
	}
	return val, nil
}

func (c *RedisCache) Set(key, value string, expiration time.Duration) error {
	err := c.client.Set(context.Background(), key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set value in Redis: %v", err)
	}
	return nil
}
