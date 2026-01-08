package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go-backend-service/internal/config"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// NewConnectionPool creates a new PostgreSQL connection pool using the provided config
// Returns the connection pool and any error encountered during initialization
func NewConnectionPool(cfg *config.DatabaseConfig) (*sql.DB, error) {
	// Build connection string
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DatabaseName,
		cfg.SSLMode,
	)

	// Open connection pool
	// Note: sql.Open doesn't actually connect, it just validates the DSN
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool settings
	db.SetMaxOpenConns(25)                 // Maximum number of open connections
	db.SetMaxIdleConns(5)                   // Maximum number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum connection lifetime
	db.SetConnMaxIdleTime(10 * time.Minute) // Maximum idle connection time

	return db, nil
}

// Ping verifies the database connection by executing a ping
// Returns an error if the connection cannot be established
func Ping(ctx context.Context, db *sql.DB) error {
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}

// Close closes the database connection pool
func Close(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}

