package middleware

import (
	"go-backend-service/internal/tracer"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/gin-gonic/gin"
)

// TracingMiddleware sets up OpenTelemetry tracing for HTTP requests
func TracingMiddleware() gin.HandlerFunc {
	return otelgin.Middleware("go-backend-service")
}

// TracingContextMiddleware extracts trace and span IDs from context and adds them to Gin context
// This should be called after TracingMiddleware
func TracingContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Get trace ID and span ID from OpenTelemetry context
		traceID := tracer.TraceIDFromContext(ctx)
		spanID := tracer.SpanIDFromContext(ctx)

		// Add trace ID and span ID to Gin context
		if traceID != "" {
			c.Set("trace_id", traceID)
		}
		if spanID != "" {
			c.Set("span_id", spanID)
		}

		// Get correlation ID (should be set by CorrelationIDMiddleware)
		correlationID, exists := c.Get("correlation_id")
		if exists && correlationID != nil {
			// Add correlation ID as attribute to the current span
			span := trace.SpanFromContext(ctx)
			if span.IsRecording() {
				span.SetAttributes(
					attribute.String("correlation_id", correlationID.(string)),
				)
			}
		}

		// Add additional attributes to span
		span := trace.SpanFromContext(ctx)
		if span.IsRecording() {
			span.SetAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.String("http.url", c.Request.URL.String()),
				attribute.String("http.scheme", c.Request.URL.Scheme),
				attribute.String("http.host", c.Request.Host),
				attribute.String("http.user_agent", c.Request.UserAgent()),
				attribute.String("http.remote_addr", c.ClientIP()),
			)
		}

		c.Next()

		// After request is processed, update span name and add response attributes
		span = trace.SpanFromContext(c.Request.Context())
		if span.IsRecording() {
			// Update span name to include route path (now available after route matching)
			route := c.FullPath()
			if route == "" {
				// If route is not matched (404), use the request path
				route = c.Request.URL.Path
			}
			// Update span name to "METHOD /path" format
			spanName := c.Request.Method + " " + route
			span.SetName(spanName)
			
			span.SetAttributes(
				attribute.String("http.route", route),
				attribute.Int("http.status_code", c.Writer.Status()),
				attribute.Int("http.response.size", c.Writer.Size()),
			)
		}
	}
}
