package repository

import (
	"context"
	"database/sql"
	"fmt"
)

// TenantSettingsInsertRepository provides methods for inserting into tenant_settings_for_insert_new
type TenantSettingsInsertRepository struct {
	db *sql.DB
}

// NewTenantSettingsInsertRepository creates a new TenantSettingsInsertRepository
func NewTenantSettingsInsertRepository(db *sql.DB) *TenantSettingsInsertRepository {
	return &TenantSettingsInsertRepository{db: db}
}

// InsertTenantSettingsForInsertNew inserts one row into tenant_settings_for_insert_new and returns the id
func (r *TenantSettingsInsertRepository) InsertTenantSettingsForInsertNew(ctx context.Context, tenantCode string) (int64, error) {
	query := `
		INSERT INTO tenant_settings_for_insert_new (tenant_code, name, status, otp_enabled, sms_provider, sms_api_key, rate_limit_per_min, signup_at, expires_at, timezone, metadata, created_at, updated_at, deleted_at)
		VALUES ($1, 'Benchmark Tenant', 'active', true, 'other', 'eac566da612f3f0b551895235e6f4a29', 60, now(), NULL, 'UTC', '{}'::jsonb, now(), now(), NULL)
		RETURNING id
	`
	var id int64
	err := r.db.QueryRowContext(ctx, query, tenantCode).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert tenant settings for insert new: %w", err)
	}
	return id, nil
}
