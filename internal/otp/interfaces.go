package otp

import (
	"context"
	"time"
)

// TenantSettingsProvider loads tenant configuration for OTP flows.
type TenantSettingsProvider interface {
	GetTenantSettings(ctx context.Context, tenantID int64) (*TenantSettings, error)
}

// OTPStore persists short-lived OTP verification state.
type OTPStore interface {
	Save(ctx context.Context, state OTPState, ttl time.Duration) error
	Get(ctx context.Context, tenantID int64, phone string) (*OTPState, error)
	IncrementAttempts(ctx context.Context, tenantID int64, phone string) (int, error)
	Delete(ctx context.Context, tenantID int64, phone string) error
}

// SMSProvider sends OTP codes through an external or simulated provider.
type SMSProvider interface {
	SendOTP(ctx context.Context, req SMSRequest) (*SMSResult, error)
}

// SendRateLimiter checks whether an OTP send request is allowed.
type SendRateLimiter interface {
	AllowSend(ctx context.Context, tenantID int64, phone string) error
}

// OTPRequestLogger persists send request and provider result logs.
type OTPRequestLogger interface {
	CreateRequest(ctx context.Context, log OTPRequestLog) error
	UpdateProviderResult(ctx context.Context, log OTPProviderResultLog) error
}

// OTPVerificationLogger persists verification attempt logs.
type OTPVerificationLogger interface {
	LogVerification(ctx context.Context, log OTPVerificationLog) error
}
