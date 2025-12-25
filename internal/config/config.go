package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	App      AppConfig
	Tracing  TracingConfig
}

// AppConfig holds application-related configuration
type AppConfig struct {
	GinMode string `koanf:"gin_mode"`
}

// TracingConfig holds OpenTelemetry tracing configuration
type TracingConfig struct {
	Enabled        bool   `koanf:"enabled"`
	ServiceName    string `koanf:"service_name"`
	ServiceVersion string `koanf:"service_version"`
	TempoEndpoint  string `koanf:"tempo_endpoint"`
	TempoEnabled   bool   `koanf:"tempo_enabled"`
	JaegerEndpoint string `koanf:"jaeger_endpoint"`
	JaegerEnabled  bool   `koanf:"jaeger_enabled"`
	RoutePolicy    RoutePolicyConfig
}

// RoutePolicyConfig holds route-based tracing policy configuration
type RoutePolicyConfig struct {
	Enabled       bool
	AlwaysRoutes  []string
	DropRoutes    []string
	RatioRoutes   map[string]float64
	DefaultPolicy string
	DefaultRatio  float64
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host                    string        `koanf:"host"`
	Port                    int           `koanf:"port"`
	ReadTimeout             time.Duration `koanf:"read_timeout"`
	WriteTimeout            time.Duration `koanf:"write_timeout"`
	IdleTimeout             time.Duration `koanf:"idle_timeout"`
	GracefulShutdownTimeout time.Duration `koanf:"graceful_shutdown_timeout"`
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

	// Load and validate application configuration
	if err := loadAppConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to load app config: %w", err)
	}

	// Load and validate tracing configuration
	if err := loadTracingConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to load tracing config: %w", err)
	}

	globalConfig = cfg
	return cfg, nil
}

// loadServerConfig loads and validates server configuration
func loadServerConfig(cfg *Config) error {
	host := os.Getenv("SERVER_HOST")
	if host == "" {
		return fmt.Errorf("SERVER_HOST is required")
	}

	portStr := os.Getenv("SERVER_PORT")
	if portStr == "" {
		return fmt.Errorf("SERVER_PORT is required")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid SERVER_PORT: %w", err)
	}

	readTimeoutStr := os.Getenv("SERVER_READ_TIMEOUT")
	if readTimeoutStr == "" {
		return fmt.Errorf("SERVER_READ_TIMEOUT is required")
	}
	readTimeout, err := time.ParseDuration(readTimeoutStr)
	if err != nil {
		return fmt.Errorf("invalid SERVER_READ_TIMEOUT: %w", err)
	}

	writeTimeoutStr := os.Getenv("SERVER_WRITE_TIMEOUT")
	if writeTimeoutStr == "" {
		return fmt.Errorf("SERVER_WRITE_TIMEOUT is required")
	}
	writeTimeout, err := time.ParseDuration(writeTimeoutStr)
	if err != nil {
		return fmt.Errorf("invalid SERVER_WRITE_TIMEOUT: %w", err)
	}

	idleTimeoutStr := os.Getenv("SERVER_IDLE_TIMEOUT")
	if idleTimeoutStr == "" {
		// Default to 120 seconds if not specified
		idleTimeoutStr = "120s"
	}
	idleTimeout, err := time.ParseDuration(idleTimeoutStr)
	if err != nil {
		return fmt.Errorf("invalid SERVER_IDLE_TIMEOUT: %w", err)
	}

	gracefulShutdownTimeoutStr := os.Getenv("SERVER_GRACEFUL_SHUTDOWN_TIMEOUT")
	if gracefulShutdownTimeoutStr == "" {
		return fmt.Errorf("SERVER_GRACEFUL_SHUTDOWN_TIMEOUT is required")
	}
	gracefulShutdownTimeout, err := time.ParseDuration(gracefulShutdownTimeoutStr)
	if err != nil {
		return fmt.Errorf("invalid SERVER_GRACEFUL_SHUTDOWN_TIMEOUT: %w", err)
	}

	cfg.Server = ServerConfig{
		Host:                    host,
		Port:                    port,
		ReadTimeout:             readTimeout,
		WriteTimeout:            writeTimeout,
		IdleTimeout:             idleTimeout,
		GracefulShutdownTimeout: gracefulShutdownTimeout,
	}

	return nil
}

