package api

import (
	"go-backend-service/internal/lifecycle"
	"go-backend-service/internal/middleware"
	"go-backend-service/internal/repository"

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
func SetupRoutes(router *gin.Engine, lifecycleMgr *lifecycle.Manager, tenantSettingsRepo *repository.TenantSettingsRepository, tenantSettingsInsertRepo *repository.TenantSettingsInsertRepository) {
	// Prometheus metrics endpoint (must be before other routes to avoid middleware)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check routes (liveness and readiness)
	// Support both GET and HEAD methods for health checks
	router.GET("/health", HealthHandler(lifecycleMgr))
	router.HEAD("/health", HealthHandler(lifecycleMgr))
	router.GET("/ready", ReadinessHandler(lifecycleMgr))
	router.HEAD("/ready", ReadinessHandler(lifecycleMgr))
	router.GET("/live", LivenessHandler(lifecycleMgr))
	router.HEAD("/live", LivenessHandler(lifecycleMgr))

	// Hello routes
	router.GET("/hello", HelloHandler)
	router.GET("/delayed-hello", DelayedHelloHandler)
	router.GET("/child-hello", ChildHelloHandler)

	// Test routes
	router.GET("/test-error", TestErrorHandler)

	// Versioned API routes
	v1 := router.Group("/v1")
	{
		// OTP routes
		otp := v1.Group("/otp")
		{
			otp.POST("/code", GenerateOTPCodeHandler)
			// Tenant settings routes
			otp.GET("/tenant-settings/:id", GetTenantSettingsByIDHandler(tenantSettingsRepo))
			otp.POST("/tenant-settings-insert-benchmark", InsertTenantSettingsBenchmarkHandler(tenantSettingsInsertRepo))
		}
	}
}
