package repository

import (
	"context"
	"testing"
	"time"

	"go-backend-service/internal/otp"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     getEnvOrDefault("REDIS_HOST", "127.0.0.1") + ":" + getEnvOrDefault("REDIS_PORT", "6379"),
		Password: getEnvOrDefault("REDIS_PASSWORD", ""),
		DB:       getEnvIntOrDefault("REDIS_DB", 0),
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		t.Skipf("Skipping test: redis not available: %v", err)
	}

	return client
}

func TestRedisOTPStoreSaveGetDelete(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	store := NewRedisOTPStore(client)
	ctx := context.Background()
	state := otp.OTPState{
		RequestID:    "request-save-get-delete",
		TenantID:     1001,
		Phone:        "+989120001001",
		CodeHash:     otp.HashCode("123456"),
		AttemptCount: 0,
		MaxAttempts:  3,
		CreatedAt:    time.Now().UTC().Round(0),
		ExpiresAt:    time.Now().UTC().Add(2 * time.Minute).Round(0),
	}

	defer client.Del(ctx, redisOTPKey(state.TenantID, state.Phone))

	err := store.Save(ctx, state, 2*time.Minute)
	require.NoError(t, err)

	got, err := store.Get(ctx, state.TenantID, state.Phone)
	require.NoError(t, err)
	assert.Equal(t, state.RequestID, got.RequestID)
	assert.Equal(t, state.TenantID, got.TenantID)
	assert.Equal(t, state.Phone, got.Phone)
	assert.Equal(t, state.CodeHash, got.CodeHash)
	assert.Equal(t, state.AttemptCount, got.AttemptCount)
	assert.Equal(t, state.MaxAttempts, got.MaxAttempts)
	assert.True(t, state.CreatedAt.Equal(got.CreatedAt))
	assert.True(t, state.ExpiresAt.Equal(got.ExpiresAt))

	err = store.Delete(ctx, state.TenantID, state.Phone)
	require.NoError(t, err)

	_, err = store.Get(ctx, state.TenantID, state.Phone)
	assert.ErrorIs(t, err, otp.ErrOTPNotFound)
}

func TestRedisOTPStoreGetMissing(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	store := NewRedisOTPStore(client)

	_, err := store.Get(context.Background(), 1002, "+989120001002")
	assert.ErrorIs(t, err, otp.ErrOTPNotFound)
}

func TestRedisOTPStoreGetMalformedValue(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	store := NewRedisOTPStore(client)
	ctx := context.Background()
	key := redisOTPKey(1003, "+989120001003")
	defer client.Del(ctx, key)

	err := client.HSet(ctx, key, map[string]interface{}{
		"request_id":    "request-malformed",
		"tenant_id":     "not-an-int",
		"phone":         "+989120001003",
		"code_hash":     otp.HashCode("123456"),
		"attempt_count": "0",
		"max_attempts":  "3",
		"created_at":    time.Now().UTC().Format(time.RFC3339Nano),
		"expires_at":    time.Now().UTC().Add(2 * time.Minute).Format(time.RFC3339Nano),
	}).Err()
	require.NoError(t, err)

	_, err = store.Get(ctx, 1003, "+989120001003")
	require.Error(t, err)
	assert.NotErrorIs(t, err, otp.ErrOTPNotFound)
}

func TestRedisOTPStoreIncrementAttempts(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	store := NewRedisOTPStore(client)
	ctx := context.Background()
	state := otp.OTPState{
		RequestID:    "request-increment-attempts",
		TenantID:     1004,
		Phone:        "+989120001004",
		CodeHash:     otp.HashCode("123456"),
		AttemptCount: 0,
		MaxAttempts:  3,
		CreatedAt:    time.Now().UTC().Round(0),
		ExpiresAt:    time.Now().UTC().Add(2 * time.Minute).Round(0),
	}
	defer client.Del(ctx, redisOTPKey(state.TenantID, state.Phone))

	err := store.Save(ctx, state, 2*time.Minute)
	require.NoError(t, err)

	attempts, err := store.IncrementAttempts(ctx, state.TenantID, state.Phone)
	require.NoError(t, err)
	assert.Equal(t, 1, attempts)

	attempts, err = store.IncrementAttempts(ctx, state.TenantID, state.Phone)
	require.NoError(t, err)
	assert.Equal(t, 2, attempts)

	got, err := store.Get(ctx, state.TenantID, state.Phone)
	require.NoError(t, err)
	assert.Equal(t, 2, got.AttemptCount)
}

func TestRedisOTPStoreIncrementAttemptsMissingKey(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	store := NewRedisOTPStore(client)

	attempts, err := store.IncrementAttempts(context.Background(), 1005, "+989120001005")

	assert.Equal(t, 0, attempts)
	assert.ErrorIs(t, err, otp.ErrOTPNotFound)
}

func TestRedisOTPStoreIncrementAttemptsMissingField(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	store := NewRedisOTPStore(client)
	ctx := context.Background()
	key := redisOTPKey(1006, "+989120001006")
	defer client.Del(ctx, key)

	err := client.HSet(ctx, key, map[string]interface{}{
		"request_id": "request-missing-attempt-count",
		"tenant_id":  "1006",
		"phone":      "+989120001006",
		"code_hash":  otp.HashCode("123456"),
	}).Err()
	require.NoError(t, err)

	attempts, err := store.IncrementAttempts(ctx, 1006, "+989120001006")

	assert.Equal(t, 0, attempts)
	require.Error(t, err)
	assert.NotErrorIs(t, err, otp.ErrOTPNotFound)
}

func TestRedisOTPStoreIncrementAttemptsNonIntegerField(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	store := NewRedisOTPStore(client)
	ctx := context.Background()
	key := redisOTPKey(1007, "+989120001007")
	defer client.Del(ctx, key)

	err := client.HSet(ctx, key, map[string]interface{}{
		"request_id":    "request-non-integer-attempt-count",
		"tenant_id":     "1007",
		"phone":         "+989120001007",
		"code_hash":     otp.HashCode("123456"),
		"attempt_count": "not-an-int",
	}).Err()
	require.NoError(t, err)

	attempts, err := store.IncrementAttempts(ctx, 1007, "+989120001007")

	assert.Equal(t, 0, attempts)
	require.Error(t, err)
	assert.NotErrorIs(t, err, otp.ErrOTPNotFound)
}
