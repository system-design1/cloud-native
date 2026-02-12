package redis

import (
	"context"
	"fmt"
	"time"

	cfg "go-backend-service/internal/config"

	"github.com/redis/go-redis/v9"
)

// NewClient creates a Redis client with pool settings from config and verifies connectivity with a ping.
func NewClient(c cfg.RedisConfig) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     c.Password,
		DB:           c.DB,
		PoolSize:     c.PoolSize,
		MinIdleConns: c.MinIdleConns,
		DialTimeout:  c.DialTimeout,
		ReadTimeout:  c.ReadTimeout,
		WriteTimeout: c.WriteTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return client, nil
}
