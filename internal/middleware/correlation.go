package middleware

import (
	"go-backend-service/internal/logger"

	"github.com/gin-gonic/gin"
)

// CorrelationIDMiddleware adds correlation ID to requests for tracing
func CorrelationIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get correlation ID from header
		correlationID := c.GetHeader("X-Correlation-ID")

		// If not present, generate a new one
		if correlationID == "" {
			correlationID = logger.GenerateCorrelationID()
		}

		// Set correlation ID in context
		c.Set("correlation_id", correlationID)

		// Add correlation ID to response headers
		c.Header("X-Correlation-ID", correlationID)

		c.Next()
	}
}

