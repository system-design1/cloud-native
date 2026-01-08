package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"go-backend-service/internal/config"
	"go-backend-service/internal/db"
	"go-backend-service/internal/middleware"
	"go-backend-service/internal/repository"
	apperrors "go-backend-service/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a test database connection
// Returns nil and skips the test if database is not available
func setupTestDB(t *testing.T) *sql.DB {
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

func TestGetTenantSettingsByIDHandler_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// For invalid ID tests, we don't need a real repository since validation happens before the repository call
	// Create a dummy repository (won't be called)
	testDB := setupTestDB(t)
	if testDB == nil {
		return
	}
	defer testDB.Close()
	repo := repository.NewTenantSettingsRepository(testDB)

	// Create a test router
	router := gin.New()
	router.Use(middleware.ErrorHandlerMiddleware())
	v1 := router.Group("/v1")
	{
		otp := v1.Group("/otp")
		{
			otp.GET("/tenant-settings/:id", GetTenantSettingsByIDHandler(repo))
		}
	}

	testCases := []struct {
		name           string
		id             string
		expectedStatus int
	}{
		{
			name:           "non-numeric id",
			id:             "abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "zero id",
			id:             "0",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "negative id",
			id:             "-1",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/v1/otp/tenant-settings/"+tc.id, nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code, "Expected HTTP status %d", tc.expectedStatus)

			// Verify error response structure
			var errorResp apperrors.ErrorResponse
			err = json.Unmarshal(w.Body.Bytes(), &errorResp)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, errorResp.Code)
			assert.NotEmpty(t, errorResp.Message)
		})
	}
}

func TestGetTenantSettingsByIDHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testDB := setupTestDB(t)
	if testDB == nil {
		return
	}
	defer testDB.Close()

	repo := repository.NewTenantSettingsRepository(testDB)

	// Create a test router
	router := gin.New()
	router.Use(middleware.ErrorHandlerMiddleware())
	v1 := router.Group("/v1")
	{
		otp := v1.Group("/otp")
		{
			otp.GET("/tenant-settings/:id", GetTenantSettingsByIDHandler(repo))
		}
	}

	// Query a non-existent ID
	req, err := http.NewRequest("GET", "/v1/otp/tenant-settings/999999", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "Expected HTTP status 404")

	// Verify error response structure
	var errorResp apperrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, errorResp.Code)
	assert.Contains(t, errorResp.Message, "not found")
}

func TestGetTenantSettingsByIDHandler_Found(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testDB := setupTestDB(t)
	if testDB == nil {
		return
	}
	defer testDB.Close()

	repo := repository.NewTenantSettingsRepository(testDB)
	ctx := context.Background()

	// Insert a test record
	insertQuery := `
		INSERT INTO tenant_settings (
			tenant_code, name, status, otp_enabled, sms_provider, sms_api_key,
			rate_limit_per_min, signup_at, timezone, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	var insertedID int64
	err := testDB.QueryRowContext(ctx, insertQuery,
		"test-tenant-handler-001",
		"Test Tenant Handler",
		"active",
		true,
		"kavenegar",
		"test-api-key-123",
		60,
		time.Now(),
		"UTC",
		`{"test_key": "test_value"}`,
	).Scan(&insertedID)

	require.NoError(t, err, "Failed to insert test record")

	// Clean up: delete the test record after the test
	defer func() {
		_, _ = testDB.ExecContext(ctx, "DELETE FROM tenant_settings WHERE id = $1", insertedID)
	}()

	// Create a test router
	router := gin.New()
	router.Use(middleware.ErrorHandlerMiddleware())
	v1 := router.Group("/v1")
	{
		otp := v1.Group("/otp")
		{
			otp.GET("/tenant-settings/:id", GetTenantSettingsByIDHandler(repo))
		}
	}

	req, err := http.NewRequest("GET", "/v1/otp/tenant-settings/"+strconv.FormatInt(insertedID, 10), nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected HTTP status 200")

	// Parse the JSON response
	var response repository.TenantSettings
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify expected keys/fields are present
	assert.Equal(t, insertedID, response.ID)
	assert.Equal(t, "test-tenant-handler-001", response.TenantCode)
	assert.Equal(t, "Test Tenant Handler", response.Name)
	assert.Equal(t, repository.TenantStatusActive, response.Status)
	assert.True(t, response.OTPEnabled)
	assert.Equal(t, repository.SMSProviderKavenegar, response.SMSProvider)
	assert.NotNil(t, response.SMSAPIKey)
	assert.Equal(t, "test-api-key-123", *response.SMSAPIKey)
	assert.Equal(t, 60, response.RateLimitPerMin)
	assert.Equal(t, "UTC", response.Timezone)
	assert.NotNil(t, response.Metadata)
	assert.Equal(t, "test_value", response.Metadata["test_key"])
	assert.NotZero(t, response.CreatedAt)
	assert.NotZero(t, response.UpdatedAt)
	assert.Nil(t, response.DeletedAt)
}

