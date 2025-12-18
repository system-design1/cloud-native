package middleware

import (
	"strconv"
	"time"

	"go-backend-service/internal/metrics"

	"github.com/gin-gonic/gin"
)

// PrometheusMiddleware collects Prometheus metrics for HTTP requests
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip metrics endpoint to avoid infinite loops
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		method := c.Request.Method

		// Record request size if available
		if c.Request.ContentLength > 0 {
			metrics.HTTPRequestSize.WithLabelValues(method, path).Observe(float64(c.Request.ContentLength))
		}

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get status code
		statusCode := strconv.Itoa(c.Writer.Status())

		// Record metrics
		metrics.HTTPRequestDuration.WithLabelValues(method, path, statusCode).Observe(duration)
		metrics.HTTPRequestTotal.WithLabelValues(method, path, statusCode).Inc()

		// Record errors (4xx and 5xx)
		if c.Writer.Status() >= 400 {
			metrics.HTTPRequestErrors.WithLabelValues(method, path, statusCode).Inc()
		}

		// Record response size if available
		if c.Writer.Size() > 0 {
			metrics.HTTPResponseSize.WithLabelValues(method, path, statusCode).Observe(float64(c.Writer.Size()))
		}
	}
}

