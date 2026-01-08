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
	// Clear required environment variables (DB now has defaults, but JWT still required)
	os.Unsetenv("JWT_SECRET_KEY")

	_, err := Load()
	if err == nil {
		t.Error("Expected error when required fields are missing, got nil")
	}
}

func TestLoadDatabaseConfigWithDefaults(t *testing.T) {
	// Set minimum required env vars (JWT, Server, etc.)
	os.Setenv("SERVER_HOST", "127.0.0.1")
	os.Setenv("SERVER_PORT", "3000")
	os.Setenv("SERVER_READ_TIMEOUT", "15s")
	os.Setenv("SERVER_WRITE_TIMEOUT", "15s")
	os.Setenv("SERVER_GRACEFUL_SHUTDOWN_TIMEOUT", "10s")
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-key")
	os.Setenv("JWT_EXPIRATION", "24h")
	os.Setenv("GIN_MODE", "debug")

	// Unset DB env vars to test defaults
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_SSLMODE")

	defer func() {
		os.Unsetenv("SERVER_HOST")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("SERVER_READ_TIMEOUT")
		os.Unsetenv("SERVER_WRITE_TIMEOUT")
		os.Unsetenv("SERVER_GRACEFUL_SHUTDOWN_TIMEOUT")
		os.Unsetenv("JWT_SECRET_KEY")
		os.Unsetenv("JWT_REFRESH_SECRET")
		os.Unsetenv("JWT_EXPIRATION")
		os.Unsetenv("GIN_MODE")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config with DB defaults: %v", err)
	}

	// Verify defaults are applied
	if cfg.Database.Host != "localhost" {
		t.Errorf("Expected DB_HOST default to be 'localhost', got %s", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("Expected DB_PORT default to be 5432, got %d", cfg.Database.Port)
	}
	if cfg.Database.User != "postgres" {
		t.Errorf("Expected DB_USER default to be 'postgres', got %s", cfg.Database.User)
	}
	if cfg.Database.Password != "postgres" {
		t.Errorf("Expected DB_PASSWORD default to be 'postgres', got %s", cfg.Database.Password)
	}
	if cfg.Database.DatabaseName != "go_backend_db" {
		t.Errorf("Expected DB_NAME default to be 'go_backend_db', got %s", cfg.Database.DatabaseName)
	}
	if cfg.Database.SSLMode != "disable" {
		t.Errorf("Expected DB_SSLMODE default to be 'disable', got %s", cfg.Database.SSLMode)
	}
}

func TestLoadRoutePolicyConfig(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedPolicy RoutePolicyConfig
		expectError    bool
	}{
		{
			name: "route policy disabled",
			envVars: map[string]string{
				"OTEL_ROUTE_POLICY_ENABLED": "false",
			},
			expectedPolicy: RoutePolicyConfig{
				Enabled:       false,
				AlwaysRoutes:  []string{},
				DropRoutes:    []string{},
				RatioRoutes:   nil,
				DefaultPolicy: "",
				DefaultRatio:  0,
			},
			expectError: false,
		},
		{
			name: "route policy enabled with all options",
			envVars: map[string]string{
				"OTEL_ROUTE_POLICY_ENABLED": "true",
				"OTEL_ROUTE_ALWAYS":         "/delayed-hello,/test-error",
				"OTEL_ROUTE_DROP":           "/metrics",
				"OTEL_ROUTE_RATIO":          "/health=0.01,/live=0.01,/ready=0.01",
				"OTEL_ROUTE_DEFAULT":        "always",
				"OTEL_ROUTE_DEFAULT_RATIO":  "1.0",
			},
			expectedPolicy: RoutePolicyConfig{
				Enabled:      true,
				AlwaysRoutes: []string{"/delayed-hello", "/test-error"},
				DropRoutes:   []string{"/metrics"},
				RatioRoutes: map[string]float64{
					"/health": 0.01,
					"/live":   0.01,
					"/ready":  0.01,
				},
				DefaultPolicy: "always",
				DefaultRatio:  1.0,
			},
			expectError: false,
		},
		{
			name: "route policy with default ratio",
			envVars: map[string]string{
				"OTEL_ROUTE_POLICY_ENABLED": "true",
				"OTEL_ROUTE_DEFAULT":        "ratio",
				"OTEL_ROUTE_DEFAULT_RATIO":  "0.5",
			},
			expectedPolicy: RoutePolicyConfig{
				Enabled:       true,
				AlwaysRoutes:  []string{},
				DropRoutes:    []string{},
				RatioRoutes:   make(map[string]float64),
				DefaultPolicy: "ratio",
				DefaultRatio:  0.5,
			},
			expectError: false,
		},
		{
			name: "invalid default policy",
			envVars: map[string]string{
				"OTEL_ROUTE_POLICY_ENABLED": "true",
				"OTEL_ROUTE_DEFAULT":        "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid ratio value",
			envVars: map[string]string{
				"OTEL_ROUTE_POLICY_ENABLED": "true",
				"OTEL_ROUTE_RATIO":         "/health=invalid",
			},
			expectError: true,
		},
		{
			name: "ratio out of range",
			envVars: map[string]string{
				"OTEL_ROUTE_POLICY_ENABLED": "true",
				"OTEL_ROUTE_RATIO":         "/health=1.5",
			},
			expectError: true,
		},
		{
			name: "invalid default ratio",
			envVars: map[string]string{
				"OTEL_ROUTE_POLICY_ENABLED": "true",
				"OTEL_ROUTE_DEFAULT":        "ratio",
				"OTEL_ROUTE_DEFAULT_RATIO":  "2.0",
			},
			expectError: true,
		},
		{
			name: "trim spaces in routes",
			envVars: map[string]string{
				"OTEL_ROUTE_POLICY_ENABLED": "true",
				"OTEL_ROUTE_ALWAYS":         " /delayed-hello , /test-error ",
				"OTEL_ROUTE_DROP":           " /metrics ",
			},
			expectedPolicy: RoutePolicyConfig{
				Enabled:      true,
				AlwaysRoutes: []string{"/delayed-hello", "/test-error"},
				DropRoutes:   []string{"/metrics"},
				RatioRoutes:  make(map[string]float64),
				DefaultPolicy: "always",
				DefaultRatio:  1.0,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Cleanup
			defer func() {
				for k := range tt.envVars {
					os.Unsetenv(k)
				}
			}()

			policy, err := loadRoutePolicyConfig()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check Enabled
			if policy.Enabled != tt.expectedPolicy.Enabled {
				t.Errorf("Expected Enabled=%v, got %v", tt.expectedPolicy.Enabled, policy.Enabled)
			}

			// Check AlwaysRoutes
			if len(policy.AlwaysRoutes) != len(tt.expectedPolicy.AlwaysRoutes) {
				t.Errorf("Expected %d AlwaysRoutes, got %d", len(tt.expectedPolicy.AlwaysRoutes), len(policy.AlwaysRoutes))
			} else {
				for i, route := range tt.expectedPolicy.AlwaysRoutes {
					if i < len(policy.AlwaysRoutes) && policy.AlwaysRoutes[i] != route {
						t.Errorf("Expected AlwaysRoutes[%d]=%s, got %s", i, route, policy.AlwaysRoutes[i])
					}
				}
			}

			// Check DropRoutes
			if len(policy.DropRoutes) != len(tt.expectedPolicy.DropRoutes) {
				t.Errorf("Expected %d DropRoutes, got %d", len(tt.expectedPolicy.DropRoutes), len(policy.DropRoutes))
			} else {
				for i, route := range tt.expectedPolicy.DropRoutes {
					if i < len(policy.DropRoutes) && policy.DropRoutes[i] != route {
						t.Errorf("Expected DropRoutes[%d]=%s, got %s", i, route, policy.DropRoutes[i])
					}
				}
			}

			// Check RatioRoutes
			if len(policy.RatioRoutes) != len(tt.expectedPolicy.RatioRoutes) {
				t.Errorf("Expected %d RatioRoutes, got %d", len(tt.expectedPolicy.RatioRoutes), len(policy.RatioRoutes))
			} else {
				for path, ratio := range tt.expectedPolicy.RatioRoutes {
					if gotRatio, exists := policy.RatioRoutes[path]; !exists {
						t.Errorf("Expected RatioRoutes[%s] to exist", path)
					} else if gotRatio != ratio {
						t.Errorf("Expected RatioRoutes[%s]=%f, got %f", path, ratio, gotRatio)
					}
				}
			}

			// Check DefaultPolicy
			if policy.DefaultPolicy != tt.expectedPolicy.DefaultPolicy {
				t.Errorf("Expected DefaultPolicy=%s, got %s", tt.expectedPolicy.DefaultPolicy, policy.DefaultPolicy)
			}

			// Check DefaultRatio
			if policy.DefaultRatio != tt.expectedPolicy.DefaultRatio {
				t.Errorf("Expected DefaultRatio=%f, got %f", tt.expectedPolicy.DefaultRatio, policy.DefaultRatio)
			}
		})
	}
}

func TestParseCommaSeparatedList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single item",
			input:    "/health",
			expected: []string{"/health"},
		},
		{
			name:     "multiple items",
			input:    "/health,/live,/ready",
			expected: []string{"/health", "/live", "/ready"},
		},
		{
			name:     "with spaces",
			input:    " /health , /live , /ready ",
			expected: []string{"/health", "/live", "/ready"},
		},
		{
			name:     "empty items",
			input:    "/health,,/live",
			expected: []string{"/health", "/live"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCommaSeparatedList(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d items, got %d", len(tt.expected), len(result))
				return
			}
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Expected result[%d]=%s, got %s", i, expected, result[i])
				}
			}
		})
	}
}

