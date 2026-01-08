package repository

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"testing"
	"time"

	"go-backend-service/internal/config"
	"go-backend-service/internal/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a test database connection
// Returns nil and skips the test if database is not available
func setupTestDB(t *testing.T) *sql.DB {
	// Use test database config or defaults
	cfg := &config.DatabaseConfig{
		Host:         getEnvOrDefault("DB_HOST", "localhost"),
		Port:         getEnvIntOrDefault("DB_PORT", 5432),
		User:         getEnvOrDefault("DB_USER", "postgres"),
		Password:     getEnvOrDefault("DB_PASSWORD", "postgres"),
		DatabaseName: getEnvOrDefault("DB_NAME", "go_backend_db"),
		SSLMode:      getEnvOrDefault("DB_SSLMODE", "disable"),
	}

	testDB, err := db.NewConnectionPool(cfg)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return nil
	}

	// Try to ping the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Ping(ctx, testDB); err != nil {
		testDB.Close()
		t.Skipf("Skipping test: database ping failed: %v", err)
		return nil
	}

	return testDB
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if result, err := strconv.Atoi(value); err == nil {
			return result
		}
	}
	return defaultValue
}

func TestGetTenantSettingsByID_NotFound(t *testing.T) {
	testDB := setupTestDB(t)
	if testDB == nil {
		return
	}
	defer testDB.Close()

	repo := NewTenantSettingsRepository(testDB)
	ctx := context.Background()

	// Try to fetch a non-existent ID
	_, err := repo.GetTenantSettingsByID(ctx, 999999)

	// Should return ErrNotFound
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetTenantSettingsByID_Found(t *testing.T) {
	testDB := setupTestDB(t)
	if testDB == nil {
		return
	}
	defer testDB.Close()

	repo := NewTenantSettingsRepository(testDB)
	ctx := context.Background()

	// First, insert a test record
	insertQuery := `
		INSERT INTO tenant_settings (
			tenant_code, name, status, otp_enabled, sms_provider,
			rate_limit_per_min, signup_at, timezone, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	var insertedID int64
	err := testDB.QueryRowContext(ctx, insertQuery,
		"test-tenant-001",
		"Test Tenant",
		"active",
		true,
		"other",
		60,
		time.Now(),
		"UTC",
		`{"test": "data"}`,
	).Scan(&insertedID)

	require.NoError(t, err, "Failed to insert test record")

	// Clean up: delete the test record after the test
	defer func() {
		_, _ = testDB.ExecContext(ctx, "DELETE FROM tenant_settings WHERE id = $1", insertedID)
	}()

	// Now fetch it
	result, err := repo.GetTenantSettingsByID(ctx, insertedID)

	// Should succeed
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify the fields
	assert.Equal(t, insertedID, result.ID)
	assert.Equal(t, "test-tenant-001", result.TenantCode)
	assert.Equal(t, "Test Tenant", result.Name)
	assert.Equal(t, TenantStatusActive, result.Status)
	assert.True(t, result.OTPEnabled)
	assert.Equal(t, SMSProviderOther, result.SMSProvider)
	assert.Equal(t, 60, result.RateLimitPerMin)
	assert.Equal(t, "UTC", result.Timezone)
	assert.NotNil(t, result.Metadata)
	assert.Nil(t, result.DeletedAt, "DeletedAt should be NULL for active records")
}

