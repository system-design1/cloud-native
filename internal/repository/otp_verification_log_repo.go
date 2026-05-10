package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go-backend-service/internal/otp"
)

// OTPVerificationLogRepository persists OTP verification attempt logs.
type OTPVerificationLogRepository struct {
	db *sql.DB
}

// NewOTPVerificationLogRepository creates a PostgreSQL-backed OTP verification logger.
func NewOTPVerificationLogRepository(db *sql.DB) *OTPVerificationLogRepository {
	return &OTPVerificationLogRepository{db: db}
}

// LogVerification inserts an OTP verification attempt log row.
func (r *OTPVerificationLogRepository) LogVerification(ctx context.Context, log otp.OTPVerificationLog) error {
	createdAt := log.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	query := `
		INSERT INTO otp_verifications (
			request_id, tenant_id, phone, result, reason, attempt_count,
			correlation_id, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	if _, err := r.db.ExecContext(
		ctx,
		query,
		log.RequestID,
		log.TenantID,
		log.Phone,
		log.Result,
		log.Reason,
		log.AttemptCount,
		nullableString(log.CorrelationID),
		createdAt,
	); err != nil {
		return fmt.Errorf("log otp verification: %w", err)
	}

	return nil
}
