package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host         string        `koanf:"host"`
	Port         int           `koanf:"port"`
	ReadTimeout  time.Duration `koanf:"read_timeout"`
	WriteTimeout time.Duration `koanf:"write_timeout"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host         string `koanf:"host"`
	Port         int    `koanf:"port"`
	User         string `koanf:"user"`
	Password     string `koanf:"password"`
	DatabaseName string `koanf:"database_name"`
	SSLMode      string `koanf:"ssl_mode"`
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	SecretKey     string        `koanf:"secret_key"`
	Expiration    time.Duration `koanf:"expiration"`
	RefreshSecret string        `koanf:"refresh_secret"`
}

var (
	globalConfig *Config
)

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	// Load and validate server configuration
	if err := loadServerConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to load server config: %w", err)
	}

	// Load and validate database configuration
	if err := loadDatabaseConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to load database config: %w", err)
	}

	// Load and validate JWT configuration
	if err := loadJWTConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to load JWT config: %w", err)
	}

	globalConfig = cfg
	return cfg, nil
}

// loadServerConfig loads and validates server configuration
func loadServerConfig(cfg *Config) error {
	host := os.Getenv("SERVER_HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	portStr := os.Getenv("SERVER_PORT")
	if portStr == "" {
		portStr = "8080"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid SERVER_PORT: %w", err)
	}

	readTimeoutStr := os.Getenv("SERVER_READ_TIMEOUT")
	if readTimeoutStr == "" {
		readTimeoutStr = "15s"
	}
	readTimeout, err := time.ParseDuration(readTimeoutStr)
	if err != nil {
		return fmt.Errorf("invalid SERVER_READ_TIMEOUT: %w", err)
	}

	writeTimeoutStr := os.Getenv("SERVER_WRITE_TIMEOUT")
	if writeTimeoutStr == "" {
		writeTimeoutStr = "15s"
	}
	writeTimeout, err := time.ParseDuration(writeTimeoutStr)
	if err != nil {
		return fmt.Errorf("invalid SERVER_WRITE_TIMEOUT: %w", err)
	}

	cfg.Server = ServerConfig{
		Host:         host,
		Port:         port,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	return nil
}

// loadDatabaseConfig loads and validates database configuration
func loadDatabaseConfig(cfg *Config) error {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	portStr := os.Getenv("DB_PORT")
	if portStr == "" {
		portStr = "5432"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid DB_PORT: %w", err)
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		return fmt.Errorf("DB_USER is required")
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		return fmt.Errorf("DB_NAME is required")
	}

	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	cfg.Database = DatabaseConfig{
		Host:         host,
		Port:         port,
		User:         user,
		Password:     password,
		DatabaseName: dbName,
		SSLMode:      sslMode,
	}

	return nil
}

// loadJWTConfig loads and validates JWT configuration
func loadJWTConfig(cfg *Config) error {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		return fmt.Errorf("JWT_SECRET_KEY is required")
	}

	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if refreshSecret == "" {
		refreshSecret = secretKey // Fallback to secret key if not provided
	}

	expirationStr := os.Getenv("JWT_EXPIRATION")
	if expirationStr == "" {
		expirationStr = "24h"
	}
	expiration, err := time.ParseDuration(expirationStr)
	if err != nil {
		return fmt.Errorf("invalid JWT_EXPIRATION: %w", err)
	}

	cfg.JWT = JWTConfig{
		SecretKey:     secretKey,
		RefreshSecret: refreshSecret,
		Expiration:    expiration,
	}

	return nil
}

// Get returns the global configuration instance
func Get() *Config {
	if globalConfig == nil {
		panic("configuration not loaded. Call Load() first")
	}
	return globalConfig
}
