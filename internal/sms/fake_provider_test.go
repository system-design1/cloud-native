package sms

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-backend-service/internal/otp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFakeProviderSendOTPSuccess(t *testing.T) {
	provider := newFakeProviderWithDelay(0, 0)
	req := otp.SMSRequest{
		RequestID: "request-success",
		TenantID:  123,
		Phone:     "+989121234567",
		Code:      "123456",
	}

	result, err := provider.SendOTP(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, fakeProviderName, result.Provider)
	assert.Equal(t, otp.RequestStatusSent, result.Status)
	assert.Equal(t, "fake-"+req.RequestID, result.MessageID)
	assert.False(t, result.SentAt.IsZero())
	require.NotNil(t, result.RawResponse)
	assert.Equal(t, fakeProviderName, result.RawResponse["provider"])
	assert.Equal(t, true, result.RawResponse["simulated"])
	assert.Equal(t, req.RequestID, result.RawResponse["request_id"])
}

func TestFakeProviderRawResponseDoesNotExposeOTPCode(t *testing.T) {
	provider := newFakeProviderWithDelay(0, 0)
	req := otp.SMSRequest{
		RequestID: "request-safe-response",
		TenantID:  123,
		Phone:     "+989121234567",
		Code:      "654321",
	}

	result, err := provider.SendOTP(context.Background(), req)

	require.NoError(t, err)
	for key, value := range result.RawResponse {
		assert.NotEqual(t, "code", key)
		assert.NotEqual(t, req.Code, value)
	}
}

func TestFakeProviderUsesRequestProvider(t *testing.T) {
	provider := newFakeProviderWithDelay(0, 0)
	req := otp.SMSRequest{
		RequestID: "request-provider",
		Provider:  "kavenegar",
		Code:      "123456",
	}

	result, err := provider.SendOTP(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, req.Provider, result.Provider)
	assert.Equal(t, req.Provider, result.RawResponse["provider"])
}

func TestFakeProviderCanceledContext(t *testing.T) {
	provider := newFakeProviderWithDelay(10*time.Millisecond, 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := provider.SendOTP(ctx, otp.SMSRequest{Code: "123456"})

	require.Nil(t, result)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestFakeProviderTimeoutContext(t *testing.T) {
	provider := newFakeProviderWithDelay(50*time.Millisecond, 50*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	result, err := provider.SendOTP(ctx, otp.SMSRequest{Code: "123456"})

	require.Nil(t, result)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))
}

func TestNewFakeProviderDefaultLatencyRange(t *testing.T) {
	provider := NewFakeProvider()

	assert.Equal(t, 20*time.Millisecond, provider.minDelay)
	assert.Equal(t, 30*time.Millisecond, provider.maxDelay)
}
