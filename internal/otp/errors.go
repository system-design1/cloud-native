package otp

import "errors"

var (
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrTenantDisabled      = errors.New("tenant disabled")
	ErrOTPAlreadyActive    = errors.New("otp already active")
	ErrOTPRateLimited      = errors.New("otp rate limited")
	ErrOTPNotFound         = errors.New("otp not found")
	ErrOTPExpired          = errors.New("otp expired")
	ErrInvalidCode         = errors.New("invalid otp code")
	ErrMaxAttemptsExceeded = errors.New("max attempts exceeded")
	ErrSMSProviderFailed   = errors.New("sms provider failed")
	ErrNotImplemented      = errors.New("otp flow not implemented")
)
