package repository

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"go-backend-service/internal/otp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTenantSettingsSource struct {
	settings *TenantSettings
	err      error
	calls    int
}

func (s *fakeTenantSettingsSource) GetTenantSettingsByID(ctx context.Context, tenantID int64) (*TenantSettings, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return s.settings, nil
}

func TestCachedTenantSettingsProviderCacheHit(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	ctx := context.Background()
	tenantID := int64(2001)
	key := tenantSettingsCacheKey(tenantID)
	defer client.Del(ctx, key)

	cached := otp.TenantSettings{
		ID:              tenantID,
		TenantCode:      "tenant-cache-hit",
		Name:            "Tenant Cache Hit",
		Status:          "active",
		OTPEnabled:      true,
		SMSProvider:     "other",
		RateLimitPerMin: 60,
		Timezone:        "UTC",
		Metadata:        map[string]interface{}{"source": "cache"},
	}
	data, err := json.Marshal(cached)
	require.NoError(t, err)
	require.NoError(t, client.Set(ctx, key, data, time.Minute).Err())

	source := &fakeTenantSettingsSource{
		settings: repositoryTenantSettings(tenantID, "source"),
	}
	provider := NewCachedTenantSettingsProvider(client, source, time.Minute)

	got, err := provider.GetTenantSettings(ctx, tenantID)

	require.NoError(t, err)
	assert.Equal(t, cached.ID, got.ID)
	assert.Equal(t, cached.TenantCode, got.TenantCode)
	assert.Equal(t, 0, source.calls)
}

func TestCachedTenantSettingsProviderCacheMissPopulatesRedis(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	ctx := context.Background()
	tenantID := int64(2002)
	key := tenantSettingsCacheKey(tenantID)
	defer client.Del(ctx, key)
	require.NoError(t, client.Del(ctx, key).Err())

	source := &fakeTenantSettingsSource{
		settings: repositoryTenantSettings(tenantID, "source"),
	}
	provider := NewCachedTenantSettingsProvider(client, source, time.Minute)

	got, err := provider.GetTenantSettings(ctx, tenantID)

	require.NoError(t, err)
	assert.Equal(t, tenantID, got.ID)
	assert.Equal(t, "tenant-source", got.TenantCode)
	assert.Equal(t, 1, source.calls)

	data, err := client.Get(ctx, key).Bytes()
	require.NoError(t, err)

	var cached otp.TenantSettings
	require.NoError(t, json.Unmarshal(data, &cached))
	assert.Equal(t, got.ID, cached.ID)
	assert.Equal(t, got.TenantCode, cached.TenantCode)
	assert.Empty(t, cached.Metadata["sms_api_key"])
}

func TestCachedTenantSettingsProviderMalformedCacheFallsBackAndOverwrites(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	ctx := context.Background()
	tenantID := int64(2003)
	key := tenantSettingsCacheKey(tenantID)
	defer client.Del(ctx, key)
	require.NoError(t, client.Set(ctx, key, []byte("{malformed-json"), time.Minute).Err())

	source := &fakeTenantSettingsSource{
		settings: repositoryTenantSettings(tenantID, "source"),
	}
	provider := NewCachedTenantSettingsProvider(client, source, time.Minute)

	got, err := provider.GetTenantSettings(ctx, tenantID)

	require.NoError(t, err)
	assert.Equal(t, tenantID, got.ID)
	assert.Equal(t, 1, source.calls)

	data, err := client.Get(ctx, key).Bytes()
	require.NoError(t, err)

	var cached otp.TenantSettings
	require.NoError(t, json.Unmarshal(data, &cached))
	assert.Equal(t, got.ID, cached.ID)
	assert.Equal(t, got.TenantCode, cached.TenantCode)
}

func TestCachedTenantSettingsProviderSourceErrorOnCacheMiss(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	ctx := context.Background()
	tenantID := int64(2004)
	key := tenantSettingsCacheKey(tenantID)
	defer client.Del(ctx, key)
	require.NoError(t, client.Del(ctx, key).Err())

	sourceErr := errors.New("source failed")
	source := &fakeTenantSettingsSource{err: sourceErr}
	provider := NewCachedTenantSettingsProvider(client, source, time.Minute)

	got, err := provider.GetTenantSettings(ctx, tenantID)

	require.Nil(t, got)
	assert.ErrorIs(t, err, sourceErr)
	assert.Equal(t, 1, source.calls)
}

func repositoryTenantSettings(id int64, codeSuffix string) *TenantSettings {
	apiKey := "secret-api-key"
	return &TenantSettings{
		ID:              id,
		TenantCode:      "tenant-" + codeSuffix,
		Name:            "Tenant " + codeSuffix,
		Status:          TenantStatusActive,
		OTPEnabled:      true,
		SMSProvider:     SMSProviderOther,
		SMSAPIKey:       &apiKey,
		RateLimitPerMin: 60,
		Timezone:        "UTC",
		Metadata:        map[string]interface{}{"source": codeSuffix},
	}
}
