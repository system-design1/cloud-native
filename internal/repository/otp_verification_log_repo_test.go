package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"go-backend-service/internal/otp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupOTPVerificationLogTestDB(t *testing.T) *sql.DB {
	testDB := setupTestDB(t)
	if testDB == nil {
		return nil
	}

	var exists bool
	err := testDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'otp_verifications'
		)
	`).Scan(&exists)
	if err != nil {
		_ = testDB.Close()
		t.Skipf("Skipping test: failed to check otp_verifications table: %v", err)
	}
	if !exists {
		_ = testDB.Close()
		t.Skip("Skipping test: otp_verifications table does not exist; apply the otp_verifications migration first")
	}

	return testDB
}

func TestOTPVerificationLogRepositoryLogVerificationSuccess(t *testing.T) {
	testDB := setupOTPVerificationLogTestDB(t)
	if testDB == nil {
		return
	}
	defer testDB.Close()

	repo := NewOTPVerificationLogRepository(testDB)
	ctx := context.Background()
	requestID := "test-verify-success-" + time.Now().UTC().Format("20060102150405.000000000")
	defer cleanupOTPVerificationLog(ctx, testDB, requestID)

	err := repo.LogVerification(ctx, otp.OTPVerificationLog{
		RequestID:     requestID,
		TenantID:      201,
		Phone:         "+989122220201",
		Result:        otp.VerificationResultSuccess,
		Reason:        otp.ReasonVerified,
		AttemptCount:  1,
		CorrelationID: "correlation-success",
		CreatedAt:     time.Now().UTC(),
	})
	require.NoError(t, err)

	var tenantID int64
	var phone string
	var result string
	var reason string
	var attemptCount int
	var correlationID sql.NullString
	err = testDB.QueryRowContext(ctx, `
		SELECT tenant_id, phone, result, reason, attempt_count, correlation_id
		FROM otp_verifications
		WHERE request_id = $1
	`, requestID).Scan(&tenantID, &phone, &result, &reason, &attemptCount, &correlationID)
	require.NoError(t, err)

	assert.Equal(t, int64(201), tenantID)
	assert.Equal(t, "+989122220201", phone)
	assert.Equal(t, otp.VerificationResultSuccess, result)
	assert.Equal(t, otp.ReasonVerified, reason)
	assert.Equal(t, 1, attemptCount)
	require.True(t, correlationID.Valid)
	assert.Equal(t, "correlation-success", correlationID.String)
}

func TestOTPVerificationLogRepositoryLogVerificationFailed(t *testing.T) {
	testDB := setupOTPVerificationLogTestDB(t)
	if testDB == nil {
		return
	}
	defer testDB.Close()

	repo := NewOTPVerificationLogRepository(testDB)
	ctx := context.Background()
	requestID := "test-verify-failed-" + time.Now().UTC().Format("20060102150405.000000000")
	defer cleanupOTPVerificationLog(ctx, testDB, requestID)

	err := repo.LogVerification(ctx, otp.OTPVerificationLog{
		RequestID:    requestID,
		TenantID:     202,
		Phone:        "+989122220202",
		Result:       otp.VerificationResultFailed,
		Reason:       otp.ReasonInvalidCode,
		AttemptCount: 2,
		CreatedAt:    time.Now().UTC(),
	})
	require.NoError(t, err)

	var result string
	var reason string
	var attemptCount int
	err = testDB.QueryRowContext(ctx, `
		SELECT result, reason, attempt_count
		FROM otp_verifications
		WHERE request_id = $1
	`, requestID).Scan(&result, &reason, &attemptCount)
	require.NoError(t, err)

	assert.Equal(t, otp.VerificationResultFailed, result)
	assert.Equal(t, otp.ReasonInvalidCode, reason)
	assert.Equal(t, 2, attemptCount)
}

func TestOTPVerificationLogRepositoryLogVerificationDefaultsCreatedAtAndNullCorrelationID(t *testing.T) {
	testDB := setupOTPVerificationLogTestDB(t)
	if testDB == nil {
		return
	}
	defer testDB.Close()

	repo := NewOTPVerificationLogRepository(testDB)
	ctx := context.Background()
	requestID := "test-verify-defaults-" + time.Now().UTC().Format("20060102150405.000000000")
	defer cleanupOTPVerificationLog(ctx, testDB, requestID)

	err := repo.LogVerification(ctx, otp.OTPVerificationLog{
		RequestID:    requestID,
		TenantID:     203,
		Phone:        "+989122220203",
		Result:       otp.VerificationResultFailed,
		Reason:       otp.ReasonNotFound,
		AttemptCount: 0,
	})
	require.NoError(t, err)

	var createdAt time.Time
	var correlationID sql.NullString
	err = testDB.QueryRowContext(ctx, `
		SELECT created_at, correlation_id
		FROM otp_verifications
		WHERE request_id = $1
	`, requestID).Scan(&createdAt, &correlationID)
	require.NoError(t, err)

	assert.False(t, createdAt.IsZero())
	assert.False(t, correlationID.Valid)
}

func cleanupOTPVerificationLog(ctx context.Context, db *sql.DB, requestID string) {
	_, _ = db.ExecContext(ctx, "DELETE FROM otp_verifications WHERE request_id = $1", requestID)
}
