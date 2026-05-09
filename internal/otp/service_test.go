package otp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTenantProvider struct {
	settings *TenantSettings
	err      error
	calls    int
}

func (p *fakeTenantProvider) GetTenantSettings(ctx context.Context, tenantID int64) (*TenantSettings, error) {
	p.calls++
	if p.err != nil {
		return nil, p.err
	}
	return p.settings, nil
}

type fakeOTPStore struct {
	saveErr error
	saved   OTPState
	ttl     time.Duration
	calls   int
}

func (s *fakeOTPStore) Save(ctx context.Context, state OTPState, ttl time.Duration) error {
	s.calls++
	s.saved = state
	s.ttl = ttl
	return s.saveErr
}

func (s *fakeOTPStore) Get(ctx context.Context, tenantID int64, phone string) (*OTPState, error) {
	return nil, ErrNotImplemented
}

func (s *fakeOTPStore) IncrementAttempts(ctx context.Context, tenantID int64, phone string) (int, error) {
	return 0, ErrNotImplemented
}

func (s *fakeOTPStore) Delete(ctx context.Context, tenantID int64, phone string) error {
	return ErrNotImplemented
}

type fakeSMSProvider struct {
	err   error
	block bool
	req   SMSRequest
	calls int
}

func (p *fakeSMSProvider) SendOTP(ctx context.Context, req SMSRequest) (*SMSResult, error) {
	p.calls++
	p.req = req
	if p.block {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	if p.err != nil {
		return nil, p.err
	}
	return &SMSResult{
		Provider:  req.Provider,
		Status:    RequestStatusSent,
		MessageID: "message-id",
		SentAt:    time.Now().UTC(),
	}, nil
}

func TestServiceSendOTPSuccess(t *testing.T) {
	tenantProvider := &fakeTenantProvider{settings: activeTenantSettings()}
	store := &fakeOTPStore{}
	smsProvider := &fakeSMSProvider{}
	config := Config{
		CodeLength:      6,
		TTL:             2 * time.Minute,
		MaxAttempts:     3,
		ProviderTimeout: time.Second,
	}
	service := NewService(tenantProvider, store, smsProvider, nil, nil, config)
	req := SendRequest{
		TenantID: 42,
		Phone:    "+989121234567",
		Metadata: map[string]interface{}{"source": "test"},
	}

	resp, err := service.SendOTP(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.RequestID)
	assert.False(t, resp.ExpiredAt.IsZero())
	assert.Equal(t, 1, store.calls)
	assert.NotEmpty(t, store.saved.CodeHash)
	assert.NotEqual(t, smsProvider.req.Code, store.saved.CodeHash)
	assert.Equal(t, req.TenantID, store.saved.TenantID)
	assert.Equal(t, req.Phone, store.saved.Phone)
	assert.Equal(t, config.MaxAttempts, store.saved.MaxAttempts)
	assert.Equal(t, config.TTL, store.ttl)
	assert.True(t, resp.ExpiredAt.Equal(store.saved.ExpiresAt))
	assert.Equal(t, 1, smsProvider.calls)
	assert.Equal(t, resp.RequestID, store.saved.RequestID)
	assert.Equal(t, resp.RequestID, smsProvider.req.RequestID)
	assert.Equal(t, req.TenantID, smsProvider.req.TenantID)
	assert.Equal(t, req.Phone, smsProvider.req.Phone)
	assert.Equal(t, "fake", smsProvider.req.Provider)
}

func TestServiceSendOTPInvalidRequest(t *testing.T) {
	tests := []struct {
		name string
		req  SendRequest
	}{
		{name: "invalid tenant id", req: SendRequest{TenantID: 0, Phone: "+989121234567"}},
		{name: "empty phone", req: SendRequest{TenantID: 42, Phone: ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantProvider := &fakeTenantProvider{settings: activeTenantSettings()}
			store := &fakeOTPStore{}
			smsProvider := &fakeSMSProvider{}
			service := NewService(tenantProvider, store, smsProvider, nil, nil, Config{})

			resp, err := service.SendOTP(context.Background(), tt.req)

			require.Nil(t, resp)
			require.Error(t, err)
			assert.Equal(t, 0, tenantProvider.calls)
			assert.Equal(t, 0, store.calls)
			assert.Equal(t, 0, smsProvider.calls)
		})
	}
}

func TestServiceSendOTPTenantLookupError(t *testing.T) {
	lookupErr := errors.New("lookup failed")
	tenantProvider := &fakeTenantProvider{err: lookupErr}
	store := &fakeOTPStore{}
	smsProvider := &fakeSMSProvider{}
	service := NewService(tenantProvider, store, smsProvider, nil, nil, Config{})

	resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

	require.Nil(t, resp)
	assert.ErrorIs(t, err, lookupErr)
	assert.Equal(t, 1, tenantProvider.calls)
	assert.Equal(t, 0, store.calls)
	assert.Equal(t, 0, smsProvider.calls)
}

func TestServiceSendOTPTenantDisabled(t *testing.T) {
	expiredAt := time.Now().UTC().Add(-time.Minute)
	tests := []struct {
		name     string
		settings *TenantSettings
	}{
		{name: "otp disabled", settings: &TenantSettings{ID: 42, Status: "active", OTPEnabled: false}},
		{name: "inactive status", settings: &TenantSettings{ID: 42, Status: "inactive", OTPEnabled: true}},
		{name: "expired tenant", settings: &TenantSettings{ID: 42, Status: "active", OTPEnabled: true, ExpiresAt: &expiredAt}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantProvider := &fakeTenantProvider{settings: tt.settings}
			store := &fakeOTPStore{}
			smsProvider := &fakeSMSProvider{}
			service := NewService(tenantProvider, store, smsProvider, nil, nil, Config{})

			resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

			require.Nil(t, resp)
			assert.ErrorIs(t, err, ErrTenantDisabled)
			assert.Equal(t, 1, tenantProvider.calls)
			assert.Equal(t, 0, store.calls)
			assert.Equal(t, 0, smsProvider.calls)
		})
	}
}