func TestParseRatioRoutes(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    map[string]float64
		expectError bool
	}{
		{
			name:        "empty string",
			input:       "",
			expected:    map[string]float64{},
			expectError: false,
		},
		{
			name:        "single route",
			input:       "/health=0.01",
			expected:    map[string]float64{"/health": 0.01},
			expectError: false,
		},
		{
			name:        "multiple routes",
			input:       "/health=0.01,/live=0.01,/ready=0.01",
			expected:    map[string]float64{"/health": 0.01, "/live": 0.01, "/ready": 0.01},
			expectError: false,
		},
		{
			name:        "with spaces",
			input:       " /health=0.01 , /live=0.01 ",
			expected:    map[string]float64{"/health": 0.01, "/live": 0.01},
			expectError: false,
		},
		{
			name:        "invalid format - no equals",
			input:       "/health",
			expectError: true,
		},
		{
			name:        "invalid ratio value",
			input:       "/health=invalid",
			expectError: true,
		},
		{
			name:        "ratio out of range - too high",
			input:       "/health=1.5",
			expectError: true,
		},
		{
			name:        "ratio out of range - zero",
			input:       "/health=0.0",
			expectError: true,
		},
		{
			name:        "ratio out of range - negative",
			input:       "/health=-0.1",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseRatioRoutes(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d routes, got %d", len(tt.expected), len(result))
				return
			}

			for path, ratio := range tt.expected {
				if gotRatio, exists := result[path]; !exists {
					t.Errorf("Expected route %s to exist", path)
				} else if gotRatio != ratio {
					t.Errorf("Expected route %s to have ratio %f, got %f", path, ratio, gotRatio)
				}
			}
		})
	}
}
