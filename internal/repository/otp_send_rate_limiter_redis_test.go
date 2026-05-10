package repository

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"go-backend-service/internal/otp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisOTPSendRateLimiterAllowsUnderLimit(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	ctx := context.Background()
	tenantID := int64(3001)
	phone := "+989123333001"
	key := redisOTPSendRateLimitKey(tenantID, phone)
	defer client.Del(ctx, key)
	require.NoError(t, client.Del(ctx, key).Err())

	limiter := NewRedisOTPSendRateLimiter(client, 2, time.Minute)

	require.NoError(t, limiter.AllowSend(ctx, tenantID, phone))
	require.NoError(t, limiter.AllowSend(ctx, tenantID, phone))
}

func TestRedisOTPSendRateLimiterBlocksAfterLimit(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	ctx := context.Background()
	tenantID := int64(3002)
	phone := "+989123333002"
	key := redisOTPSendRateLimitKey(tenantID, phone)
	defer client.Del(ctx, key)
	require.NoError(t, client.Del(ctx, key).Err())

	limiter := NewRedisOTPSendRateLimiter(client, 1, time.Minute)

	require.NoError(t, limiter.AllowSend(ctx, tenantID, phone))
	err := limiter.AllowSend(ctx, tenantID, phone)

	assert.ErrorIs(t, err, otp.ErrOTPRateLimited)
}

func TestRedisOTPSendRateLimiterSetsTTL(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	ctx := context.Background()
	tenantID := int64(3003)
	phone := "+989123333003"
	key := redisOTPSendRateLimitKey(tenantID, phone)
	defer client.Del(ctx, key)
	require.NoError(t, client.Del(ctx, key).Err())

	limiter := NewRedisOTPSendRateLimiter(client, 2, time.Minute)

	require.NoError(t, limiter.AllowSend(ctx, tenantID, phone))

	ttl, err := client.TTL(ctx, key).Result()
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0))
}

func TestRedisOTPSendRateLimiterIsolatesTenantIDs(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	ctx := context.Background()
	phone := "+989123333004"
	keyA := redisOTPSendRateLimitKey(3004, phone)
	keyB := redisOTPSendRateLimitKey(3005, phone)
	defer client.Del(ctx, keyA, keyB)
	require.NoError(t, client.Del(ctx, keyA, keyB).Err())

	limiter := NewRedisOTPSendRateLimiter(client, 1, time.Minute)

	require.NoError(t, limiter.AllowSend(ctx, 3004, phone))
	require.NoError(t, limiter.AllowSend(ctx, 3005, phone))
	assert.ErrorIs(t, limiter.AllowSend(ctx, 3004, phone), otp.ErrOTPRateLimited)
}

func TestRedisOTPSendRateLimiterIsolatesPhones(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	ctx := context.Background()
	tenantID := int64(3006)
	phoneA := "+989123333006"
	phoneB := "+989123333007"
	keyA := redisOTPSendRateLimitKey(tenantID, phoneA)
	keyB := redisOTPSendRateLimitKey(tenantID, phoneB)
	defer client.Del(ctx, keyA, keyB)
	require.NoError(t, client.Del(ctx, keyA, keyB).Err())

	limiter := NewRedisOTPSendRateLimiter(client, 1, time.Minute)

	require.NoError(t, limiter.AllowSend(ctx, tenantID, phoneA))
	require.NoError(t, limiter.AllowSend(ctx, tenantID, phoneB))
	assert.ErrorIs(t, limiter.AllowSend(ctx, tenantID, phoneA), otp.ErrOTPRateLimited)
}

func TestRedisOTPSendRateLimiterInvalidConfig(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	tests := []struct {
		name    string
		limiter *RedisOTPSendRateLimiter
	}{
		{name: "invalid limit", limiter: NewRedisOTPSendRateLimiter(client, 0, time.Minute)},
		{name: "invalid window", limiter: NewRedisOTPSendRateLimiter(client, 1, 0)},
		{name: "nil client", limiter: NewRedisOTPSendRateLimiter(nil, 1, time.Minute)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.limiter.AllowSend(context.Background(), 3007, "+989123333008")

			require.Error(t, err)
			assert.False(t, errors.Is(err, otp.ErrOTPRateLimited))
		})
	}
}

func TestRedisOTPSendRateLimiterRepairsMissingTTL(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	ctx := context.Background()
	tenantID := int64(3008)
	phone := "+989123333009"
	key := redisOTPSendRateLimitKey(tenantID, phone)
	defer client.Del(ctx, key)
	require.NoError(t, client.Del(ctx, key).Err())
	require.NoError(t, client.Set(ctx, key, "1", 0).Err())

	limiter := NewRedisOTPSendRateLimiter(client, 3, time.Minute)

	require.NoError(t, limiter.AllowSend(ctx, tenantID, phone))

	ttl, err := client.TTL(ctx, key).Result()
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0))
}

func TestRedisOTPSendRateLimiterConcurrentCalls(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	ctx := context.Background()
	tenantID := int64(3009)
	phone := "+989123333010"
	key := redisOTPSendRateLimitKey(tenantID, phone)
	defer client.Del(ctx, key)
	require.NoError(t, client.Del(ctx, key).Err())

	limit := 3
	limiter := NewRedisOTPSendRateLimiter(client, limit, time.Minute)
	totalCalls := 10
	errs := make(chan error, totalCalls)

	var wg sync.WaitGroup
	for i := 0; i < totalCalls; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- limiter.AllowSend(ctx, tenantID, phone)
		}()
	}
	wg.Wait()
	close(errs)

	allowed := 0
	rateLimited := 0
	for err := range errs {
		switch {
		case err == nil:
			allowed++
		case errors.Is(err, otp.ErrOTPRateLimited):
			rateLimited++
		default:
			require.NoError(t, err)
		}
	}

	assert.Equal(t, limit, allowed)
	assert.Equal(t, totalCalls-limit, rateLimited)
}
