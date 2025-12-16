package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-backend-service/internal/config"
	"go-backend-service/internal/logger"
	"go-backend-service/internal/middleware"
	apperrors "go-backend-service/pkg/errors"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize logger
	logger.Init()
	log := logger.Get()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Set Gin mode based on environment
	env := os.Getenv("GIN_MODE")
	if env == "" {
		env = "debug"
	}
	gin.SetMode(env)

	// Create Gin router
	router := gin.New()

	// Middleware chain (order matters!)
	// 1. Recovery middleware (catches panics)
	router.Use(gin.Recovery())
	
	// 2. Correlation ID middleware (must be before logging)
	router.Use(middleware.CorrelationIDMiddleware())
	
	// 3. Request/Response logging middleware
	router.Use(middleware.RequestResponseLoggingMiddleware())
	
	// 4. Global error handler middleware (must be last)
	router.Use(middleware.ErrorHandlerMiddleware())

	// Health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Hello route
	router.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
		})
	})

	// Test error route (for testing error handling)
	router.GET("/test-error", func(c *gin.Context) {
		middleware.ErrorHandler(c, apperrors.ErrBadRequest("This is a test error"))
		c.Abort()
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Info().
			Str("host", cfg.Server.Host).
			Int("port", cfg.Server.Port).
			Msg("Starting HTTP server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}
