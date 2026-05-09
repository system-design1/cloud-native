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

type fakeRequestLogger struct {
	createErr   error
	updateErr   error
	createLog   OTPRequestLog
	updateLogs  []OTPProviderResultLog
	createCalls int
	updateCalls int
}

func (l *fakeRequestLogger) CreateRequest(ctx context.Context, log OTPRequestLog) error {
	l.createCalls++
	l.createLog = log
	return l.createErr
}

func (l *fakeRequestLogger) UpdateProviderResult(ctx context.Context, log OTPProviderResultLog) error {
	l.updateCalls++
	l.updateLogs = append(l.updateLogs, log)
	return l.updateErr
}

func TestServiceSendOTPSuccess(t *testing.T) {
	tenantProvider := &fakeTenantProvider{settings: activeTenantSettings()}
	store := &fakeOTPStore{}
	smsProvider := &fakeSMSProvider{}
	requestLogger := &fakeRequestLogger{}
	config := Config{
		CodeLength:      6,
		TTL:             2 * time.Minute,
		MaxAttempts:     3,
		ProviderTimeout: time.Second,
	}
	service := NewService(tenantProvider, store, smsProvider, requestLogger, nil, config)
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
	assert.Equal(t, 1, requestLogger.createCalls)
	assert.Equal(t, resp.RequestID, requestLogger.createLog.RequestID)
	assert.Equal(t, req.TenantID, requestLogger.createLog.TenantID)
	assert.Equal(t, req.Phone, requestLogger.createLog.Phone)
	assert.Equal(t, RequestStatusPending, requestLogger.createLog.Status)
	assert.Equal(t, "fake", requestLogger.createLog.ProviderName)
	assert.Equal(t, "", requestLogger.createLog.CorrelationID)
	assert.Equal(t, req.Metadata, requestLogger.createLog.Metadata)
	require.Equal(t, 1, requestLogger.updateCalls)
	successLog := requestLogger.updateLogs[0]
	assert.Equal(t, resp.RequestID, successLog.RequestID)
	assert.Equal(t, RequestStatusSent, successLog.Status)
	assert.Equal(t, "fake", successLog.ProviderName)
	assert.Empty(t, successLog.ErrorMessage)
	assert.Equal(t, "fake", successLog.ProviderResponse["provider"])
	assert.Equal(t, RequestStatusSent, successLog.ProviderResponse["status"])
	assert.Equal(t, "message-id", successLog.ProviderResponse["message_id"])
	assert.Contains(t, successLog.ProviderResponse, "sent_at")
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
			requestLogger := &fakeRequestLogger{}
			service := NewService(tenantProvider, store, smsProvider, requestLogger, nil, Config{})

			resp, err := service.SendOTP(context.Background(), tt.req)

			require.Nil(t, resp)
			require.Error(t, err)
			assert.Equal(t, 0, tenantProvider.calls)
			assert.Equal(t, 0, store.calls)
			assert.Equal(t, 0, smsProvider.calls)
			assert.Equal(t, 0, requestLogger.createCalls)
			assert.Equal(t, 0, requestLogger.updateCalls)
		})
	}
}

func TestServiceSendOTPTenantLookupError(t *testing.T) {
	lookupErr := errors.New("lookup failed")
	tenantProvider := &fakeTenantProvider{err: lookupErr}
	store := &fakeOTPStore{}
	smsProvider := &fakeSMSProvider{}
	requestLogger := &fakeRequestLogger{}
	service := NewService(tenantProvider, store, smsProvider, requestLogger, nil, Config{})

	resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

	require.Nil(t, resp)
	assert.ErrorIs(t, err, lookupErr)
	assert.Equal(t, 1, tenantProvider.calls)
	assert.Equal(t, 0, store.calls)
	assert.Equal(t, 0, smsProvider.calls)
	assert.Equal(t, 0, requestLogger.createCalls)
	assert.Equal(t, 0, requestLogger.updateCalls)
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
			requestLogger := &fakeRequestLogger{}
			service := NewService(tenantProvider, store, smsProvider, requestLogger, nil, Config{})

			resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

			require.Nil(t, resp)
			assert.ErrorIs(t, err, ErrTenantDisabled)
			assert.Equal(t, 1, tenantProvider.calls)
			assert.Equal(t, 0, store.calls)
			assert.Equal(t, 0, smsProvider.calls)
			assert.Equal(t, 0, requestLogger.createCalls)
			assert.Equal(t, 0, requestLogger.updateCalls)
		})
	}
}

