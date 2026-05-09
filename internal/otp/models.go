package otp

import "time"

// Status constants used by the OTP send flow and persisted request logs.
const (
	RequestStatusPending  = "pending"
	RequestStatusSent     = "sent"
	RequestStatusFailed   = "failed"
	RequestStatusVerified = "verified"
)

// Verification result constants.
const (
	VerificationResultSuccess = "success"
	VerificationResultFailed  = "failed"
)

// Verification reason constants.
const (
	ReasonInvalidCode         = "invalid_code"
	ReasonExpired             = "expired"
	ReasonNotFound            = "not_found"
	ReasonMaxAttemptsExceeded = "max_attempts_exceeded"
	ReasonVerified            = "verified"
)

// SendRequest is the application-level input for sending an OTP.
type SendRequest struct {
	Phone    string                 `json:"phone"`
	TenantID int64                  `json:"tenant_id"`
	Token    string                 `json:"token,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SendResponse is returned after an OTP send request is accepted.
type SendResponse struct {
	RequestID string    `json:"request_id"`
	ExpiredAt time.Time `json:"expired_at"`
}

// VerifyRequest is the application-level input for verifying an OTP.
type VerifyRequest struct {
	TenantID int64  `json:"tenant_id"`
	Phone    string `json:"phone"`
	Code     string `json:"code"`
}

// VerifyResponse represents the outcome of an OTP verification attempt.
type VerifyResponse struct {
	Verified  bool   `json:"verified"`
	RequestID string `json:"request_id,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// TenantSettings contains the subset of tenant configuration needed by OTP flows.
type TenantSettings struct {
	ID              int64                  `json:"id"`
	TenantCode      string                 `json:"tenant_code"`
	Name            string                 `json:"name"`
	Status          string                 `json:"status"`
	OTPEnabled      bool                   `json:"otp_enabled"`
	SMSProvider     string                 `json:"sms_provider"`
	RateLimitPerMin int                    `json:"rate_limit_per_min"`
	Timezone        string                 `json:"timezone"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	ExpiresAt       *time.Time             `json:"expires_at,omitempty"`
}

// OTPState is the Redis-backed verification state. CodeHash must never contain plaintext OTP.
type OTPState struct {
	RequestID    string    `json:"request_id"`
	TenantID     int64     `json:"tenant_id"`
	Phone        string    `json:"phone"`
	CodeHash     string    `json:"code_hash"`
	AttemptCount int       `json:"attempt_count"`
	MaxAttempts  int       `json:"max_attempts"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// SMSRequest is sent to an SMS provider adapter.
type SMSRequest struct {
	RequestID string                 `json:"request_id"`
	TenantID  int64                  `json:"tenant_id"`
	Phone     string                 `json:"phone"`
	Code      string                 `json:"code"`
	Provider  string                 `json:"provider"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// SMSResult describes the provider response in a transport-neutral shape.
type SMSResult struct {
	Provider    string                 `json:"provider"`
	Status      string                 `json:"status"`
	MessageID   string                 `json:"message_id,omitempty"`
	RawResponse map[string]interface{} `json:"raw_response,omitempty"`
	SentAt      time.Time              `json:"sent_at"`
}

// OTPRequestLog captures the reportable send request data for PostgreSQL.
type OTPRequestLog struct {
	RequestID     string                 `json:"request_id"`
	TenantID      int64                  `json:"tenant_id"`
	Phone         string                 `json:"phone"`
	Status        string                 `json:"status"`
	ProviderName  string                 `json:"provider_name"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// OTPProviderResultLog captures the provider result for a previously created request.
type OTPProviderResultLog struct {
	RequestID        string                 `json:"request_id"`
	Status           string                 `json:"status"`
	ProviderName     string                 `json:"provider_name"`
	ProviderResponse map[string]interface{} `json:"provider_response,omitempty"`
	ErrorMessage     string                 `json:"error_message,omitempty"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// OTPVerificationLog captures the reportable verification attempt data.
type OTPVerificationLog struct {
	RequestID     string    `json:"request_id"`
	TenantID      int64     `json:"tenant_id"`
	Phone         string    `json:"phone"`
	Result        string    `json:"result"`
	Reason        string    `json:"reason,omitempty"`
	AttemptCount  int       `json:"attempt_count"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
