package config

import (
	"os"
	"testing"
)

func TestLoadServerConfig(t *testing.T) {
	// Set test environment variables
	os.Setenv("SERVER_HOST", "127.0.0.1")
	os.Setenv("SERVER_PORT", "3000")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")

	defer func() {
		os.Unsetenv("SERVER_HOST")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("JWT_SECRET_KEY")
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
