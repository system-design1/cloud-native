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
	saveErr         error
	getErr          error
	incrementErr    error
	deleteErr       error
	state           *OTPState
	incrementResult int
	saved           OTPState
	ttl             time.Duration
	calls           int
	getCalls        int
	incrementCalls  int
	deleteCalls     int
}

func (s *fakeOTPStore) Save(ctx context.Context, state OTPState, ttl time.Duration) error {
	s.calls++
	s.saved = state
	s.ttl = ttl
	return s.saveErr
}

func (s *fakeOTPStore) Get(ctx context.Context, tenantID int64, phone string) (*OTPState, error) {
	s.getCalls++
	if s.getErr != nil {
		return nil, s.getErr
	}
	return s.state, nil
}

func (s *fakeOTPStore) IncrementAttempts(ctx context.Context, tenantID int64, phone string) (int, error) {
	s.incrementCalls++
	if s.incrementErr != nil {
		return 0, s.incrementErr
	}
	if s.incrementResult != 0 {
		return s.incrementResult, nil
	}
	if s.state != nil {
		s.state.AttemptCount++
		return s.state.AttemptCount, nil
	}
	return 0, nil
}

func (s *fakeOTPStore) Delete(ctx context.Context, tenantID int64, phone string) error {
	s.deleteCalls++
	if s.deleteErr != nil {
		return s.deleteErr
	}
	s.state = nil
	return nil
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

type fakeSendRateLimiter struct {
	err      error
	calls    int
	tenantID int64
	phone    string
}

func (l *fakeSendRateLimiter) AllowSend(ctx context.Context, tenantID int64, phone string) error {
	l.calls++
	l.tenantID = tenantID
	l.phone = phone
	return l.err
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

type fakeVerificationLogger struct {
	err   error
	logs  []OTPVerificationLog
	calls int
}

func (l *fakeVerificationLogger) LogVerification(ctx context.Context, log OTPVerificationLog) error {
	l.calls++
	l.logs = append(l.logs, log)
	return l.err
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
	assert.Equal(t, 1, store.getCalls)
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
			assert.Equal(t, 0, store.getCalls)
			assert.Equal(t, 0, store.calls)
			assert.Equal(t, 0, smsProvider.calls)
			assert.Equal(t, 0, requestLogger.createCalls)
			assert.Equal(t, 0, requestLogger.updateCalls)
		})
	}
}

func TestServiceSendOTPExistingActiveOTPBlocked(t *testing.T) {
	tenantProvider := &fakeTenantProvider{settings: activeTenantSettings()}
	store := &fakeOTPStore{state: activeOTPState("123456")}
	smsProvider := &fakeSMSProvider{}
	requestLogger := &fakeRequestLogger{}
	service := NewService(tenantProvider, store, smsProvider, requestLogger, nil, Config{})

	resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

	require.Nil(t, resp)
	assert.ErrorIs(t, err, ErrOTPAlreadyActive)
	assert.Equal(t, 1, tenantProvider.calls)
	assert.Equal(t, 1, store.getCalls)
	assert.Equal(t, 0, requestLogger.createCalls)
	assert.Equal(t, 0, requestLogger.updateCalls)
	assert.Equal(t, 0, store.calls)
	assert.Equal(t, 0, smsProvider.calls)
	assert.Equal(t, 0, store.deleteCalls)
}

func TestServiceSendOTPExistingExpiredOTPDeletedAndProceeds(t *testing.T) {
	expired := activeOTPState("123456")
	expired.ExpiresAt = time.Now().UTC().Add(-time.Second)
	tenantProvider := &fakeTenantProvider{settings: activeTenantSettings()}
	store := &fakeOTPStore{state: expired}
	smsProvider := &fakeSMSProvider{}
	requestLogger := &fakeRequestLogger{}
	service := NewService(tenantProvider, store, smsProvider, requestLogger, nil, Config{})

	resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 1, store.getCalls)
	assert.Equal(t, 1, store.deleteCalls)
	assert.Equal(t, 1, requestLogger.createCalls)
	assert.Equal(t, 1, store.calls)
	assert.Equal(t, 1, smsProvider.calls)
}

func TestServiceSendOTPExistingExpiredOTPDeleteFailureStillProceeds(t *testing.T) {
	expired := activeOTPState("123456")
	expired.ExpiresAt = time.Now().UTC().Add(-time.Second)
	tenantProvider := &fakeTenantProvider{settings: activeTenantSettings()}
	store := &fakeOTPStore{state: expired, deleteErr: errors.New("delete failed")}
	smsProvider := &fakeSMSProvider{}
	requestLogger := &fakeRequestLogger{}
	service := NewService(tenantProvider, store, smsProvider, requestLogger, nil, Config{})

	resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 1, store.getCalls)
	assert.Equal(t, 1, store.deleteCalls)
	assert.Equal(t, 1, requestLogger.createCalls)
	assert.Equal(t, 1, store.calls)
	assert.Equal(t, 1, smsProvider.calls)
}

