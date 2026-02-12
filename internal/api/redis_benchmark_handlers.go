package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"go-backend-service/internal/middleware"
	"go-backend-service/internal/repository"
	apperrors "go-backend-service/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const otpTTL = 120 * time.Second

// Redis OTP value stored in Redis for benchmark.
type redisOTPValue struct {
	TenantID    string `json:"tenant_id"`
	PhoneNumber string `json:"phone_number"`
	OTPCode     string `json:"otp_code"`
}

// RedisOTPSetHandler handles POST /v1/redis/otp/set (benchmark: set OTP in Redis with fixed code and TTL).
func RedisOTPSetHandler(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.Query("tenant_id")
		phoneNumber := c.Query("phone_number")
		if tenantID == "" {
			middleware.ErrorHandler(c, apperrors.ErrBadRequest("tenant_id is required"))
			return
		}
		if phoneNumber == "" {
			middleware.ErrorHandler(c, apperrors.ErrBadRequest("phone_number is required"))
			return
		}

		key := "otp:" + tenantID + ":" + phoneNumber
		val := redisOTPValue{TenantID: tenantID, PhoneNumber: phoneNumber, OTPCode: "123456"}
		data, err := json.Marshal(val)
		if err != nil {
			middleware.ErrorHandler(c, apperrors.ErrInternalServerError("failed to marshal OTP value"))
			return
		}

		ctx := c.Request.Context()
		if err := rdb.Set(ctx, key, data, otpTTL).Err(); err != nil {
			middleware.ErrorHandler(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// RedisOTPGetHandler handles GET /v1/redis/otp/get (benchmark: get OTP from Redis by tenant_id and phone_number).
func RedisOTPGetHandler(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.Query("tenant_id")
		phoneNumber := c.Query("phone_number")
		if tenantID == "" {
			middleware.ErrorHandler(c, apperrors.ErrBadRequest("tenant_id is required"))
			return
		}
		if phoneNumber == "" {
			middleware.ErrorHandler(c, apperrors.ErrBadRequest("phone_number is required"))
			return
		}

		key := "otp:" + tenantID + ":" + phoneNumber
		ctx := c.Request.Context()
		data, err := rdb.Get(ctx, key).Bytes()
		if err != nil {
			if err == redis.Nil {
				c.JSON(http.StatusOK, gin.H{"found": false})
				return
			}
			middleware.ErrorHandler(c, err)
			return
		}

		var val redisOTPValue
		if err := json.Unmarshal(data, &val); err != nil {
			middleware.ErrorHandler(c, apperrors.ErrInternalServerError("failed to unmarshal OTP value"))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"found":        true,
			"tenant_id":    val.TenantID,
			"phone_number": val.PhoneNumber,
			"otp_code":     val.OTPCode,
		})
	}
}

const benchmarkTTL = 120 * time.Second

// RedisSetBenchmarkHandler handles POST /v1/redis/set (generic benchmark: set key/value with TTL).
func RedisSetBenchmarkHandler(repo *repository.RedisBenchmarkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Query("key")
		value := c.Query("value")
		if key == "" {
			middleware.ErrorHandler(c, apperrors.ErrBadRequest("key is required"))
			return
		}
		if value == "" {
			middleware.ErrorHandler(c, apperrors.ErrBadRequest("value is required"))
			return
		}

		ttl := benchmarkTTL
		if ttlStr := c.Query("ttl"); ttlStr != "" {
			if d, err := time.ParseDuration(ttlStr); err == nil && d > 0 {
				ttl = d
			}
		}

		ctx := c.Request.Context()
		if err := repo.SetBenchmarkKey(ctx, key, value, ttl); err != nil {
			middleware.ErrorHandler(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// RedisGetBenchmarkHandler handles GET /v1/redis/get (generic benchmark: get key).
func RedisGetBenchmarkHandler(repo *repository.RedisBenchmarkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Query("key")
		if key == "" {
			middleware.ErrorHandler(c, apperrors.ErrBadRequest("key is required"))
			return
		}

		ctx := c.Request.Context()
		val, err := repo.GetBenchmarkKey(ctx, key)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				c.JSON(http.StatusOK, gin.H{"found": false})
				return
			}
			middleware.ErrorHandler(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"found": true, "value": val})

		// If you want to parse the value as json, you can do it like this:

		// var parsed redisOTPValue
		// if err := json.Unmarshal([]byte(val), &parsed); err != nil {
		// 	middleware.ErrorHandler(c, apperrors.ErrInternalServerError("invalid stored json"))
		// 	return
		// }

		// c.JSON(http.StatusOK, gin.H{
		// 	"found":        true,
		// 	"tenant_id":   parsed.TenantID,
		// 	"phone_number": parsed.PhoneNumber,
		// 	"otp_code":    parsed.OTPCode,
		// })
	}
}
