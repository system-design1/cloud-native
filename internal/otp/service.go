package otp

import "context"

// Service coordinates the OTP send and verification use cases.
type Service struct {
	tenantSettings TenantSettingsProvider
	store          OTPStore
	smsProvider    SMSProvider
	requestLogger  OTPRequestLogger
	verifyLogger   OTPVerificationLogger
	config         Config
}

// NewService creates an OTP service with domain dependencies and defaults.
func NewService(
	tenantSettings TenantSettingsProvider,
	store OTPStore,
	smsProvider SMSProvider,
	requestLogger OTPRequestLogger,
	verifyLogger OTPVerificationLogger,
	config Config,
) *Service {
	if config.CodeLength == 0 {
		config = DefaultConfig()
	}

	return &Service{
		tenantSettings: tenantSettings,
		store:          store,
		smsProvider:    smsProvider,
		requestLogger:  requestLogger,
		verifyLogger:   verifyLogger,
		config:         config,
	}
}

// SendOTP will orchestrate tenant lookup, OTP storage, provider send, and logging.
func (s *Service) SendOTP(ctx context.Context, req SendRequest) (*SendResponse, error) {
	return nil, ErrNotImplemented
}

// VerifyOTP will orchestrate Redis state lookup, attempt tracking, and verification logging.
func (s *Service) VerifyOTP(ctx context.Context, req VerifyRequest) (*VerifyResponse, error) {
	return nil, ErrNotImplemented
}
