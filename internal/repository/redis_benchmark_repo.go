package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisBenchmarkRepository provides Redis SET/GET operations for benchmarking.
type RedisBenchmarkRepository struct {
	client *redis.Client
}

// NewRedisBenchmarkRepository creates a new RedisBenchmarkRepository.
func NewRedisBenchmarkRepository(client *redis.Client) *RedisBenchmarkRepository {
	return &RedisBenchmarkRepository{client: client}
}

// SetBenchmarkKey sets a key with value and TTL.
func (r *RedisBenchmarkRepository) SetBenchmarkKey(ctx context.Context, key string, value string, ttl time.Duration) error {
	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("set benchmark key: %w", err)
	}
	return nil
}

// GetBenchmarkKey retrieves a key's value. Returns redis.Nil if key does not exist.
func (r *RedisBenchmarkRepository) GetBenchmarkKey(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("get benchmark key: %w", err)
	}
	return val, nil
}
