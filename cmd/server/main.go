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
	"go-backend-service/internal/db"
	"go-backend-service/internal/lifecycle"
	"go-backend-service/internal/logger"
	"go-backend-service/internal/redis"
	"go-backend-service/internal/repository"
	"go-backend-service/internal/server"
	"go-backend-service/internal/tracer"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file into environment if present (no error if file is missing)
	_ = godotenv.Load()

	// Initialize logger (reads LOG_LEVEL from environment)
	logger.Init()
	log := logger.Get()

	// Initialize lifecycle manager
	lifecycleMgr := lifecycle.NewManager()
	lifecycleMgr.SetState(lifecycle.StateStarting)

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
		Int("db_max_open_conns", cfg.Database.MaxOpenConns).
		Int("db_max_idle_conns", cfg.Database.MaxIdleConns).
		Int("redis_pool_size", cfg.Redis.PoolSize).
		Int("redis_min_idle_conns", cfg.Redis.MinIdleConns).
		Msg("Configuration loaded successfully")

	rdb, err := redis.NewClient(cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Redis client")
	}
	defer rdb.Close()

	// Initialize database connection pool
	log.Debug().Msg("Initializing database connection pool...")
	database, err := db.NewConnectionPool(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database connection pool")
	}

	// Verify database connectivity with ping (with retry mechanism)
	log.Debug().
		Str("host", cfg.Database.Host).
		Int("port", cfg.Database.Port).
		Str("database", cfg.Database.DatabaseName).
		Msg("Verifying database connectivity...")

	maxRetries := 10
	retryDelay := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		pingCtx, pingCancel := context.WithTimeout(context.Background(), 3*time.Second)
		err := db.Ping(pingCtx, database)
		pingCancel()

		if err == nil {
			log.Info().
				Int("attempt", i+1).
				Msg("Database connection successful")
			break
		}

		if i < maxRetries-1 {
			log.Warn().
				Err(err).
				Int("attempt", i+1).
				Int("max_retries", maxRetries).
				Dur("retry_delay", retryDelay).
				Str("host", cfg.Database.Host).
				Int("port", cfg.Database.Port).
				Msg("Database ping failed, retrying...")
			time.Sleep(retryDelay)
			if retryDelay < 5*time.Second {
				retryDelay += 500 * time.Millisecond // Gradual increase instead of exponential
			}
		} else {
			log.Fatal().
				Err(err).
				Int("total_attempts", maxRetries).
				Str("host", cfg.Database.Host).
				Int("port", cfg.Database.Port).
				Str("database", cfg.Database.DatabaseName).
				Msg("Failed to ping database after all retries. Make sure the database is running and accessible.")
		}
	}
	log.Info().
		Str("db_host", cfg.Database.Host).
		Int("db_port", cfg.Database.Port).
		Str("db_name", cfg.Database.DatabaseName).
		Msg("Database connection pool initialized and verified successfully")

	// Initialize repositories
	log.Debug().Msg("Initializing repositories...")
	tenantSettingsRepo := repository.NewTenantSettingsRepository(database)
	tenantSettingsInsertRepo := repository.NewTenantSettingsInsertRepository(database)
	redisRepo := repository.NewRedisBenchmarkRepository(rdb)
	log.Info().Msg("Repositories initialized successfully")

	// Set Gin mode from configuration
	gin.SetMode(cfg.App.GinMode)

	// Disable Gin's default debug output to ensure all logs are structured JSON
	// In debug mode, Gin outputs non-JSON debug messages. We use our own logger instead.
	if cfg.App.GinMode == "debug" {
		gin.DefaultWriter = gin.DefaultErrorWriter // Only log errors, not debug info
	}

	log.Debug().Str("gin_mode", cfg.App.GinMode).Msg("Gin mode set")

	// Initialize OpenTelemetry tracing
	if cfg.Tracing.Enabled {
		log.Debug().Msg("Initializing OpenTelemetry tracing...")
		if err := tracer.Init(cfg); err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize OpenTelemetry tracing")
		}
		log.Info().
			Str("service_name", cfg.Tracing.ServiceName).
			Str("service_version", cfg.Tracing.ServiceVersion).
			Bool("tempo_enabled", cfg.Tracing.TempoEnabled).
			Bool("jaeger_enabled", cfg.Tracing.JaegerEnabled).
			Msg("OpenTelemetry tracing initialized")
	} else {
		log.Debug().Msg("OpenTelemetry tracing is disabled")
	}

	// Create Gin router (without default logger to use our structured JSON logger)
	router := gin.New()
	log.Debug().Msg("Gin router created")

	// Setup middleware
	log.Debug().Msg("Setting up middleware...")
	api.SetupMiddleware(router)
	log.Info().Msg("Middleware setup completed")

	// Setup routes (pass lifecycle manager and repositories)
	log.Debug().Msg("Setting up routes...")
	api.SetupRoutes(router, lifecycleMgr, tenantSettingsRepo, tenantSettingsInsertRepo, redisRepo)
	log.Info().Msg("Routes setup completed")

	// Create and start server
	log.Debug().Msg("Creating HTTP server...")
	srv := server.New(cfg, router)
	if err := srv.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}

	// Mark application as ready
	lifecycleMgr.SetState(lifecycle.StateReady)

	// Log server startup
	log.Info().
		Str("address", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)).
		Str("mode", cfg.App.GinMode).
		Str("state", lifecycleMgr.GetState().String()).
		Msg("HTTP server is running and ready to accept connections")

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	// Mark application as shutting down
	lifecycleMgr.SetState(lifecycle.StateShuttingDown)

	log.Info().
		Str("signal", sig.String()).
		Str("state", lifecycleMgr.GetState().String()).
		Msg("Received shutdown signal, initiating graceful shutdown...")

	// Graceful shutdown with timeout from configuration
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.GracefulShutdownTimeout)
	defer cancel()

	// Shutdown HTTP server
	log.Info().Dur("timeout", cfg.Server.GracefulShutdownTimeout).Msg("Shutting down HTTP server...")
	if err := srv.Shutdown(ctx); err != nil {
		if err == context.DeadlineExceeded {
			log.Error().
				Err(err).
				Dur("timeout", cfg.Server.GracefulShutdownTimeout).
				Msg("Shutdown timeout exceeded, forcing termination")
			// Force shutdown if timeout exceeded
			os.Exit(1)
		} else {
			log.Error().Err(err).Msg("Error during HTTP server shutdown")
		}
	} else {
		log.Info().Msg("HTTP server shutdown completed successfully")
	}

	// Close database connection pool
	log.Info().Msg("Closing database connection pool...")
	if err := db.Close(database); err != nil {
		log.Error().Err(err).Msg("Error closing database connection pool")
	} else {
		log.Info().Msg("Database connection pool closed successfully")
	}

	// Shutdown tracer
	if cfg.Tracing.Enabled {
		log.Info().Msg("Shutting down OpenTelemetry tracer...")
		tracerCtx, tracerCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer tracerCancel()

		if err := tracer.Shutdown(tracerCtx); err != nil {
			if err == context.DeadlineExceeded {
				log.Error().Err(err).Msg("Tracer shutdown timeout exceeded")
			} else {
				log.Error().Err(err).Msg("Error shutting down tracer")
			}
		} else {
			log.Info().Msg("Tracer shut down successfully")
		}
	}

	// Mark application as shutdown
	lifecycleMgr.SetState(lifecycle.StateShutdown)

	log.Info().
		Str("state", lifecycleMgr.GetState().String()).
		Msg("Server exited gracefully")
}
