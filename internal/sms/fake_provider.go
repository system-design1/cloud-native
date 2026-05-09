package sms

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go-backend-service/internal/otp"
)

const fakeProviderName = "fake"

// FakeProvider simulates SMS delivery for local development and benchmarks.
type FakeProvider struct {
	minDelay time.Duration
	maxDelay time.Duration
}

// NewFakeProvider creates a fake SMS provider with realistic simulated latency.
func NewFakeProvider() *FakeProvider {
	return newFakeProviderWithDelay(20*time.Millisecond, 30*time.Millisecond)
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

	return &otp.SMSResult{
		Provider:  provider,
		Status:    otp.RequestStatusSent,
		MessageID: messageID,
		RawResponse: map[string]interface{}{
			"provider":   provider,
			"simulated":  true,
			"request_id": req.RequestID,
		},
		SentAt: time.Now().UTC(),
	}, nil
}

func (p *FakeProvider) delay() time.Duration {
	if p.maxDelay <= p.minDelay {
		return p.minDelay
	}

	delta := p.maxDelay - p.minDelay
	return p.minDelay + time.Duration(rand.Int63n(int64(delta)+1))
}
