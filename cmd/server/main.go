package main

import (
	"context"
	"fmt"
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
	// Initialize logger (reads LOG_LEVEL from environment)
	logger.Init()
	log := logger.Get()

	log.Info().Msg("Initializing application...")

	// Load configuration
	log.Debug().Msg("Loading configuration from environment variables...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}
	log.Info().
		Str("server_host", cfg.Server.Host).
		Int("server_port", cfg.Server.Port).
		Str("db_host", cfg.Database.Host).
		Int("db_port", cfg.Database.Port).
		Str("db_name", cfg.Database.DatabaseName).
		Msg("Configuration loaded successfully")

	// Set Gin mode based on environment
	env := os.Getenv("GIN_MODE")
	if env == "" {
		env = "debug"
	}
	gin.SetMode(env)
	log.Debug().Str("gin_mode", env).Msg("Gin mode set")

	// Create Gin router
	router := gin.New()
	log.Debug().Msg("Gin router created")

	// Setup middleware
	log.Debug().Msg("Setting up middleware...")
	api.SetupMiddleware(router)
	log.Info().Msg("Middleware setup completed")

	// Setup routes
	log.Debug().Msg("Setting up routes...")
	api.SetupRoutes(router)
	log.Info().Msg("Routes setup completed")

	// Create and start server
	log.Debug().Msg("Creating HTTP server...")
	srv := server.New(cfg, router)
	if err := srv.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}

	// Log server startup
	log.Info().
		Str("address", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)).
		Str("mode", env).
		Msg("HTTP server is running and ready to accept connections")

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Received shutdown signal, initiating graceful shutdown...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited gracefully")
}
