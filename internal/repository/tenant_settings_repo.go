package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	apperrors "go-backend-service/pkg/errors"
)

// TenantStatus represents the status of a tenant
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusInactive  TenantStatus = "inactive"
	TenantStatusSuspended TenantStatus = "suspended"
)

// SMSProvider represents the SMS provider type
type SMSProvider string

const (
	SMSProviderKavenegar SMSProvider = "kavenegar"
	SMSProviderTwilio    SMSProvider = "twilio"
	SMSProviderGhasedak  SMSProvider = "ghasedak"
	SMSProviderOther     SMSProvider = "other"
)

// TenantSettings represents a row from the tenant_settings table
type TenantSettings struct {
	ID             int64                  `json:"id"`
	TenantCode     string                 `json:"tenant_code"`
	Name           string                 `json:"name"`
	Status         TenantStatus            `json:"status"`
	OTPEnabled    bool                   `json:"otp_enabled"`
	SMSProvider    SMSProvider            `json:"sms_provider"`
	SMSAPIKey      *string                `json:"sms_api_key,omitempty"` // Nullable
	RateLimitPerMin int                   `json:"rate_limit_per_min"`
	SignupAt       time.Time              `json:"signup_at"`
	ExpiresAt     *time.Time              `json:"expires_at,omitempty"` // Nullable
	Timezone      string                  `json:"timezone"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	DeletedAt     *time.Time              `json:"deleted_at,omitempty"` // Nullable (should be NULL for active records)
}

// TenantSettingsRepository provides methods for interacting with tenant_settings table
type TenantSettingsRepository struct {
	db *sql.DB
}

// NewTenantSettingsRepository creates a new TenantSettingsRepository
func NewTenantSettingsRepository(db *sql.DB) *TenantSettingsRepository {
	return &TenantSettingsRepository{db: db}
}

// GetTenantSettingsByID retrieves a tenant settings record by ID
// Returns ErrNotFound if the record doesn't exist or is deleted
func (r *TenantSettingsRepository) GetTenantSettingsByID(ctx context.Context, id int64) (*TenantSettings, error) {
	query := `
		SELECT 
			id, tenant_code, name, status, otp_enabled, sms_provider, sms_api_key,
			rate_limit_per_min, signup_at, expires_at, timezone, metadata,
			created_at, updated_at, deleted_at
		FROM tenant_settings
		WHERE id = $1 AND deleted_at IS NULL
	`

	var ts TenantSettings
	var metadataJSON []byte
	var smsAPIKey sql.NullString
	var expiresAt sql.NullTime
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ts.ID,
		&ts.TenantCode,
		&ts.Name,
		&ts.Status,
		&ts.OTPEnabled,
		&ts.SMSProvider,
		&smsAPIKey,
		&ts.RateLimitPerMin,
		&ts.SignupAt,
		&expiresAt,
		&ts.Timezone,
		&metadataJSON,
		&ts.CreatedAt,
		&ts.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound(fmt.Sprintf("tenant settings with id %d not found", id))
		}
		return nil, fmt.Errorf("failed to get tenant settings by id: %w", err)
	}

	// Handle nullable fields
	if smsAPIKey.Valid {
		ts.SMSAPIKey = &smsAPIKey.String
	}
	if expiresAt.Valid {
		ts.ExpiresAt = &expiresAt.Time
	}
	if deletedAt.Valid {
		ts.DeletedAt = &deletedAt.Time
	}

	// Parse JSONB metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &ts.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	} else {
		ts.Metadata = make(map[string]interface{})
	}

	return &ts, nil
}

