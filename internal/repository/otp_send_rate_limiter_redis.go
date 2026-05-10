package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-backend-service/internal/otp"

	"github.com/redis/go-redis/v9"
)

// RedisOTPSendRateLimiter limits OTP send attempts with a Redis fixed window.
type RedisOTPSendRateLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

var otpSendRateLimitScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 or redis.call("PTTL", KEYS[1]) < 0 then
	redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return current
`)

// NewRedisOTPSendRateLimiter creates a Redis-backed OTP send rate limiter.
func NewRedisOTPSendRateLimiter(client *redis.Client, limit int, window time.Duration) *RedisOTPSendRateLimiter {
	return &RedisOTPSendRateLimiter{
		client: client,
		limit:  limit,
		window: window,
	}
}

// AllowSend returns nil when an OTP send is allowed, or otp.ErrOTPRateLimited when the limit is exceeded.
func (l *RedisOTPSendRateLimiter) AllowSend(ctx context.Context, tenantID int64, phone string) error {
	if l.client == nil {
		return fmt.Errorf("redis otp send rate limiter: client is nil")
	}
	if l.limit <= 0 {
		return fmt.Errorf("redis otp send rate limiter: limit must be positive")
	}
	if l.window <= 0 {
		return fmt.Errorf("redis otp send rate limiter: window must be positive")
	}

	key := redisOTPSendRateLimitKey(tenantID, phone)
	count, err := otpSendRateLimitScript.Run(ctx, l.client, []string{key}, strconv.FormatInt(l.window.Milliseconds(), 10)).Int()
	if err != nil {
		return fmt.Errorf("redis otp send rate limiter allow send: %w", err)
	}
	if count > l.limit {
		return otp.ErrOTPRateLimited
	}

	return nil
}

func redisOTPSendRateLimitKey(tenantID int64, phone string) string {
	return fmt.Sprintf("otp:rate:send:%d:%s", tenantID, phone)
}