func TestServiceSendOTPExistingOTPNotFoundProceeds(t *testing.T) {
	tenantProvider := &fakeTenantProvider{settings: activeTenantSettings()}
	store := &fakeOTPStore{getErr: ErrOTPNotFound}
	smsProvider := &fakeSMSProvider{}
	requestLogger := &fakeRequestLogger{}
	service := NewService(tenantProvider, store, smsProvider, requestLogger, nil, Config{})

	resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 1, store.getCalls)
	assert.Equal(t, 1, requestLogger.createCalls)
	assert.Equal(t, 1, store.calls)
	assert.Equal(t, 1, smsProvider.calls)
}

func TestServiceSendOTPExistingOTPGetErrorAborts(t *testing.T) {
	getErr := errors.New("redis get failed")
	tenantProvider := &fakeTenantProvider{settings: activeTenantSettings()}
	store := &fakeOTPStore{getErr: getErr}
	smsProvider := &fakeSMSProvider{}
	requestLogger := &fakeRequestLogger{}
	service := NewService(tenantProvider, store, smsProvider, requestLogger, nil, Config{})

	resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

	require.Nil(t, resp)
	assert.ErrorIs(t, err, getErr)
	assert.Equal(t, 1, store.getCalls)
	assert.Equal(t, 0, requestLogger.createCalls)
	assert.Equal(t, 0, requestLogger.updateCalls)
	assert.Equal(t, 0, store.calls)
	assert.Equal(t, 0, smsProvider.calls)
}

func TestServiceSendOTPLimiterAllowsSend(t *testing.T) {
	tenantProvider := &fakeTenantProvider{settings: activeTenantSettings()}
	store := &fakeOTPStore{}
	smsProvider := &fakeSMSProvider{}
	requestLogger := &fakeRequestLogger{}
	limiter := &fakeSendRateLimiter{}
	service := NewService(tenantProvider, store, smsProvider, requestLogger, nil, Config{})
	service.SetSendRateLimiter(limiter)

	resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 1, limiter.calls)
	assert.Equal(t, int64(42), limiter.tenantID)
	assert.Equal(t, "+989121234567", limiter.phone)
	assert.Equal(t, 1, requestLogger.createCalls)
	assert.Equal(t, 1, store.calls)
	assert.Equal(t, 1, smsProvider.calls)
}

func TestServiceSendOTPLimiterRateLimited(t *testing.T) {
	tenantProvider := &fakeTenantProvider{settings: activeTenantSettings()}
	store := &fakeOTPStore{}
	smsProvider := &fakeSMSProvider{}
	requestLogger := &fakeRequestLogger{}
	limiter := &fakeSendRateLimiter{err: ErrOTPRateLimited}
	service := NewService(tenantProvider, store, smsProvider, requestLogger, nil, Config{})
	service.SetSendRateLimiter(limiter)

	resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

	require.Nil(t, resp)
	assert.ErrorIs(t, err, ErrOTPRateLimited)
	assert.Equal(t, 1, limiter.calls)
	assert.Equal(t, 0, requestLogger.createCalls)
	assert.Equal(t, 0, requestLogger.updateCalls)
	assert.Equal(t, 0, store.calls)
	assert.Equal(t, 0, smsProvider.calls)
}

func TestServiceSendOTPLimiterInfrastructureError(t *testing.T) {
	limitErr := errors.New("limiter unavailable")
	tenantProvider := &fakeTenantProvider{settings: activeTenantSettings()}
	store := &fakeOTPStore{}
	smsProvider := &fakeSMSProvider{}
	requestLogger := &fakeRequestLogger{}
	limiter := &fakeSendRateLimiter{err: limitErr}
	service := NewService(tenantProvider, store, smsProvider, requestLogger, nil, Config{})
	service.SetSendRateLimiter(limiter)

	resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

	require.Nil(t, resp)
	assert.ErrorIs(t, err, limitErr)
	assert.Equal(t, 1, limiter.calls)
	assert.Equal(t, 0, requestLogger.createCalls)
	assert.Equal(t, 0, requestLogger.updateCalls)
	assert.Equal(t, 0, store.calls)
	assert.Equal(t, 0, smsProvider.calls)
}

