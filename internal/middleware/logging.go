package middleware

import (
	"time"

	"go-backend-service/internal/logger"

	"github.com/gin-gonic/gin"
)

// RequestResponseLoggingMiddleware creates a middleware for logging HTTP requests and responses
func RequestResponseLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Get correlation ID from context (set by correlation ID middleware)
		correlationIDValue, exists := c.Get("correlation_id")
		var correlationID string
		if exists {
			correlationID = correlationIDValue.(string)
		} else {
			correlationID = ""
		}

		// Create logger with correlation ID (logger.Get already adds correlation_id to logs)
		log := logger.Get(correlationID)

		// Extract request details
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		ip := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Extract response details
		statusCode := c.Writer.Status()
		responseSize := c.Writer.Size()

		// Log request and response details in structured JSON format
		// Note: correlation_id is already included by logger.Get()
		log.Info().
			Str("method", method).
			Str("path", path).
			Str("query", query).
			Str("ip", ip).
			Str("user_agent", userAgent).
			Int("status_code", statusCode).
			Int("response_size", responseSize).
			Dur("latency_ms", latency).
			Msg("HTTP request/response")
	}
}