func TestServiceSendOTPStoreSaveError(t *testing.T) {
	saveErr := errors.New("save failed")
	tenantProvider := &fakeTenantProvider{settings: activeTenantSettings()}
	store := &fakeOTPStore{saveErr: saveErr}
	smsProvider := &fakeSMSProvider{}
	service := NewService(tenantProvider, store, smsProvider, nil, nil, Config{})

	resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

	require.Nil(t, resp)
	assert.ErrorIs(t, err, saveErr)
	assert.Equal(t, 1, store.calls)
	assert.Equal(t, 0, smsProvider.calls)
}

func TestServiceSendOTPSMSProviderError(t *testing.T) {
	smsErr := errors.New("provider failed")
	tenantProvider := &fakeTenantProvider{settings: activeTenantSettings()}
	store := &fakeOTPStore{}
	smsProvider := &fakeSMSProvider{err: smsErr}
	service := NewService(tenantProvider, store, smsProvider, nil, nil, Config{})

	resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

	require.Nil(t, resp)
	assert.ErrorIs(t, err, ErrSMSProviderFailed)
	assert.ErrorIs(t, err, smsErr)
	assert.Equal(t, 1, store.calls)
	assert.Equal(t, 1, smsProvider.calls)
}

func TestServiceSendOTPSMSProviderTimeout(t *testing.T) {
	tenantProvider := &fakeTenantProvider{settings: activeTenantSettings()}
	store := &fakeOTPStore{}
	smsProvider := &fakeSMSProvider{block: true}
	service := NewService(tenantProvider, store, smsProvider, nil, nil, Config{ProviderTimeout: time.Millisecond})

	resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

	require.Nil(t, resp)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSMSProviderFailed)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Equal(t, 1, store.calls)
	assert.Equal(t, 1, smsProvider.calls)
}

func activeTenantSettings() *TenantSettings {
	return &TenantSettings{
		ID:              42,
		TenantCode:      "tenant-42",
		Name:            "Tenant 42",
		Status:          "active",
		OTPEnabled:      true,
		SMSProvider:     "fake",
		RateLimitPerMin: 60,
		Timezone:        "UTC",
	}
}
