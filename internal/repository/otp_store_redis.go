package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-backend-service/internal/otp"

	"github.com/redis/go-redis/v9"
)

// RedisOTPStore stores short-lived OTP verification state in Redis hashes.
type RedisOTPStore struct {
	client *redis.Client
}

// NewRedisOTPStore creates a Redis-backed OTP store.
func NewRedisOTPStore(client *redis.Client) *RedisOTPStore {
	return &RedisOTPStore{client: client}
}

// Save stores OTP verification state with a Redis TTL.
func (s *RedisOTPStore) Save(ctx context.Context, state otp.OTPState, ttl time.Duration) error {
	if ttl <= 0 {
		return fmt.Errorf("redis otp store save: ttl must be positive")
	}

	key := redisOTPKey(state.TenantID, state.Phone)
	fields := map[string]interface{}{
		"request_id":    state.RequestID,
		"tenant_id":     strconv.FormatInt(state.TenantID, 10),
		"phone":         state.Phone,
		"code_hash":     state.CodeHash,
		"attempt_count": strconv.Itoa(state.AttemptCount),
		"max_attempts":  strconv.Itoa(state.MaxAttempts),
		"created_at":    state.CreatedAt.Format(time.RFC3339Nano),
		"expires_at":    state.ExpiresAt.Format(time.RFC3339Nano),
	}

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, key, fields)
	pipe.Expire(ctx, key, ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis otp store save: %w", err)
	}

	return nil
}

// Get retrieves OTP verification state from Redis.
func (s *RedisOTPStore) Get(ctx context.Context, tenantID int64, phone string) (*otp.OTPState, error) {
	key := redisOTPKey(tenantID, phone)
	values, err := s.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("redis otp store get: %w", err)
	}
	if len(values) == 0 {
		return nil, otp.ErrOTPNotFound
	}

	state, err := parseOTPState(values)
	if err != nil {
		return nil, fmt.Errorf("redis otp store get: %w", err)
	}

	return state, nil
}

// IncrementAttempts is intentionally left for the next phase, where it will be atomic.
func (s *RedisOTPStore) IncrementAttempts(ctx context.Context, tenantID int64, phone string) (int, error) {
	return 0, otp.ErrNotImplemented
}

// Delete removes OTP verification state from Redis.
func (s *RedisOTPStore) Delete(ctx context.Context, tenantID int64, phone string) error {
	key := redisOTPKey(tenantID, phone)
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis otp store delete: %w", err)
	}
	return nil
}

func redisOTPKey(tenantID int64, phone string) string {
	return fmt.Sprintf("otp:%d:%s", tenantID, phone)
}

func parseOTPState(values map[string]string) (*otp.OTPState, error) {
	requestID, err := parseStringField(values, "request_id")
	if err != nil {
		return nil, err
	}
	tenantID, err := parseInt64Field(values, "tenant_id")
	if err != nil {
		return nil, err
	}
	phone, err := parseStringField(values, "phone")
	if err != nil {
		return nil, err
	}
	codeHash, err := parseStringField(values, "code_hash")
	if err != nil {
		return nil, err
	}
	attemptCount, err := parseIntField(values, "attempt_count")
	if err != nil {
		return nil, err
	}
	maxAttempts, err := parseIntField(values, "max_attempts")
	if err != nil {
		return nil, err
	}
	createdAt, err := parseTimeField(values, "created_at")
	if err != nil {
		return nil, err
	}
	expiresAt, err := parseTimeField(values, "expires_at")
	if err != nil {
		return nil, err
	}

	return &otp.OTPState{
		RequestID:    requestID,
		TenantID:     tenantID,
		Phone:        phone,
		CodeHash:     codeHash,
		AttemptCount: attemptCount,
		MaxAttempts:  maxAttempts,
		CreatedAt:    createdAt,
		ExpiresAt:    expiresAt,
	}, nil
}

func parseStringField(values map[string]string, field string) (string, error) {
	value, ok := values[field]
	if !ok {
		return "", fmt.Errorf("missing field %q", field)
	}
	if value == "" {
		return "", fmt.Errorf("empty field %q", field)
	}

	return value, nil
}

func parseInt64Field(values map[string]string, field string) (int64, error) {
	value, ok := values[field]
	if !ok {
		return 0, fmt.Errorf("missing field %q", field)
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid field %q: %w", field, err)
	}

	return parsed, nil
}

func parseIntField(values map[string]string, field string) (int, error) {
	value, ok := values[field]
	if !ok {
		return 0, fmt.Errorf("missing field %q", field)
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid field %q: %w", field, err)
	}

	return parsed, nil
}

func parseTimeField(values map[string]string, field string) (time.Time, error) {
	value, ok := values[field]
	if !ok {
		return time.Time{}, fmt.Errorf("missing field %q", field)
	}

	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid field %q: %w", field, err)
	}

	return parsed, nil
}
