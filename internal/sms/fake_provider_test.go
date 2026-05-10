package sms

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"
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

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if result, err := strconv.Atoi(value); err == nil {
			return result
		}
	}
	return defaultValue
}

func TestFakeProviderSendOTPSuccess(t *testing.T) {
	provider := newFakeProviderWithDelay(0, 0)
	req := otp.SMSRequest{
		RequestID: "request-success",
		TenantID:  123,
		Phone:     "+989121234567",
		Code:      "123456",
	}

	result, err := provider.SendOTP(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, fakeProviderName, result.Provider)
	assert.Equal(t, otp.RequestStatusSent, result.Status)
	assert.Equal(t, "fake-"+req.RequestID, result.MessageID)
	assert.False(t, result.SentAt.IsZero())
	require.NotNil(t, result.RawResponse)
	assert.Equal(t, fakeProviderName, result.RawResponse["provider"])
	assert.Equal(t, true, result.RawResponse["simulated"])
	assert.Equal(t, req.RequestID, result.RawResponse["request_id"])
}

func TestFakeProviderRawResponseDoesNotExposeOTPCode(t *testing.T) {
	provider := newFakeProviderWithDelay(0, 0)
	req := otp.SMSRequest{
		RequestID: "request-safe-response",
		TenantID:  123,
		Phone:     "+989121234567",
		Code:      "654321",
	}

	result, err := provider.SendOTP(context.Background(), req)

	require.NoError(t, err)
	for key, value := range result.RawResponse {
		assert.NotEqual(t, "code", key)
		assert.NotEqual(t, req.Code, value)
	}
}

func TestFakeProviderUsesRequestProvider(t *testing.T) {
	provider := newFakeProviderWithDelay(0, 0)
	req := otp.SMSRequest{
		RequestID: "request-provider",
		Provider:  "kavenegar",
		Code:      "123456",
	}

	result, err := provider.SendOTP(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, req.Provider, result.Provider)
	assert.Equal(t, req.Provider, result.RawResponse["provider"])
}

func TestFakeProviderCanceledContext(t *testing.T) {
	provider := newFakeProviderWithDelay(10*time.Millisecond, 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := provider.SendOTP(ctx, otp.SMSRequest{Code: "123456"})

	require.Nil(t, result)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestFakeProviderTimeoutContext(t *testing.T) {
	provider := newFakeProviderWithDelay(50*time.Millisecond, 50*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	result, err := provider.SendOTP(ctx, otp.SMSRequest{Code: "123456"})

	require.Nil(t, result)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))
}

func TestNewFakeProviderDefaultLatencyRange(t *testing.T) {
	provider := NewFakeProvider()

	assert.Equal(t, 20*time.Millisecond, provider.minDelay)
	assert.Equal(t, 30*time.Millisecond, provider.maxDelay)
}

func TestNewFakeProviderWithDelay(t *testing.T) {
	provider := NewFakeProviderWithDelay(5*time.Millisecond, 10*time.Millisecond)

	assert.Equal(t, 5*time.Millisecond, provider.minDelay)
	assert.Equal(t, 10*time.Millisecond, provider.maxDelay)
}

func TestFakeProviderDebugCodeCaptureStoresCodeInRedis(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	provider := NewFakeProviderWithDebugCodeCapture(client, time.Minute)
	provider.minDelay = 0
	provider.maxDelay = 0

	req := otp.SMSRequest{
		RequestID: "request-debug-code",
		TenantID:  456,
		Phone:     "+989121234568",
		Code:      "112233",
		Provider:  "fake",
	}
	key := debugCodeKey(req.TenantID, req.Phone)
	ctx := context.Background()
	defer client.Del(ctx, key)
	require.NoError(t, client.Del(ctx, key).Err())

	result, err := provider.SendOTP(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotContains(t, result.RawResponse, "code")

	data, err := client.Get(ctx, key).Bytes()
	require.NoError(t, err)

	var stored debugCodeValue
	require.NoError(t, json.Unmarshal(data, &stored))
	assert.Equal(t, req.RequestID, stored.RequestID)
	assert.Equal(t, req.TenantID, stored.TenantID)
	assert.Equal(t, req.Phone, stored.Phone)
	assert.Equal(t, req.Code, stored.Code)
	assert.Equal(t, req.Provider, stored.Provider)
	assert.False(t, stored.CreatedAt.IsZero())

	ttl, err := client.TTL(ctx, key).Result()
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0))
}

func TestFakeProviderDebugCodeCaptureCanceledContextDoesNotWrite(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	provider := NewFakeProviderWithDebugCodeCapture(client, time.Minute)
	provider.minDelay = 50 * time.Millisecond
	provider.maxDelay = 50 * time.Millisecond

	req := otp.SMSRequest{
		RequestID: "request-debug-canceled",
		TenantID:  457,
		Phone:     "+989121234569",
		Code:      "445566",
	}
	key := debugCodeKey(req.TenantID, req.Phone)
	ctx := context.Background()
	defer client.Del(ctx, key)
	require.NoError(t, client.Del(ctx, key).Err())

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Millisecond)
	defer cancel()

	result, err := provider.SendOTP(timeoutCtx, req)

	require.Nil(t, result)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))

	exists, err := client.Exists(ctx, key).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), exists)
}
