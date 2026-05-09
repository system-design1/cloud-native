package otp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

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
	config = withDefaults(config)

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
	if err := validateSendRequest(req); err != nil {
		return nil, err
	}

	tenant, err := s.tenantSettings.GetTenantSettings(ctx, req.TenantID)
	if err != nil {
		return nil, err
	}
	if err := validateTenant(tenant, time.Now().UTC()); err != nil {
		return nil, err
	}

	requestID := uuid.NewString()
	code, err := GenerateCode(s.config.CodeLength)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	expiredAt := now.Add(s.config.TTL)
	state := OTPState{
		RequestID:    requestID,
		TenantID:     req.TenantID,
		Phone:        req.Phone,
		CodeHash:     HashCode(code),
		AttemptCount: 0,
		MaxAttempts:  s.config.MaxAttempts,
		CreatedAt:    now,
		ExpiresAt:    expiredAt,
	}

	if s.requestLogger != nil {
		if err := s.requestLogger.CreateRequest(ctx, OTPRequestLog{
			RequestID:     requestID,
			TenantID:      req.TenantID,
			Phone:         req.Phone,
			Status:        RequestStatusPending,
			ProviderName:  tenant.SMSProvider,
			CorrelationID: "",
			Metadata:      req.Metadata,
			CreatedAt:     now,
			UpdatedAt:     now,
		}); err != nil {
			return nil, err
		}
	}

	if err := s.store.Save(ctx, state, s.config.TTL); err != nil {
		saveErr := fmt.Errorf("save otp state: %w", err)
		s.updateProviderResult(ctx, OTPProviderResultLog{
			RequestID:    requestID,
			Status:       RequestStatusFailed,
			ProviderName: tenant.SMSProvider,
			ErrorMessage: saveErr.Error(),
			UpdatedAt:    time.Now().UTC(),
		})
		return nil, fmt.Errorf("save otp state: %w", err)
	}

	providerCtx, cancel := context.WithTimeout(ctx, s.config.ProviderTimeout)
	defer cancel()

	result, err := s.smsProvider.SendOTP(providerCtx, SMSRequest{
		RequestID: requestID,
		TenantID:  req.TenantID,
		Phone:     req.Phone,
		Code:      code,
		Provider:  tenant.SMSProvider,
		Metadata:  req.Metadata,
	})
	if err != nil {
		s.updateProviderResult(ctx, OTPProviderResultLog{
			RequestID:    requestID,
			Status:       RequestStatusFailed,
			ProviderName: tenant.SMSProvider,
			ErrorMessage: err.Error(),
			UpdatedAt:    time.Now().UTC(),
		})
		return nil, fmt.Errorf("%w: %w", ErrSMSProviderFailed, err)
	}

	if err := s.updateProviderResult(ctx, OTPProviderResultLog{
		RequestID:        requestID,
		Status:           RequestStatusSent,
		ProviderName:     tenant.SMSProvider,
		ProviderResponse: smsProviderResponse(result),
		UpdatedAt:        time.Now().UTC(),
	}); err != nil {
		return nil, err
	}

	return &SendResponse{
		RequestID: requestID,
		ExpiredAt: expiredAt,
	}, nil
}

// VerifyOTP will orchestrate Redis state lookup, attempt tracking, and verification logging.
func (s *Service) VerifyOTP(ctx context.Context, req VerifyRequest) (*VerifyResponse, error) {
	return nil, ErrNotImplemented
}

func (s *Service) updateProviderResult(ctx context.Context, log OTPProviderResultLog) error {
	if s.requestLogger == nil {
		return nil
	}
	return s.requestLogger.UpdateProviderResult(ctx, log)
}

func smsProviderResponse(result *SMSResult) map[string]interface{} {
	if result == nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"provider":   result.Provider,
		"status":     result.Status,
		"message_id": result.MessageID,
		"sent_at":    result.SentAt,
	}
}

func withDefaults(config Config) Config {
	defaults := DefaultConfig()
	if config.CodeLength == 0 {
		config.CodeLength = defaults.CodeLength
	}
	if config.TTL == 0 {
		config.TTL = defaults.TTL
	}
	if config.MaxAttempts == 0 {
		config.MaxAttempts = defaults.MaxAttempts
	}
	if config.TenantCacheTTL == 0 {
		config.TenantCacheTTL = defaults.TenantCacheTTL
	}
	if config.ProviderTimeout == 0 {
		config.ProviderTimeout = defaults.ProviderTimeout
	}
	return config
}

func validateSendRequest(req SendRequest) error {
	if req.TenantID <= 0 {
		return fmt.Errorf("tenant_id must be greater than 0")
	}
	if strings.TrimSpace(req.Phone) == "" {
		return fmt.Errorf("phone must not be empty")
	}
	return nil
}

func validateTenant(tenant *TenantSettings, now time.Time) error {
	if tenant == nil {
		return ErrTenantNotFound
	}
	if !tenant.OTPEnabled {
		return ErrTenantDisabled
	}
	if tenant.Status != "" && !strings.EqualFold(tenant.Status, "active") {
		return ErrTenantDisabled
	}
	if tenant.ExpiresAt != nil && !tenant.ExpiresAt.After(now) {
		return ErrTenantDisabled
	}
	return nil
}