// loadDatabaseConfig loads and validates database configuration
func loadDatabaseConfig(cfg *Config) error {
	host := os.Getenv("DB_HOST")
	if host == "" {
		return fmt.Errorf("DB_HOST is required")
	}

	portStr := os.Getenv("DB_PORT")
	if portStr == "" {
		return fmt.Errorf("DB_PORT is required")
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
		return fmt.Errorf("DB_SSLMODE is required")
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
		return fmt.Errorf("JWT_REFRESH_SECRET is required")
	}

	expirationStr := os.Getenv("JWT_EXPIRATION")
	if expirationStr == "" {
		return fmt.Errorf("JWT_EXPIRATION is required")
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

// loadAppConfig loads and validates application configuration
func loadAppConfig(cfg *Config) error {
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		return fmt.Errorf("GIN_MODE is required")
	}

	cfg.App = AppConfig{
		GinMode: ginMode,
	}

	return nil
}

// loadTracingConfig loads and validates tracing configuration
func loadTracingConfig(cfg *Config) error {
	enabledStr := os.Getenv("OTEL_TRACING_ENABLED")
	enabled := enabledStr == "true" || enabledStr == "1"

	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "go-backend-service"
	}

	serviceVersion := os.Getenv("OTEL_SERVICE_VERSION")
	if serviceVersion == "" {
		serviceVersion = "1.0.0"
	}

	tempoEndpoint := os.Getenv("OTEL_TEMPO_ENDPOINT")
	tempoEnabledStr := os.Getenv("OTEL_TEMPO_ENABLED")
	tempoEnabled := tempoEnabledStr == "true" || tempoEnabledStr == "1"

	jaegerEndpoint := os.Getenv("OTEL_JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "localhost:4320" // Default to OTLP HTTP endpoint for local development
	}
	jaegerEnabledStr := os.Getenv("OTEL_JAEGER_ENABLED")
	jaegerEnabled := jaegerEnabledStr == "true" || jaegerEnabledStr == "1"

	// Load route policy configuration
	routePolicy, err := loadRoutePolicyConfig()
	if err != nil {
		return fmt.Errorf("failed to load route policy config: %w", err)
	}

	cfg.Tracing = TracingConfig{
		Enabled:        enabled,
		ServiceName:    serviceName,
		ServiceVersion: serviceVersion,
		TempoEndpoint:  tempoEndpoint,
		TempoEnabled:   tempoEnabled,
		JaegerEndpoint: jaegerEndpoint,
		JaegerEnabled:  jaegerEnabled,
		RoutePolicy:    routePolicy,
	}

	return nil
}

// loadRoutePolicyConfig loads and validates route policy configuration
func loadRoutePolicyConfig() (RoutePolicyConfig, error) {
	policy := RoutePolicyConfig{}

	// Check if route policy is enabled
	enabledStr := os.Getenv("OTEL_ROUTE_POLICY_ENABLED")
	policy.Enabled = enabledStr == "true" || enabledStr == "1"

	if !policy.Enabled {
		// Return default config when disabled
		return policy, nil
	}

	// Parse ALWAYS routes (comma-separated)
	alwaysStr := os.Getenv("OTEL_ROUTE_ALWAYS")
	if alwaysStr != "" {
		policy.AlwaysRoutes = parseCommaSeparatedList(alwaysStr)
	}

	// Parse DROP routes (comma-separated)
	dropStr := os.Getenv("OTEL_ROUTE_DROP")
	if dropStr != "" {
		policy.DropRoutes = parseCommaSeparatedList(dropStr)
	}

	// Parse RATIO routes (comma-separated path=ratio pairs)
	ratioStr := os.Getenv("OTEL_ROUTE_RATIO")
	if ratioStr != "" {
		ratioMap, err := parseRatioRoutes(ratioStr)
		if err != nil {
			return policy, fmt.Errorf("invalid OTEL_ROUTE_RATIO: %w", err)
		}
		policy.RatioRoutes = ratioMap
	} else {
		policy.RatioRoutes = make(map[string]float64)
	}

	// Parse default policy
	defaultPolicy := os.Getenv("OTEL_ROUTE_DEFAULT")
	if defaultPolicy == "" {
		defaultPolicy = "always" // Default to always
	}
	if defaultPolicy != "always" && defaultPolicy != "ratio" && defaultPolicy != "drop" {
		return policy, fmt.Errorf("invalid OTEL_ROUTE_DEFAULT: must be 'always', 'ratio', or 'drop'")
	}
	policy.DefaultPolicy = defaultPolicy

	// Parse default ratio (only used when default policy is 'ratio')
	defaultRatioStr := os.Getenv("OTEL_ROUTE_DEFAULT_RATIO")
	if defaultRatioStr == "" {
		policy.DefaultRatio = 1.0 // Default to 100%
	} else {
		defaultRatio, err := strconv.ParseFloat(defaultRatioStr, 64)
		if err != nil {
			return policy, fmt.Errorf("invalid OTEL_ROUTE_DEFAULT_RATIO: %w", err)
		}
		if defaultRatio <= 0.0 || defaultRatio > 1.0 {
			return policy, fmt.Errorf("OTEL_ROUTE_DEFAULT_RATIO must be between 0.0 and 1.0")
		}
		policy.DefaultRatio = defaultRatio
	}

	return policy, nil
}

// parseCommaSeparatedList parses a comma-separated list and trims spaces
func parseCommaSeparatedList(s string) []string {
	if s == "" {
		return []string{}
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// parseRatioRoutes parses comma-separated path=ratio pairs
// Example: "/health=0.01,/live=0.01,/ready=0.01"
func parseRatioRoutes(s string) (map[string]float64, error) {
	result := make(map[string]float64)

	if s == "" {
		return result, nil
	}

	parts := strings.Split(s, ",")
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		// Split by '=' to get path and ratio
		kv := strings.SplitN(trimmed, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid ratio format: %s (expected path=ratio)", trimmed)
		}

		path := strings.TrimSpace(kv[0])
		ratioStr := strings.TrimSpace(kv[1])

		ratio, err := strconv.ParseFloat(ratioStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid ratio value for %s: %w", path, err)
		}

		// Validate ratio range
		if ratio <= 0.0 || ratio > 1.0 {
			return nil, fmt.Errorf("ratio for %s must be between 0.0 and 1.0, got %f", path, ratio)
		}

		result[path] = ratio
	}

	return result, nil
}

// Get returns the global configuration instance
func Get() *Config {
	if globalConfig == nil {
		panic("configuration not loaded. Call Load() first")
	}
	return globalConfig
}
