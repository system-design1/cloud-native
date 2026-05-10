package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"go-backend-service/internal/otp"

	"github.com/redis/go-redis/v9"
)

const (
	fakeProviderName      = "fake"
	defaultDebugCodeTTL   = time.Minute
	debugCodeKeyNamespace = "debug:otp-code"
)

// FakeProvider simulates SMS delivery for local development and benchmarks.
type FakeProvider struct {
	minDelay        time.Duration
	maxDelay        time.Duration
	debugCodeClient *redis.Client
	debugCodeTTL    time.Duration
}

type debugCodeValue struct {
	RequestID string    `json:"request_id"`
	TenantID  int64     `json:"tenant_id"`
	Phone     string    `json:"phone"`
	Code      string    `json:"code"`
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
}

// NewFakeProvider creates a fake SMS provider with realistic simulated latency.
func NewFakeProvider() *FakeProvider {
	return NewFakeProviderWithDelay(20*time.Millisecond, 30*time.Millisecond)
}

// NewFakeProviderWithDelay creates a fake SMS provider with configurable simulated latency.
func NewFakeProviderWithDelay(minDelay, maxDelay time.Duration) *FakeProvider {
	return newFakeProviderWithDelay(minDelay, maxDelay)
}

// NewFakeProviderWithDebugCodeCapture creates a fake provider that best-effort
// stores plaintext OTP codes in Redis for local manual testing only.
func NewFakeProviderWithDebugCodeCapture(client *redis.Client, ttl time.Duration) *FakeProvider {
	provider := NewFakeProvider()
	provider.enableDebugCodeCapture(client, ttl)
	return provider
}

// NewFakeProviderWithDelayAndDebugCodeCapture creates a fake provider with
// configurable latency and local-only Redis debug code capture.
func NewFakeProviderWithDelayAndDebugCodeCapture(minDelay, maxDelay time.Duration, client *redis.Client, ttl time.Duration) *FakeProvider {
	provider := NewFakeProviderWithDelay(minDelay, maxDelay)
	provider.enableDebugCodeCapture(client, ttl)
	return provider
}

func (p *FakeProvider) enableDebugCodeCapture(client *redis.Client, ttl time.Duration) {
	p.debugCodeClient = client
	p.debugCodeTTL = ttl
	if p.debugCodeTTL <= 0 {
		p.debugCodeTTL = defaultDebugCodeTTL
	}
}

func newFakeProviderWithDelay(minDelay, maxDelay time.Duration) *FakeProvider {
	return &FakeProvider{
		minDelay: minDelay,
		maxDelay: maxDelay,
	}
}

// SendOTP simulates sending an OTP through an SMS provider.
func (p *FakeProvider) SendOTP(ctx context.Context, req otp.SMSRequest) (*otp.SMSResult, error) {
	delay := p.delay()
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("fake sms provider canceled: %w", ctx.Err())
	case <-timer.C:
	}

	provider := req.Provider
	if provider == "" {
		provider = fakeProviderName
	}

	messageID := req.RequestID
	if messageID == "" {
		messageID = fmt.Sprintf("fake-%d-%d", time.Now().UTC().UnixNano(), rand.Int63())
	} else {
		messageID = "fake-" + messageID
	}

	sentAt := time.Now().UTC()
	p.captureDebugCode(ctx, req, provider, sentAt)

	return &otp.SMSResult{
		Provider:  provider,
		Status:    otp.RequestStatusSent,
		MessageID: messageID,
		RawResponse: map[string]interface{}{
			"provider":   provider,
			"simulated":  true,
			"request_id": req.RequestID,
		},
		SentAt: sentAt,
	}, nil
}

func (p *FakeProvider) captureDebugCode(ctx context.Context, req otp.SMSRequest, provider string, createdAt time.Time) {
	if p.debugCodeClient == nil {
		return
	}

	value := debugCodeValue{
		RequestID: req.RequestID,
		TenantID:  req.TenantID,
		Phone:     req.Phone,
		Code:      req.Code,
		Provider:  provider,
		CreatedAt: createdAt,
	}
	data, err := json.Marshal(value)
	if err != nil {
		return
	}

	_ = p.debugCodeClient.Set(ctx, debugCodeKey(req.TenantID, req.Phone), data, p.debugCodeTTL).Err()
}

func debugCodeKey(tenantID int64, phone string) string {
	return fmt.Sprintf("%s:%d:%s", debugCodeKeyNamespace, tenantID, phone)
}

func (p *FakeProvider) delay() time.Duration {
	if p.maxDelay <= p.minDelay {
		return p.minDelay
	}

	delta := p.maxDelay - p.minDelay
	return p.minDelay + time.Duration(rand.Int63n(int64(delta)+1))
}
