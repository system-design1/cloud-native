package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-backend-service/internal/otp"

	"github.com/redis/go-redis/v9"
)

type tenantSettingsSource interface {
	GetTenantSettingsByID(ctx context.Context, tenantID int64) (*TenantSettings, error)
}

// CachedTenantSettingsProvider loads tenant settings with Redis cache-aside behavior.
type CachedTenantSettingsProvider struct {
	client *redis.Client
	source tenantSettingsSource
	ttl    time.Duration
}

// NewCachedTenantSettingsProvider creates a tenant settings cache-aside provider.
func NewCachedTenantSettingsProvider(
	client *redis.Client,
	source tenantSettingsSource,
	ttl time.Duration,
) *CachedTenantSettingsProvider {
	return &CachedTenantSettingsProvider{
		client: client,
		source: source,
		ttl:    ttl,
	}
}

// GetTenantSettings returns cached tenant settings or falls back to the source repository.
func (p *CachedTenantSettingsProvider) GetTenantSettings(ctx context.Context, tenantID int64) (*otp.TenantSettings, error) {
	key := tenantSettingsCacheKey(tenantID)

	cached, err := p.client.Get(ctx, key).Bytes()
	if err == nil {
		var settings otp.TenantSettings
		if err := json.Unmarshal(cached, &settings); err == nil {
			return &settings, nil
		}
	} else if err != redis.Nil {
		// Redis failures are non-fatal for tenant lookup; fall back to source.
	}

	sourceSettings, err := p.source.GetTenantSettingsByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	settings := mapTenantSettingsToOTP(sourceSettings)
	if p.ttl > 0 {
		if data, err := json.Marshal(settings); err == nil {
			_ = p.client.Set(ctx, key, data, p.ttl).Err()
		}
	}

	return settings, nil
}

func tenantSettingsCacheKey(tenantID int64) string {
	return fmt.Sprintf("tenant:%d:settings", tenantID)
}

func mapTenantSettingsToOTP(settings *TenantSettings) *otp.TenantSettings {
	return &otp.TenantSettings{
		ID:              settings.ID,
		TenantCode:      settings.TenantCode,
		Name:            settings.Name,
		Status:          string(settings.Status),
		OTPEnabled:      settings.OTPEnabled,
		SMSProvider:     string(settings.SMSProvider),
		RateLimitPerMin: settings.RateLimitPerMin,
		Timezone:        settings.Timezone,
		Metadata:        settings.Metadata,
		ExpiresAt:       settings.ExpiresAt,
	}
}
