package api

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"go-backend-service/internal/lifecycle"
	"go-backend-service/internal/middleware"
	apperrors "go-backend-service/pkg/errors"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests (backward compatibility)
// Returns OK if application is ready, Service Unavailable if shutting down
// Supports both GET and HEAD methods for health checks
func HealthHandler(lifecycleMgr *lifecycle.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		state := lifecycleMgr.GetState()
		if state == lifecycle.StateReady {
			if c.Request.Method == "HEAD" {
				c.Status(http.StatusOK)
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"state":  state.String(),
			})
		} else {
			if c.Request.Method == "HEAD" {
				c.Status(http.StatusServiceUnavailable)
				return
			}
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unavailable",
				"state":  state.String(),
			})
		}
	}
}

// ReadinessHandler handles readiness probe requests
// Returns OK only when application is ready to serve requests
func ReadinessHandler(lifecycleMgr *lifecycle.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if lifecycleMgr.IsReady() {
			c.JSON(http.StatusOK, gin.H{
				"status": "ready",
				"state":  lifecycleMgr.GetState().String(),
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not_ready",
				"state":  lifecycleMgr.GetState().String(),
			})
		}
	}
}

// LivenessHandler handles liveness probe requests
// Returns OK as long as application is not completely shut down
// Supports both GET and HEAD methods for health checks
func LivenessHandler(lifecycleMgr *lifecycle.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		state := lifecycleMgr.GetState()
		if state == lifecycle.StateShutdown {
			if c.Request.Method == "HEAD" {
				c.Status(http.StatusServiceUnavailable)
				return
			}
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "shutdown",
				"state":  state.String(),
			})
		} else {
			if c.Request.Method == "HEAD" {
				c.Status(http.StatusOK)
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"status": "alive",
				"state":  state.String(),
			})
		}
	}
}

// HelloHandler handles hello requests
func HelloHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, World!",
	})
}

// DelayedHelloHandler handles delayed hello requests with random delay between 1 and 3 seconds
func DelayedHelloHandler(c *gin.Context) {
	// Generate random delay between 1 and 3 seconds (1000-3000 milliseconds)
	// Using rand.Intn(2000) + 1000 to get range [1000, 2999] milliseconds
	delayMs := rand.Intn(2000) + 1000
	delayDuration := time.Duration(delayMs) * time.Millisecond
	delaySeconds := float64(delayMs) / 1000.0

	// Wait for the random delay
	time.Sleep(delayDuration)

	// Respond with message including the delay time
	// Logging is automatically handled by RequestResponseLoggingMiddleware
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Hello after delay: %.2f seconds", delaySeconds),
	})
}

// TestErrorHandler handles test error requests (for testing error handling)
func TestErrorHandler(c *gin.Context) {
	middleware.ErrorHandler(c, apperrors.ErrBadRequest("This is a test error"))
	c.Abort()
}
