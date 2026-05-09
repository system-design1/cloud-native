package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go-backend-service/internal/otp"
)

// OTPRequestLogRepository persists OTP request and provider result logs.
type OTPRequestLogRepository struct {
	db *sql.DB
}

// NewOTPRequestLogRepository creates a PostgreSQL-backed OTP request logger.
func NewOTPRequestLogRepository(db *sql.DB) *OTPRequestLogRepository {
	return &OTPRequestLogRepository{db: db}
}

// CreateRequest inserts an initial OTP request log row.
func (r *OTPRequestLogRepository) CreateRequest(ctx context.Context, log otp.OTPRequestLog) error {
	metadata, err := marshalJSONMap(log.Metadata)
	if err != nil {
		return fmt.Errorf("create otp request log: marshal metadata: %w", err)
	}

	createdAt := log.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	updatedAt := log.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}

	query := `
		INSERT INTO otp_requests (
			request_id, tenant_id, phone, status, provider_name, error_message,
			metadata, correlation_id, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, $10)
	`

	if _, err := r.db.ExecContext(
		ctx,
		query,
		log.RequestID,
		log.TenantID,
		log.Phone,
		log.Status,
		log.ProviderName,
		nullableString(log.ErrorMessage),
		string(metadata),
		nullableString(log.CorrelationID),
		createdAt,
		updatedAt,
	); err != nil {
		return fmt.Errorf("create otp request log: %w", err)
	}

	return nil
}

// UpdateProviderResult updates provider result fields for an existing request.
func (r *OTPRequestLogRepository) UpdateProviderResult(ctx context.Context, log otp.OTPProviderResultLog) error {
	providerResponse, err := marshalJSONMap(log.ProviderResponse)
	if err != nil {
		return fmt.Errorf("update otp provider result: marshal provider response: %w", err)
	}

	updatedAt := log.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	query := `
		UPDATE otp_requests
		SET status = $1,
			provider_name = $2,
			provider_response = $3::jsonb,
			error_message = $4,
			updated_at = $5
		WHERE request_id = $6
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		log.Status,
		log.ProviderName,
		string(providerResponse),
		nullableString(log.ErrorMessage),
		updatedAt,
		log.RequestID,
	)
	if err != nil {
		return fmt.Errorf("update otp provider result: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update otp provider result: rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("update otp provider result: request_id %q not found", log.RequestID)
	}

	return nil
}

func marshalJSONMap(value map[string]interface{}) ([]byte, error) {
	if value == nil {
		value = map[string]interface{}{}
	}
	return json.Marshal(value)
}

func nullableString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}
