package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"go-backend-service/internal/otp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupOTPRequestLogTestDB(t *testing.T) *sql.DB {
	testDB := setupTestDB(t)
	if testDB == nil {
		return nil
	}

	var exists bool
	err := testDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'otp_requests'
		)
	`).Scan(&exists)
	if err != nil {
		_ = testDB.Close()
		t.Skipf("Skipping test: failed to check otp_requests table: %v", err)
	}
	if !exists {
		_ = testDB.Close()
		t.Skip("Skipping test: otp_requests table does not exist; apply the otp_requests migration first")
	}

	return testDB
}

func TestOTPRequestLogRepositoryCreateRequest(t *testing.T) {
	testDB := setupOTPRequestLogTestDB(t)
	if testDB == nil {
		return
	}
	defer testDB.Close()

	repo := NewOTPRequestLogRepository(testDB)
	ctx := context.Background()
	requestID := "test-create-" + time.Now().UTC().Format("20060102150405.000000000")
	defer cleanupOTPRequestLog(ctx, testDB, requestID)

	err := repo.CreateRequest(ctx, otp.OTPRequestLog{
		RequestID:     requestID,
		TenantID:      101,
		Phone:         "+989121110101",
		Status:        otp.RequestStatusPending,
		ProviderName:  "fake",
		Metadata:      map[string]interface{}{"source": "test"},
		CorrelationID: "correlation-create",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	})
	require.NoError(t, err)

	var tenantID int64
	var phone string
	var status string
	var providerName string
	var metadataJSON []byte
	var correlationID sql.NullString
	err = testDB.QueryRowContext(ctx, `
		SELECT tenant_id, phone, status, provider_name, metadata, correlation_id
		FROM otp_requests
		WHERE request_id = $1
	`, requestID).Scan(&tenantID, &phone, &status, &providerName, &metadataJSON, &correlationID)
	require.NoError(t, err)

	var metadata map[string]interface{}
	require.NoError(t, json.Unmarshal(metadataJSON, &metadata))
	assert.Equal(t, int64(101), tenantID)
	assert.Equal(t, "+989121110101", phone)
	assert.Equal(t, otp.RequestStatusPending, status)
	assert.Equal(t, "fake", providerName)
	assert.Equal(t, "test", metadata["source"])
	require.True(t, correlationID.Valid)
	assert.Equal(t, "correlation-create", correlationID.String)
}

func TestOTPRequestLogRepositoryUpdateProviderResult(t *testing.T) {
	testDB := setupOTPRequestLogTestDB(t)
	if testDB == nil {
		return
	}
	defer testDB.Close()

	repo := NewOTPRequestLogRepository(testDB)
	ctx := context.Background()
	requestID := "test-update-" + time.Now().UTC().Format("20060102150405.000000000")
	defer cleanupOTPRequestLog(ctx, testDB, requestID)

	err := repo.CreateRequest(ctx, otp.OTPRequestLog{
		RequestID:    requestID,
		TenantID:     102,
		Phone:        "+989121110102",
		Status:       otp.RequestStatusPending,
		ProviderName: "fake",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	})
	require.NoError(t, err)

	err = repo.UpdateProviderResult(ctx, otp.OTPProviderResultLog{
		RequestID:        requestID,
		Status:           otp.RequestStatusFailed,
		ProviderName:     "fake",
		ProviderResponse: map[string]interface{}{"message_id": "message-102", "simulated": true},
		ErrorMessage:     "provider timeout",
		UpdatedAt:        time.Now().UTC(),
	})
	require.NoError(t, err)

	var status string
	var providerName string
	var providerResponseJSON []byte
	var errorMessage sql.NullString
	err = testDB.QueryRowContext(ctx, `
		SELECT status, provider_name, provider_response, error_message
		FROM otp_requests
		WHERE request_id = $1
	`, requestID).Scan(&status, &providerName, &providerResponseJSON, &errorMessage)
	require.NoError(t, err)

	var providerResponse map[string]interface{}
	require.NoError(t, json.Unmarshal(providerResponseJSON, &providerResponse))
	assert.Equal(t, otp.RequestStatusFailed, status)
	assert.Equal(t, "fake", providerName)
	assert.Equal(t, "message-102", providerResponse["message_id"])
	assert.Equal(t, true, providerResponse["simulated"])
	require.True(t, errorMessage.Valid)
	assert.Equal(t, "provider timeout", errorMessage.String)
}

func TestOTPRequestLogRepositoryUpdateProviderResultNotFound(t *testing.T) {
	testDB := setupOTPRequestLogTestDB(t)
	if testDB == nil {
		return
	}
	defer testDB.Close()

	repo := NewOTPRequestLogRepository(testDB)
	err := repo.UpdateProviderResult(context.Background(), otp.OTPProviderResultLog{
		RequestID:    "missing-request-id",
		Status:       otp.RequestStatusFailed,
		ProviderName: "fake",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestOTPRequestLogRepositoryCreateRequestDuplicateRequestID(t *testing.T) {
	testDB := setupOTPRequestLogTestDB(t)
	if testDB == nil {
		return
	}
	defer testDB.Close()

	repo := NewOTPRequestLogRepository(testDB)
	ctx := context.Background()
	requestID := "test-duplicate-" + time.Now().UTC().Format("20060102150405.000000000")
	defer cleanupOTPRequestLog(ctx, testDB, requestID)

	log := otp.OTPRequestLog{
		RequestID: requestID,
		TenantID:  103,
		Phone:     "+989121110103",
		Status:    otp.RequestStatusPending,
	}

	require.NoError(t, repo.CreateRequest(ctx, log))
	err := repo.CreateRequest(ctx, log)

	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "create otp request log"))
}

func cleanupOTPRequestLog(ctx context.Context, db *sql.DB, requestID string) {
	_, _ = db.ExecContext(ctx, "DELETE FROM otp_requests WHERE request_id = $1", requestID)
}
