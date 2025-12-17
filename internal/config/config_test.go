package config

import (
	"os"
	"testing"
)

func TestLoadServerConfig(t *testing.T) {
	// Set test environment variables
	os.Setenv("SERVER_HOST", "127.0.0.1")
	os.Setenv("SERVER_PORT", "3000")
	os.Setenv("SERVER_READ_TIMEOUT", "15s")
	os.Setenv("SERVER_WRITE_TIMEOUT", "15s")
	os.Setenv("SERVER_GRACEFUL_SHUTDOWN_TIMEOUT", "10s")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SSLMODE", "disable")
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-key")
	os.Setenv("JWT_EXPIRATION", "24h")
	os.Setenv("GIN_MODE", "debug")

	defer func() {
		os.Unsetenv("SERVER_HOST")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("SERVER_READ_TIMEOUT")
		os.Unsetenv("SERVER_WRITE_TIMEOUT")
		os.Unsetenv("SERVER_GRACEFUL_SHUTDOWN_TIMEOUT")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")
		os.Unsetenv("JWT_SECRET_KEY")
		os.Unsetenv("JWT_REFRESH_SECRET")
		os.Unsetenv("JWT_EXPIRATION")
		os.Unsetenv("GIN_MODE")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Expected SERVER_HOST to be 127.0.0.1, got %s", cfg.Server.Host)
	}

	if cfg.Server.Port != 3000 {
		t.Errorf("Expected SERVER_PORT to be 3000, got %d", cfg.Server.Port)
	}
}

func TestLoadConfigMissingRequiredFields(t *testing.T) {
	// Clear required environment variables
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("JWT_SECRET_KEY")

	_, err := Load()
	if err == nil {
		t.Error("Expected error when required fields are missing, got nil")
	}
}
