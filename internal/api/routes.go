package api

import (
	"go-backend-service/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupMiddleware registers all middleware with the router
// Order matters: middleware are applied in the order they are registered
func SetupMiddleware(router *gin.Engine) {
	// 1. Recovery middleware (catches panics)
	router.Use(gin.Recovery())

	// 2. Prometheus metrics middleware (must be early to capture all requests)
	router.Use(middleware.PrometheusMiddleware())

	// 3. OpenTelemetry tracing middleware (must be early to capture all spans)
	router.Use(middleware.TracingMiddleware())

	// 4. Correlation ID middleware (must be before tracing context and logging)
	router.Use(middleware.CorrelationIDMiddleware())

	// 5. Tracing context middleware (extracts trace/span IDs and adds to context)
	router.Use(middleware.TracingContextMiddleware())

	// 6. Request/Response logging middleware (includes trace/span IDs in logs)
	router.Use(middleware.RequestResponseLoggingMiddleware())

	// 7. Global error handler middleware (must be last)
	router.Use(middleware.ErrorHandlerMiddleware())
}

// SetupRoutes registers all routes with the router
func SetupRoutes(router *gin.Engine) {
	// Prometheus metrics endpoint (must be before other routes to avoid middleware)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check route
	router.GET("/health", HealthHandler)

	// Hello routes
	router.GET("/hello", HelloHandler)
	router.GET("/delayed-hello", DelayedHelloHandler)

	// Test routes
	router.GET("/test-error", TestErrorHandler)
}
