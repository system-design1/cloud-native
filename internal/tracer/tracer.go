package tracer

import (
	"context"
	"fmt"
	"net/url"

	"go-backend-service/internal/config"

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

	// Setup exporter(s) based on configuration
	// Support multiple exporters (Tempo and/or Jaeger)
	if cfg.Tracing.TempoEnabled && cfg.Tracing.TempoEndpoint != "" {
		// Use OTLP HTTP exporter for Tempo
		tempoExporter, err := otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(cfg.Tracing.TempoEndpoint),
			otlptracehttp.WithInsecure(), // Use WithInsecure for development, use WithTLSClientConfig for production
		)
		if err != nil {
			return fmt.Errorf("failed to create Tempo OTLP exporter: %w", err)
		}
		exporters = append(exporters, tempoExporter)
	}

	if cfg.Tracing.JaegerEnabled && cfg.Tracing.JaegerEndpoint != "" {
		// Parse Jaeger endpoint - support both OTLP format (host:port) and HTTP format (http://host:port/path)
		jaegerEndpoint := cfg.Tracing.JaegerEndpoint
		
		// If endpoint starts with http:// or https://, extract host:port
		if parsedURL, err := url.Parse(jaegerEndpoint); err == nil && (parsedURL.Scheme == "http" || parsedURL.Scheme == "https") {
			jaegerEndpoint = parsedURL.Host
		}
		
		// Replace localhost with jaeger for Docker Compose environments
		// From inside a container, localhost refers to the container itself, not the host
		// So we need to use the service name "jaeger" to connect to Jaeger container
		// Check if endpoint contains localhost (with or without port)
		if jaegerEndpoint == "localhost" || 
		   jaegerEndpoint == "localhost:4320" || 
		   jaegerEndpoint == "localhost:4318" ||
		   (len(jaegerEndpoint) > 9 && jaegerEndpoint[:9] == "localhost:") {
			// Replace localhost with jaeger and use port 4318 (OTLP HTTP port inside container)
			jaegerEndpoint = "jaeger:4318"
		} else if jaegerEndpoint == "" || jaegerEndpoint == "jaeger" {
			// Default to OTLP port
			jaegerEndpoint = "jaeger:4318"
		}
		
		// Log the final endpoint for debugging (remove in production)
		// fmt.Printf("DEBUG: Jaeger endpoint configured as: %s\n", jaegerEndpoint)
		
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
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // Sample all traces
	}

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
