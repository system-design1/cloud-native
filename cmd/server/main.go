package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-backend-service/internal/api"
	"go-backend-service/internal/config"
	"go-backend-service/internal/logger"
	"go-backend-service/internal/server"

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

	// Setup middleware
	api.SetupMiddleware(router)

	// Setup routes
	api.SetupRoutes(router)

	// Create and start server
	srv := server.New(cfg, router)
	if err := srv.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}