func TestServiceSendOTPActiveOTPBeforeLimiter(t *testing.T) {
	tenantProvider := &fakeTenantProvider{settings: activeTenantSettings()}
	store := &fakeOTPStore{state: activeOTPState("123456")}
	smsProvider := &fakeSMSProvider{}
	requestLogger := &fakeRequestLogger{}
	limiter := &fakeSendRateLimiter{}
	service := NewService(tenantProvider, store, smsProvider, requestLogger, nil, Config{})
	service.SetSendRateLimiter(limiter)

	resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

	require.Nil(t, resp)
	assert.ErrorIs(t, err, ErrOTPAlreadyActive)
	assert.Equal(t, 0, limiter.calls)
	assert.Equal(t, 0, requestLogger.createCalls)
	assert.Equal(t, 0, store.calls)
	assert.Equal(t, 0, smsProvider.calls)
}

func TestServiceSendOTPTenantDisabledBeforeLimiter(t *testing.T) {
	tenantProvider := &fakeTenantProvider{settings: &TenantSettings{ID: 42, Status: "inactive", OTPEnabled: true}}
	store := &fakeOTPStore{}
	smsProvider := &fakeSMSProvider{}
	requestLogger := &fakeRequestLogger{}
	limiter := &fakeSendRateLimiter{}
	service := NewService(tenantProvider, store, smsProvider, requestLogger, nil, Config{})
	service.SetSendRateLimiter(limiter)

	resp, err := service.SendOTP(context.Background(), SendRequest{TenantID: 42, Phone: "+989121234567"})

	require.Nil(t, resp)
	assert.ErrorIs(t, err, ErrTenantDisabled)
	assert.Equal(t, 0, limiter.calls)
	assert.Equal(t, 0, store.getCalls)
	assert.Equal(t, 0, requestLogger.createCalls)
	assert.Equal(t, 0, store.calls)
	assert.Equal(t, 0, smsProvider.calls)
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

func TestServiceVerifyOTPSuccess(t *testing.T) {
	store := &fakeOTPStore{state: activeOTPState("123456")}
	verifyLogger := &fakeVerificationLogger{}
	service := NewService(nil, store, nil, nil, verifyLogger, Config{})

	resp, err := service.VerifyOTP(context.Background(), VerifyRequest{
		TenantID: 42,
		Phone:    "+989121234567",
		Code:     "123456",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Verified)
	assert.Equal(t, "request-verify", resp.RequestID)
	assert.Equal(t, 1, store.deleteCalls)
	assert.Equal(t, 0, store.incrementCalls)
	assertVerificationLog(t, verifyLogger, VerificationResultSuccess, ReasonVerified, "request-verify", 0)
}

func TestServiceVerifyOTPInvalidRequest(t *testing.T) {
	tests := []struct {
		name string
		req  VerifyRequest
	}{
		{name: "invalid tenant id", req: VerifyRequest{TenantID: 0, Phone: "+989121234567", Code: "123456"}},
		{name: "empty phone", req: VerifyRequest{TenantID: 42, Phone: "", Code: "123456"}},
		{name: "empty code", req: VerifyRequest{TenantID: 42, Phone: "+989121234567", Code: ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeOTPStore{state: activeOTPState("123456")}
			verifyLogger := &fakeVerificationLogger{}
			service := NewService(nil, store, nil, nil, verifyLogger, Config{})

			resp, err := service.VerifyOTP(context.Background(), tt.req)

			require.Nil(t, resp)
			require.Error(t, err)
			assert.Equal(t, 0, store.getCalls)
			assert.Equal(t, 0, verifyLogger.calls)
		})
	}
}

func TestServiceVerifyOTPNotFound(t *testing.T) {
	store := &fakeOTPStore{getErr: ErrOTPNotFound}
	verifyLogger := &fakeVerificationLogger{}
	service := NewService(nil, store, nil, nil, verifyLogger, Config{})

	resp, err := service.VerifyOTP(context.Background(), VerifyRequest{
		TenantID: 42,
		Phone:    "+989121234567",
		Code:     "123456",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Verified)
	assert.Empty(t, resp.RequestID)
	assert.Equal(t, ReasonNotFound, resp.Reason)
	assertVerificationLog(t, verifyLogger, VerificationResultFailed, ReasonNotFound, "", 0)
}

func TestServiceVerifyOTPStoreGetError(t *testing.T) {
	getErr := errors.New("get failed")
	store := &fakeOTPStore{getErr: getErr}
	verifyLogger := &fakeVerificationLogger{}
	service := NewService(nil, store, nil, nil, verifyLogger, Config{})

	resp, err := service.VerifyOTP(context.Background(), VerifyRequest{
		TenantID: 42,
		Phone:    "+989121234567",
		Code:     "123456",
	})

	require.Nil(t, resp)
	assert.ErrorIs(t, err, getErr)
	assert.Equal(t, 0, verifyLogger.calls)
}

func TestServiceVerifyOTPExpired(t *testing.T) {
	state := activeOTPState("123456")
	state.ExpiresAt = time.Now().UTC()
	store := &fakeOTPStore{state: state}
	verifyLogger := &fakeVerificationLogger{}
	service := NewService(nil, store, nil, nil, verifyLogger, Config{})

	resp, err := service.VerifyOTP(context.Background(), VerifyRequest{
		TenantID: 42,
		Phone:    "+989121234567",
		Code:     "123456",
	})

	require.NoError(t, err)
	assert.False(t, resp.Verified)
	assert.Equal(t, "request-verify", resp.RequestID)
	assert.Equal(t, ReasonExpired, resp.Reason)
	assert.Equal(t, 1, store.deleteCalls)
	assert.Equal(t, 0, store.incrementCalls)
	assertVerificationLog(t, verifyLogger, VerificationResultFailed, ReasonExpired, "request-verify", state.AttemptCount)
}

func TestServiceVerifyOTPMaxAttemptsAlreadyReached(t *testing.T) {
	state := activeOTPState("123456")
	state.AttemptCount = 3
	state.MaxAttempts = 3
	store := &fakeOTPStore{state: state}
	verifyLogger := &fakeVerificationLogger{}
	service := NewService(nil, store, nil, nil, verifyLogger, Config{})

	resp, err := service.VerifyOTP(context.Background(), VerifyRequest{
		TenantID: 42,
		Phone:    "+989121234567",
		Code:     "123456",
	})

	require.NoError(t, err)
	assert.False(t, resp.Verified)
	assert.Equal(t, ReasonMaxAttemptsExceeded, resp.Reason)
	assert.Equal(t, 1, store.deleteCalls)
	assert.Equal(t, 0, store.incrementCalls)
	assertVerificationLog(t, verifyLogger, VerificationResultFailed, ReasonMaxAttemptsExceeded, "request-verify", state.AttemptCount)
}

func TestServiceVerifyOTPInvalidCodeUnderMax(t *testing.T) {
	state := activeOTPState("123456")
	state.AttemptCount = 0
	state.MaxAttempts = 3
	store := &fakeOTPStore{state: state, incrementResult: 1}
	verifyLogger := &fakeVerificationLogger{}
	service := NewService(nil, store, nil, nil, verifyLogger, Config{})

	resp, err := service.VerifyOTP(context.Background(), VerifyRequest{
		TenantID: 42,
		Phone:    "+989121234567",
		Code:     "000000",
	})

	require.NoError(t, err)
	assert.False(t, resp.Verified)
	assert.Equal(t, "request-verify", resp.RequestID)
	assert.Equal(t, ReasonInvalidCode, resp.Reason)
	assert.Equal(t, 1, store.incrementCalls)
	assert.Equal(t, 0, store.deleteCalls)
	assertVerificationLog(t, verifyLogger, VerificationResultFailed, ReasonInvalidCode, "request-verify", 1)
}

func TestServiceVerifyOTPInvalidCodeReachesMax(t *testing.T) {
	state := activeOTPState("123456")
	state.AttemptCount = 2
	state.MaxAttempts = 3
	store := &fakeOTPStore{state: state, incrementResult: 3}
	verifyLogger := &fakeVerificationLogger{}
	service := NewService(nil, store, nil, nil, verifyLogger, Config{})

	resp, err := service.VerifyOTP(context.Background(), VerifyRequest{
		TenantID: 42,
		Phone:    "+989121234567",
		Code:     "000000",
	})

	require.NoError(t, err)
	assert.False(t, resp.Verified)
	assert.Equal(t, ReasonMaxAttemptsExceeded, resp.Reason)
	assert.Equal(t, 1, store.incrementCalls)
	assert.Equal(t, 1, store.deleteCalls)
	assertVerificationLog(t, verifyLogger, VerificationResultFailed, ReasonMaxAttemptsExceeded, "request-verify", 3)
}

func TestServiceVerifyOTPIncrementError(t *testing.T) {
	incrementErr := errors.New("increment failed")
	store := &fakeOTPStore{state: activeOTPState("123456"), incrementErr: incrementErr}
	verifyLogger := &fakeVerificationLogger{}
	service := NewService(nil, store, nil, nil, verifyLogger, Config{})

	resp, err := service.VerifyOTP(context.Background(), VerifyRequest{
		TenantID: 42,
		Phone:    "+989121234567",
		Code:     "000000",
	})

	require.Nil(t, resp)
	assert.ErrorIs(t, err, incrementErr)
	assert.Equal(t, 0, verifyLogger.calls)
}

func TestServiceVerifyOTPIncrementNotFound(t *testing.T) {
	store := &fakeOTPStore{state: activeOTPState("123456"), incrementErr: ErrOTPNotFound}
	verifyLogger := &fakeVerificationLogger{}
	service := NewService(nil, store, nil, nil, verifyLogger, Config{})

	resp, err := service.VerifyOTP(context.Background(), VerifyRequest{
		TenantID: 42,
		Phone:    "+989121234567",
		Code:     "000000",
	})

	require.NoError(t, err)
	assert.False(t, resp.Verified)
	assert.Empty(t, resp.RequestID)
	assert.Equal(t, ReasonNotFound, resp.Reason)
	assertVerificationLog(t, verifyLogger, VerificationResultFailed, ReasonNotFound, "request-verify", 0)
}

func TestServiceVerifyOTPLoggerErrorDoesNotChangeResponse(t *testing.T) {
	store := &fakeOTPStore{state: activeOTPState("123456")}
	verifyLogger := &fakeVerificationLogger{err: errors.New("log failed")}
	service := NewService(nil, store, nil, nil, verifyLogger, Config{})

	resp, err := service.VerifyOTP(context.Background(), VerifyRequest{
		TenantID: 42,
		Phone:    "+989121234567",
		Code:     "123456",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Verified)
	assert.Equal(t, "request-verify", resp.RequestID)
	assert.Equal(t, 1, verifyLogger.calls)
}

func TestServiceVerifyOTPSuccessDeleteError(t *testing.T) {
	deleteErr := errors.New("delete failed")
	store := &fakeOTPStore{state: activeOTPState("123456"), deleteErr: deleteErr}
	verifyLogger := &fakeVerificationLogger{}
	service := NewService(nil, store, nil, nil, verifyLogger, Config{})

	resp, err := service.VerifyOTP(context.Background(), VerifyRequest{
		TenantID: 42,
		Phone:    "+989121234567",
		Code:     "123456",
	})

	require.Nil(t, resp)
	assert.ErrorIs(t, err, deleteErr)
	assert.Equal(t, 0, verifyLogger.calls)
}

func TestServiceVerifyOTPMaxAttemptsFallback(t *testing.T) {
	state := activeOTPState("123456")
	state.AttemptCount = 1
	state.MaxAttempts = 0
	store := &fakeOTPStore{state: state, incrementResult: 2}
	verifyLogger := &fakeVerificationLogger{}
	service := NewService(nil, store, nil, nil, verifyLogger, Config{MaxAttempts: 2})

	resp, err := service.VerifyOTP(context.Background(), VerifyRequest{
		TenantID: 42,
		Phone:    "+989121234567",
		Code:     "000000",
	})

	require.NoError(t, err)
	assert.False(t, resp.Verified)
	assert.Equal(t, ReasonMaxAttemptsExceeded, resp.Reason)
	assert.Equal(t, 1, store.incrementCalls)
	assert.Equal(t, 1, store.deleteCalls)
	assertVerificationLog(t, verifyLogger, VerificationResultFailed, ReasonMaxAttemptsExceeded, "request-verify", 2)
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

func assertVerificationLog(t *testing.T, logger *fakeVerificationLogger, result string, reason string, requestID string, attemptCount int) {
	t.Helper()

	require.Equal(t, 1, logger.calls)
	require.Len(t, logger.logs, 1)
	log := logger.logs[0]
	assert.Equal(t, requestID, log.RequestID)
	assert.Equal(t, int64(42), log.TenantID)
	assert.Equal(t, "+989121234567", log.Phone)
	assert.Equal(t, result, log.Result)
	assert.Equal(t, reason, log.Reason)
	assert.Equal(t, attemptCount, log.AttemptCount)
	assert.Equal(t, "", log.CorrelationID)
	assert.False(t, log.CreatedAt.IsZero())
}

func activeOTPState(code string) *OTPState {
	return &OTPState{
		RequestID:    "request-verify",
		TenantID:     42,
		Phone:        "+989121234567",
		CodeHash:     HashCode(code),
		AttemptCount: 0,
		MaxAttempts:  3,
		CreatedAt:    time.Now().UTC().Add(-time.Minute),
		ExpiresAt:    time.Now().UTC().Add(time.Minute),
	}
}
