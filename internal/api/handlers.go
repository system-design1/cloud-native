package api

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"go-backend-service/internal/middleware"
	apperrors "go-backend-service/pkg/errors"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
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
