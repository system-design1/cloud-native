package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go-backend-service/internal/middleware"
	"go-backend-service/internal/repository"
	apperrors "go-backend-service/pkg/errors"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

const mongoBenchmarkTTL = 120 * time.Second

// mongoOTPValue matches Redis benchmark stored value format for identical GET response.
type mongoOTPValue struct {
	TenantID    string `json:"tenant_id"`
	PhoneNumber string `json:"phone_number"`
	OTPCode     string `json:"otp_code"`
}

// MongoSetBenchmarkHandler handles POST /v1/mongo/set (benchmark: set key/value with TTL, key=otp:tenant:phone).
// Stores value as JSON string matching Redis format: {"tenant_id":"...","phone_number":"...","otp_code":"123456"}
func MongoSetBenchmarkHandler(repo *repository.MongoBenchmarkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenant := c.Query("tenant")
		phone := c.Query("phone")
		valueParam := c.Query("value")
		if tenant == "" {
			middleware.ErrorHandler(c, apperrors.ErrBadRequest("tenant is required"))
			return
		}
		if phone == "" {
			middleware.ErrorHandler(c, apperrors.ErrBadRequest("phone is required"))
			return
		}

		var value string
		if valueParam != "" {
			value = valueParam
		} else {
			val := mongoOTPValue{TenantID: tenant, PhoneNumber: phone, OTPCode: "123456"}
			data, err := json.Marshal(val)
			if err != nil {
				middleware.ErrorHandler(c, apperrors.ErrInternalServerError("failed to marshal value"))
				return
			}
			value = string(data)
		}

		ttl := mongoBenchmarkTTL
		if ttlStr := c.Query("ttl"); ttlStr != "" {
			if d, err := time.ParseDuration(ttlStr); err == nil && d > 0 {
				ttl = d
			}
		}

		key := fmt.Sprintf("otp:%s:%s", tenant, phone)
		ctx := c.Request.Context()
		if err := repo.SetBenchmarkKey(ctx, key, value, ttl); err != nil {
			middleware.ErrorHandler(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// MongoGetBenchmarkHandler handles GET /v1/mongo/get (benchmark: get key, simulates Redis TTL).
func MongoGetBenchmarkHandler(repo *repository.MongoBenchmarkRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenant := c.Query("tenant")
		phone := c.Query("phone")
		if tenant == "" {
			middleware.ErrorHandler(c, apperrors.ErrBadRequest("tenant is required"))
			return
		}
		if phone == "" {
			middleware.ErrorHandler(c, apperrors.ErrBadRequest("phone is required"))
			return
		}

		key := fmt.Sprintf("otp:%s:%s", tenant, phone)
		ctx := c.Request.Context()
		val, expiresAt, err := repo.GetBenchmarkKey(ctx, key)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				c.JSON(http.StatusOK, gin.H{"found": false})
				return
			}
			middleware.ErrorHandler(c, err)
			return
		}

		if time.Now().After(expiresAt) {
			_ = repo.DeleteBenchmarkKey(ctx, key)
			c.JSON(http.StatusOK, gin.H{"found": false})
			return
		}

		// Normalize to Redis format: value must be JSON string {"tenant_id":"...","phone_number":"...","otp_code":"..."}
		value := val
		var parsed mongoOTPValue
		if json.Unmarshal([]byte(val), &parsed) != nil || parsed.TenantID == "" || parsed.PhoneNumber == "" {
			parsed = mongoOTPValue{TenantID: tenant, PhoneNumber: phone, OTPCode: val}
			data, _ := json.Marshal(parsed)
			value = string(data)
		}

		c.JSON(http.StatusOK, gin.H{"found": true, "value": value})
	}
}
