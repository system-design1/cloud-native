package otp

import "time"

// Config contains OTP domain defaults. Environment binding can be added later.
type Config struct {
	CodeLength      int
	TTL             time.Duration
	MaxAttempts     int
	TenantCacheTTL  time.Duration
	ProviderTimeout time.Duration
}

// DefaultConfig returns conservative defaults for the first real OTP flow.
func DefaultConfig() Config {
	return Config{
		CodeLength:      6,
		TTL:             2 * time.Minute,
		MaxAttempts:     3,
		TenantCacheTTL:  5 * time.Minute,
		ProviderTimeout: 2 * time.Second,
	}
}
