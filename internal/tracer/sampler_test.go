package tracer

import (
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func TestExtractRoutePath(t *testing.T) {
	sampler := &RoutePolicySampler{}

	tests := []struct {
		name     string
		spanName string
		expected string
	}{
		{
			name:     "simple format",
			spanName: "GET /health",
			expected: "/health",
		},
		{
			name:     "with service name prefix",
			spanName: "go-backend-service: GET /metrics",
			expected: "/metrics",
		},
		{
			name:     "with query string",
			spanName: "GET /health?check=1",
			expected: "/health",
		},
		{
			name:     "with service name and query string",
			spanName: "go-backend-service: GET /metrics?format=prometheus",
			expected: "/metrics",
		},
		{
			name:     "path without leading slash",
			spanName: "GET health",
			expected: "/health",
		},
		{
			name:     "HEAD method",
			spanName: "HEAD /health",
			expected: "/health",
		},
		{
			name:     "delayed-hello",
			spanName: "go-backend-service: GET /delayed-hello",
			expected: "/delayed-hello",
		},
		{
			name:     "test-error",
			spanName: "go-backend-service: GET /test-error",
			expected: "/test-error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sampler.extractRoutePath(tt.spanName)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestRoutePolicySampler_Drop(t *testing.T) {
	sampler := NewRoutePolicySampler(
		[]string{},           // no always routes
		[]string{"/metrics"}, // drop /metrics
		map[string]float64{}, // no ratio routes
		"always",             // default: always
		1.0,                  // default ratio
	)

	// Test DROP for /metrics
	params := sdktrace.SamplingParameters{
		Name:    "go-backend-service: GET /metrics",
		TraceID: trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
	}

	result := sampler.ShouldSample(params)
	if result.Decision != sdktrace.Drop {
		t.Errorf("Expected DROP decision for /metrics, got %v", result.Decision)
	}
}

func TestRoutePolicySampler_Always(t *testing.T) {
	sampler := NewRoutePolicySampler(
		[]string{"/delayed-hello", "/test-error"}, // always routes
		[]string{},           // no drop routes
		map[string]float64{}, // no ratio routes
		"always",             // default: always
		1.0,                  // default ratio
	)

	// Test ALWAYS for /delayed-hello
	params := sdktrace.SamplingParameters{
		Name:    "go-backend-service: GET /delayed-hello",
		TraceID: trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
	}

	result := sampler.ShouldSample(params)
	if result.Decision != sdktrace.RecordAndSample {
		t.Errorf("Expected RecordAndSample decision for /delayed-hello, got %v", result.Decision)
	}
}

func TestRoutePolicySampler_Precedence(t *testing.T) {
	// Test that DROP has higher precedence than ALWAYS
	sampler := NewRoutePolicySampler(
		[]string{"/metrics"}, // /metrics in ALWAYS
		[]string{"/metrics"}, // /metrics also in DROP
		map[string]float64{}, // no ratio routes
		"always",             // default: always
		1.0,                  // default ratio
	)

	// DROP should win
	params := sdktrace.SamplingParameters{
		Name:    "go-backend-service: GET /metrics",
		TraceID: trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
	}

	result := sampler.ShouldSample(params)
	if result.Decision != sdktrace.Drop {
		t.Errorf("Expected DROP decision (precedence), got %v", result.Decision)
	}
}