func TestServiceSendOTPCreateRequestError(t *testing.T) {
	createErr := errors.New("create log failed")
	tenantProvider := &fakeTenantProvider{settings: activeTenantSettings()}
	store := &fakeOTPStore{}
	smsProvider := &fakeSMSProvider{}
	requestLogger := &fakeRequestLogger{createErr: createErr}
	service := NewService(tenantProvider, store, smsProvider, requestLogger, nil, Config{})

	resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

	require.Nil(t, resp)
	assert.ErrorIs(t, err, createErr)
	assert.Equal(t, 1, requestLogger.createCalls)
	assert.Equal(t, 0, requestLogger.updateCalls)
	assert.Equal(t, 0, store.calls)
	assert.Equal(t, 0, smsProvider.calls)
}

func TestServiceSendOTPStoreSaveError(t *testing.T) {
	saveErr := errors.New("save failed")
	tenantProvider := &fakeTenantProvider{settings: activeTenantSettings()}
	store := &fakeOTPStore{saveErr: saveErr}
	smsProvider := &fakeSMSProvider{}
	requestLogger := &fakeRequestLogger{}
	service := NewService(tenantProvider, store, smsProvider, requestLogger, nil, Config{})

	resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

	require.Nil(t, resp)
	assert.ErrorIs(t, err, saveErr)
	assert.Equal(t, 1, requestLogger.createCalls)
	require.Equal(t, 1, requestLogger.updateCalls)
	assert.Equal(t, RequestStatusFailed, requestLogger.updateLogs[0].Status)
	assert.Equal(t, "fake", requestLogger.updateLogs[0].ProviderName)
	assert.Contains(t, requestLogger.updateLogs[0].ErrorMessage, "save otp state")
	assert.Equal(t, 1, store.calls)
	assert.Equal(t, 0, smsProvider.calls)
}

func TestServiceSendOTPSMSProviderError(t *testing.T) {
	smsErr := errors.New("provider failed")
	tenantProvider := &fakeTenantProvider{settings: activeTenantSettings()}
	store := &fakeOTPStore{}
	smsProvider := &fakeSMSProvider{err: smsErr}
	requestLogger := &fakeRequestLogger{}
	service := NewService(tenantProvider, store, smsProvider, requestLogger, nil, Config{})

	resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

	require.Nil(t, resp)
	assert.ErrorIs(t, err, ErrSMSProviderFailed)
	assert.ErrorIs(t, err, smsErr)
	assert.Equal(t, 1, requestLogger.createCalls)
	require.Equal(t, 1, requestLogger.updateCalls)
	assert.Equal(t, RequestStatusFailed, requestLogger.updateLogs[0].Status)
	assert.Equal(t, "fake", requestLogger.updateLogs[0].ProviderName)
	assert.Contains(t, requestLogger.updateLogs[0].ErrorMessage, smsErr.Error())
	assert.Equal(t, 1, store.calls)
	assert.Equal(t, 1, smsProvider.calls)
}

func TestServiceSendOTPSMSProviderTimeout(t *testing.T) {
	tenantProvider := &fakeTenantProvider{settings: activeTenantSettings()}
	store := &fakeOTPStore{}
	smsProvider := &fakeSMSProvider{block: true}
	requestLogger := &fakeRequestLogger{}
	service := NewService(tenantProvider, store, smsProvider, requestLogger, nil, Config{ProviderTimeout: time.Millisecond})

	resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

	require.Nil(t, resp)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSMSProviderFailed)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Equal(t, 1, requestLogger.createCalls)
	require.Equal(t, 1, requestLogger.updateCalls)
	assert.Equal(t, RequestStatusFailed, requestLogger.updateLogs[0].Status)
	assert.Contains(t, requestLogger.updateLogs[0].ErrorMessage, context.DeadlineExceeded.Error())
	assert.Equal(t, 1, store.calls)
	assert.Equal(t, 1, smsProvider.calls)
}

func TestServiceSendOTPSuccessUpdateProviderResultError(t *testing.T) {
	updateErr := errors.New("update log failed")
	tenantProvider := &fakeTenantProvider{settings: activeTenantSettings()}
	store := &fakeOTPStore{}
	smsProvider := &fakeSMSProvider{}
	requestLogger := &fakeRequestLogger{updateErr: updateErr}
	service := NewService(tenantProvider, store, smsProvider, requestLogger, nil, Config{})

	resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

	require.Nil(t, resp)
	assert.ErrorIs(t, err, updateErr)
	assert.Equal(t, 1, requestLogger.createCalls)
	assert.Equal(t, 1, requestLogger.updateCalls)
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
