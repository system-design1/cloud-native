package tracer

import (
	"strings"

	"go-backend-service/internal/logger"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// RoutePolicySampler implements a route-based sampling policy
// It supports three behaviors: ALWAYS, RATIO, and DROP
type RoutePolicySampler struct {
	alwaysRoutes  map[string]bool
	dropRoutes    map[string]bool
	ratioRoutes   map[string]float64
	defaultPolicy string
	defaultRatio  float64
}

// NewRoutePolicySampler creates a new route-based sampler
func NewRoutePolicySampler(
	alwaysRoutes []string,
	dropRoutes []string,
	ratioRoutes map[string]float64,
	defaultPolicy string,
	defaultRatio float64,
) *RoutePolicySampler {
	// Build sets for fast lookup
	alwaysSet := make(map[string]bool)
	for _, route := range alwaysRoutes {
		route = strings.TrimSpace(route)
		if route != "" {
			alwaysSet[route] = true
		}
	}

	dropSet := make(map[string]bool)
	for _, route := range dropRoutes {
		route = strings.TrimSpace(route)
		if route != "" {
			dropSet[route] = true
		}
	}

	return &RoutePolicySampler{
		alwaysRoutes:  alwaysSet,
		dropRoutes:    dropSet,
		ratioRoutes:   ratioRoutes,
		defaultPolicy: defaultPolicy,
		defaultRatio:  defaultRatio,
	}
}

// ShouldSample implements sdktrace.Sampler interface
func (s *RoutePolicySampler) ShouldSample(params sdktrace.SamplingParameters) sdktrace.SamplingResult {
	// Extract route path from span name
	// Span name format from otelgin: "METHOD /path"
	routePath := s.extractRoutePath(params.Name)
	
	// Apply precedence rules:
	// 1. DROP (highest priority)
	if s.dropRoutes[routePath] {
		// Log at info level so we can see it's working
		log := logger.Get()
		log.Info().
			Str("span_name", params.Name).
			Str("extracted_path", routePath).
			Msg("Dropping trace for route (OTEL_ROUTE_DROP)")
		return sdktrace.SamplingResult{
			Decision: sdktrace.Drop,
		}
	}

	// Apply precedence rules:
	// 1. DROP (highest priority) - already handled above

	// 2. ALWAYS
	if s.alwaysRoutes[routePath] {
		return sdktrace.SamplingResult{
			Decision:   sdktrace.RecordAndSample,
			Attributes: []attribute.KeyValue{},
		}
	}

	// 3. RATIO
	if ratio, exists := s.ratioRoutes[routePath]; exists {
		ratioSampler := sdktrace.TraceIDRatioBased(ratio)
		return ratioSampler.ShouldSample(params)
	}

	// 4. DEFAULT policy
	switch s.defaultPolicy {
	case "drop":
		return sdktrace.SamplingResult{
			Decision: sdktrace.Drop,
		}
	case "ratio":
		ratioSampler := sdktrace.TraceIDRatioBased(s.defaultRatio)
		return ratioSampler.ShouldSample(params)
	case "always":
		fallthrough
	default:
		return sdktrace.SamplingResult{
			Decision:   sdktrace.RecordAndSample,
			Attributes: []attribute.KeyValue{},
		}
	}
}

// Description returns a description of the sampler
func (s *RoutePolicySampler) Description() string {
	return "RoutePolicySampler"
}

// extractRoutePath extracts the route path from span name
// Span name format from otelgin can be:
// - "METHOD /path" (simple format)
// - "SERVICE_NAME: METHOD /path" (with service name prefix)
// - "METHOD /path?query" (with query string)
// We need to extract just the path part
func (s *RoutePolicySampler) extractRoutePath(spanName string) string {
	// Handle format with service name prefix: "SERVICE_NAME: METHOD /path"
	// Split by ":" first to remove service name prefix if present
	if idx := strings.Index(spanName, ":"); idx != -1 {
		// Remove service name prefix (everything before ":")
		spanName = strings.TrimSpace(spanName[idx+1:])
	}

	// Split by space to separate METHOD and path
	parts := strings.Fields(spanName)
	if len(parts) < 2 {
		// If format is unexpected, try to extract path directly
		// Check if it starts with "/"
		if strings.HasPrefix(spanName, "/") {
			path := spanName
			// Remove query string if present
			if idx := strings.Index(path, "?"); idx != -1 {
				path = path[:idx]
			}
			return path
		}
		// If format is unexpected, return as-is
		return spanName
	}

	// Get the path part (second element after METHOD)
	path := parts[1]

	// Remove query string if present
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	// Normalize path (ensure it starts with /)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return path
}
