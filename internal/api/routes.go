package api

import (
	"go-backend-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupMiddleware registers all middleware with the router
// Order matters: middleware are applied in the order they are registered
func SetupMiddleware(router *gin.Engine) {
	// 1. Recovery middleware (catches panics)
	router.Use(gin.Recovery())

	// 2. Correlation ID middleware (must be before logging)
	router.Use(middleware.CorrelationIDMiddleware())

	// 3. Request/Response logging middleware
	router.Use(middleware.RequestResponseLoggingMiddleware())

	// 4. Global error handler middleware (must be last)
	router.Use(middleware.ErrorHandlerMiddleware())
}

// SetupRoutes registers all routes with the router
func SetupRoutes(router *gin.Engine) {
	// Health check route
	router.GET("/health", HealthHandler)

	// Hello routes
	router.GET("/hello", HelloHandler)
	router.GET("/delayed-hello", DelayedHelloHandler)

	// Test routes
	router.GET("/test-error", TestErrorHandler)
}
