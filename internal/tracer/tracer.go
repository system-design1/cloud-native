package tracer

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"go-backend-service/internal/config"
	"go-backend-service/internal/logger"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracerProvider *sdktrace.TracerProvider
	tracer         trace.Tracer
)

// Init initializes OpenTelemetry tracing
func Init(cfg *config.Config) error {
	if !cfg.Tracing.Enabled {
		// If tracing is disabled, use a no-op tracer
		tracerProvider = sdktrace.NewTracerProvider()
		tracer = tracerProvider.Tracer(cfg.Tracing.ServiceName)
		return nil
	}

	ctx := context.Background()

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.Tracing.ServiceName),
			semconv.ServiceVersion(cfg.Tracing.ServiceVersion),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	var exporters []sdktrace.SpanExporter

	// Helper function to normalize endpoint for Docker
	normalizeEndpoint := func(endpoint string, defaultService string) string {
		originalEndpoint := endpoint
		
		// If endpoint starts with http:// or https://, extract host:port
		if parsedURL, err := url.Parse(endpoint); err == nil && (parsedURL.Scheme == "http" || parsedURL.Scheme == "https") {
			endpoint = parsedURL.Host
		}
		
		// Detect if we're running in Docker by checking HOSTNAME
		// In Docker containers, HOSTNAME is usually a container ID (12 hex chars) or container name
		isDockerEnv := false
		hostname := os.Getenv("HOSTNAME")
		if hostname != "" {
			// Container IDs are 12 hex characters (like "78dec8ce8710")
			// Check if it's exactly 12 chars and looks like hex
			if len(hostname) == 12 {
				// Simple check: if it's 12 chars, it's likely a container ID
				isDockerEnv = true
			} else if strings.Contains(hostname, "-") && len(hostname) < 30 {
				// Container names usually contain hyphens (like "go-backend-api")
				isDockerEnv = true
			}
		}
		
		// Debug log (always log to see what's happening)
		log := logger.Get()
		log.Info().
			Str("endpoint", endpoint).
			Str("hostname", hostname).
			Bool("is_docker", isDockerEnv).
			Str("service", defaultService).
			Msg("Normalizing endpoint")
		
		// If in Docker and endpoint contains localhost, replace with service name
		if isDockerEnv && strings.Contains(endpoint, "localhost") {
			// Replace localhost with service name, keep port
			if strings.Contains(endpoint, ":") {
				parts := strings.Split(endpoint, ":")
				if len(parts) == 2 {
					endpoint = defaultService + ":" + parts[1]
				} else {
					endpoint = defaultService + ":4318"
				}
			} else {
				endpoint = defaultService + ":4318"
			}
			
			// Log the normalization
			log := logger.Get()
			log.Info().
				Str("original_endpoint", originalEndpoint).
				Str("normalized_endpoint", endpoint).
				Str("service", defaultService).
				Str("hostname", hostname).
				Bool("is_docker", isDockerEnv).
				Msg("Endpoint normalized for Docker")
		}
		
		return endpoint
	}

	// Setup exporter(s) based on configuration
	// Support multiple exporters (Tempo and/or Jaeger)
	if cfg.Tracing.TempoEnabled && cfg.Tracing.TempoEndpoint != "" {
		tempoEndpoint := normalizeEndpoint(cfg.Tracing.TempoEndpoint, "tempo")
		tempoExporter, err := otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(tempoEndpoint),
			otlptracehttp.WithInsecure(), // Use WithInsecure for development, use WithTLSClientConfig for production
		)
		if err != nil {
			return fmt.Errorf("failed to create Tempo OTLP exporter: %w", err)
		}
		exporters = append(exporters, tempoExporter)
	}

	if cfg.Tracing.JaegerEnabled && cfg.Tracing.JaegerEndpoint != "" {
		jaegerEndpoint := normalizeEndpoint(cfg.Tracing.JaegerEndpoint, "jaeger")
		jaegerExporter, err := otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(jaegerEndpoint),
			otlptracehttp.WithInsecure(), // Use WithInsecure for development
		)
		if err != nil {
			return fmt.Errorf("failed to create Jaeger OTLP exporter: %w", err)
		}
		exporters = append(exporters, jaegerExporter)
	}

	// If no exporters configured, use console exporter
	if len(exporters) == 0 {
		exporters = append(exporters, &consoleExporter{})
	}

	// Create tracer provider with all exporters
	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
	}

	// Configure sampler based on route policy
	var sampler sdktrace.Sampler
	log := logger.Get()
	if cfg.Tracing.RoutePolicy.Enabled {
		// Use route-based policy sampler
		log.Info().
			Strs("always_routes", cfg.Tracing.RoutePolicy.AlwaysRoutes).
			Strs("drop_routes", cfg.Tracing.RoutePolicy.DropRoutes).
			Str("default_policy", cfg.Tracing.RoutePolicy.DefaultPolicy).
			Msg("Route-based tracing policy enabled")
		
		routeSampler := NewRoutePolicySampler(
			cfg.Tracing.RoutePolicy.AlwaysRoutes,
			cfg.Tracing.RoutePolicy.DropRoutes,
			cfg.Tracing.RoutePolicy.RatioRoutes,
			cfg.Tracing.RoutePolicy.DefaultPolicy,
			cfg.Tracing.RoutePolicy.DefaultRatio,
		)
		// Use ParentBased to ensure parent-child consistency
		// For root spans (HTTP requests without parent), ParentBased uses the root sampler (routeSampler)
		// For child spans, if parent is sampled, child is also sampled (preserves trace integrity)
		sampler = sdktrace.ParentBased(routeSampler)
	} else {
		// Default behavior: sample all traces
		// When route policy is disabled, use AlwaysSample directly
		// ParentBased is not needed here since we want to sample everything
		log.Info().Msg("Route-based tracing policy disabled, using AlwaysSample")
		sampler = sdktrace.AlwaysSample()
	}

	opts = append(opts, sdktrace.WithSampler(sampler))

	// Add batchers for each exporter
	for _, exporter := range exporters {
		opts = append(opts, sdktrace.WithBatcher(exporter))
	}

	tracerProvider = sdktrace.NewTracerProvider(opts...)

	// Set global tracer provider
	otel.SetTracerProvider(tracerProvider)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Get tracer
	tracer = otel.Tracer(cfg.Tracing.ServiceName)

	return nil
}

// Shutdown gracefully shuts down the tracer provider
func Shutdown(ctx context.Context) error {
	if tracerProvider != nil {
		return tracerProvider.Shutdown(ctx)
	}
	return nil
}

// GetTracer returns the global tracer instance
func GetTracer() trace.Tracer {
	if tracer == nil {
		// Return a no-op tracer if not initialized
		return trace.NewNoopTracerProvider().Tracer("noop")
	}
	return tracer
}

// StartSpan starts a new span with the given name and options
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return GetTracer().Start(ctx, name, opts...)
}

// SpanFromContext returns the span from the context
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// TraceIDFromContext extracts trace ID from context as string
func TraceIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// SpanIDFromContext extracts span ID from context as string
func SpanIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// AddSpanAttributes adds attributes to the current span
func AddSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// consoleExporter is a simple console exporter for development
type consoleExporter struct{}

func (e *consoleExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	// In development, we can log spans to console
	// In production, use OTLP exporter to send to Tempo/Jaeger/etc.
	return nil
}

func (e *consoleExporter) Shutdown(ctx context.Context) error {
	return nil
}
