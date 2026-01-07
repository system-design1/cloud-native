# Load Testing Tracing Sampling Guide

This document explains how to prevent Jaeger/Tempo UI from becoming unresponsive during high-throughput k6 load tests by adjusting OpenTelemetry sampling policies.

---

## Table of Contents

- [Problem Statement](#problem-statement)
- [Why This Happens](#why-this-happens)
- [Symptoms](#symptoms)
- [Solution: Load Test Sampling Policy](#solution-load-test-sampling-policy)
- [Configuration Example](#configuration-example)
- [How to Apply](#how-to-apply)
- [Validation Steps](#validation-steps)
- [Load Testing Checklist](#load-testing-checklist)
- [Reverting to Normal Sampling](#reverting-to-normal-sampling)

---

## Problem Statement

During high-throughput k6 load tests against endpoints like `/hello` (e.g., ~10k–13k RPS on a local machine), the Jaeger UI may become unresponsive:

- **Endless loading spinner** when trying to view traces
- **Query timeouts** when searching for traces
- **Missing traces** in the UI despite successful HTTP requests
- **High disk usage** for trace storage

**Important:** The load test itself may still look "healthy" (0% HTTP errors, acceptable latency), but observability tools become unusable due to trace overload.

---

## Why This Happens

When tracing is sampled too aggressively for high-traffic routes:

1. **Trace Volume Explosion:**
   - At 10k RPS with 100% sampling, the service generates **600,000 traces per minute**
   - Over a 5-minute test: **3 million traces**
   - Each trace contains multiple spans (HTTP request, middleware, handlers, etc.)

2. **Tracing Pipeline Overload:**
   - **OTel Exporter/Collector:** Cannot keep up with the ingestion rate
   - **Tempo/Jaeger Storage:** Disk I/O becomes a bottleneck
   - **Query Engine:** Cannot process queries in reasonable time
   - **Network:** High bandwidth usage for trace export

3. **Result:**
   - Ingestion backpressure / dropped spans
   - Query timeouts in Jaeger UI
   - Increased disk usage for trace storage
   - Observability becomes unusable

---

## Symptoms

If you experience any of the following during or after a load test, you likely have a trace overload issue:

- ✅ **Jaeger UI:** Endless loading spinner, no traces appear
- ✅ **Jaeger UI:** Query timeouts when searching for traces
- ✅ **Jaeger UI:** "No traces found" despite successful HTTP requests
- ✅ **Tempo UI:** Similar symptoms as Jaeger
- ✅ **Disk Usage:** Rapid increase in disk usage for trace storage
- ✅ **Logs:** Warnings about dropped spans or export failures
- ✅ **Load Test:** Still shows 0% errors and acceptable latency (the service is fine, but observability is broken)

---

## Solution: Load Test Sampling Policy

The solution is to apply a **route-based sampling policy** that reduces trace volume for high-traffic endpoints while keeping full tracing for error-focused endpoints.

### Recommended Strategy for Load Tests

| Endpoint | Sampling Strategy | Reason |
|----------|------------------|--------|
| `/hello` | **DROP** or **RATIO=0.001** (0.1%) | High-volume endpoint during load tests |
| `/health` | **DROP** or **RATIO=0.001** (0.1%) | High-frequency health checks |
| `/ready` | **DROP** or **RATIO=0.001** (0.1%) | High-frequency readiness checks |
| `/live` | **DROP** or **RATIO=0.001** (0.1%) | High-frequency liveness checks |
| `/metrics` | **DROP** | Prometheus scraping (no need to trace) |
| `/test-error` | **ALWAYS** (100%) | Error-focused endpoint (keep full tracing) |
| `/delayed-hello` | **ALWAYS** (100%) | Low-volume, important for debugging |

### Policy Precedence

The route-based sampling policy applies rules in the following order (highest priority first):

1. **DROP** (highest priority)
2. **ALWAYS**
3. **RATIO**
4. **DEFAULT** policy

If a route appears in multiple lists, **DROP** takes precedence.

---

## Configuration Example

This project uses the **Route-Based Tracing Policy** implemented in Task 11. The following configuration is optimized for load testing:

### Option 1: Drop High-Volume Endpoints (Recommended)

This configuration **drops** all traces for high-volume endpoints:

```env
# Enable route-based policy
OTEL_ROUTE_POLICY_ENABLED=true

# Always trace error-focused endpoints
OTEL_ROUTE_ALWAYS=/test-error,/delayed-hello

# Drop high-volume endpoints
OTEL_ROUTE_DROP=/hello,/health,/ready,/live,/metrics

# Default policy for other routes (drop to minimize noise)
OTEL_ROUTE_DEFAULT=drop

# Default ratio (not used when default is 'drop')
OTEL_ROUTE_DEFAULT_RATIO=1.0
```

**Result:**
- `/hello`, `/health`, `/ready`, `/live`, `/metrics`: **0% sampling** (no traces)
- `/test-error`, `/delayed-hello`: **100% sampling** (all traces)
- Other routes: **0% sampling** (no traces)

### Option 2: Sample at 0.1% for High-Volume Endpoints

This configuration **samples at 0.1%** for high-volume endpoints (useful if you want to see some traces):

```env
# Enable route-based policy
OTEL_ROUTE_POLICY_ENABLED=true

# Always trace error-focused endpoints
OTEL_ROUTE_ALWAYS=/test-error,/delayed-hello

# Drop metrics endpoint
OTEL_ROUTE_DROP=/metrics

# Sample high-volume endpoints at 0.1% (0.001)
OTEL_ROUTE_RATIO=/hello=0.001,/health=0.001,/ready=0.001,/live=0.001

# Default policy for other routes (drop to minimize noise)
OTEL_ROUTE_DEFAULT=drop

# Default ratio (not used when default is 'drop')
OTEL_ROUTE_DEFAULT_RATIO=1.0
```

**Result:**
- `/hello`, `/health`, `/ready`, `/live`: **0.1% sampling** (~1 trace per 1000 requests)
- `/metrics`: **0% sampling** (no traces)
- `/test-error`, `/delayed-hello`: **100% sampling** (all traces)
- Other routes: **0% sampling** (no traces)

### Comparison: Drop vs. 0.1% Sampling

| Strategy | Trace Volume (10k RPS) | Jaeger UI | Use Case |
|----------|------------------------|-----------|----------|
| **DROP** | 0 traces/min | ✅ Always responsive | Maximum performance, no traces for high-volume routes |
| **0.1% RATIO** | ~6 traces/min | ✅ Responsive | Want to see some traces, still very low volume |

**Recommendation:** Use **DROP** for maximum performance during load tests. Use **0.1% RATIO** only if you need to see some traces for high-volume endpoints.

---

## How to Apply

### Step 1: Update `.env` File

Edit your `.env` file and update the route-based tracing policy configuration:

```env
# Load Test Sampling Policy
OTEL_ROUTE_POLICY_ENABLED=true
OTEL_ROUTE_ALWAYS=/test-error,/delayed-hello
OTEL_ROUTE_DROP=/hello,/health,/ready,/live,/metrics
OTEL_ROUTE_DEFAULT=drop
OTEL_ROUTE_DEFAULT_RATIO=1.0
```

### Step 2: Restart the Application

If running with Docker Compose:

```bash
make docker-up-rebuild
```

If running locally with `air`:

```bash
# Stop current instance (Ctrl+C)
# Restart
make dev-run
```

If running locally without `air`:

```bash
# Stop current instance (Ctrl+C)
# Rebuild and run
go run cmd/server/main.go
```

### Step 3: Verify Configuration

Check that the configuration is loaded correctly:

```bash
# For Docker
docker-compose exec api printenv | grep OTEL_ROUTE

# For local
printenv | grep OTEL_ROUTE
```

You should see:

```
OTEL_ROUTE_POLICY_ENABLED=true
OTEL_ROUTE_ALWAYS=/test-error,/delayed-hello
OTEL_ROUTE_DROP=/hello,/health,/ready,/live,/metrics
OTEL_ROUTE_DEFAULT=drop
OTEL_ROUTE_DEFAULT_RATIO=1.0
```

---

## Validation Steps

### 1. Run a Load Test

Run your k6 load test against the `/hello` endpoint:

```bash
k6 run k6/03-hello-concurrency-no-sleep.js
```

**Expected:** The load test should complete successfully with 0% errors.

### 2. Check Jaeger UI Responsiveness

1. Open Jaeger UI: `http://localhost:16686`
2. Select service: `go-backend-service`
3. Click **"Find Traces"**

**Expected:** 
- ✅ Jaeger UI loads **quickly** (no endless spinner)
- ✅ Query completes in **< 5 seconds**
- ✅ UI remains responsive

### 3. Verify Traces for `/test-error`

1. In Jaeger UI, search for operation: `GET /test-error`
2. Make a test request to `/test-error`:

```bash
curl http://localhost:8080/test-error
```

3. Click **"Find Traces"** again

**Expected:**
- ✅ Traces for `/test-error` **appear** in Jaeger
- ✅ Full trace details are available

### 4. Verify No Traces for `/hello` (if using DROP)

1. In Jaeger UI, search for operation: `GET /hello`
2. Click **"Find Traces"**

**Expected (if using DROP):**
- ✅ **No traces** for `/hello` appear in Jaeger
- ✅ Or very few traces if using 0.1% RATIO

### 5. Monitor Disk Usage (Optional)

Monitor disk usage for trace storage:

```bash
# For Docker
docker stats

# Check Tempo/Jaeger container disk usage
docker exec -it tempo du -sh /var/tempo
docker exec -it jaeger du -sh /tmp
```

**Expected:**
- ✅ Disk usage **remains stable** during load test
- ✅ No rapid increase in disk usage

---

## Load Testing Checklist

Before running a high-throughput load test, ensure:

- [ ] **Load test sampling policy is enabled** in `.env`
- [ ] **High-volume endpoints are dropped or sampled at 0.1%**
- [ ] **Error-focused endpoints are set to ALWAYS**
- [ ] **Application is restarted** with new configuration
- [ ] **Configuration is verified** (check environment variables)
- [ ] **Jaeger/Tempo are running** and accessible
- [ ] **Log verbosity is reduced** (optional, set `LOG_LEVEL=warn` for load tests)
- [ ] **CPU and disk usage are monitored** (optional, but recommended)

### Optional: Reduce Log Verbosity

During load tests, you may want to reduce log verbosity to minimize I/O:

```env
LOG_LEVEL=warn
```

This reduces log volume while keeping important warnings and errors.

---

## Reverting to Normal Sampling

After completing the load test, revert to normal sampling for development/production:

### Option 1: Use Default Demo-Friendly Policy

```env
# Default demo-friendly policy
OTEL_ROUTE_POLICY_ENABLED=true
OTEL_ROUTE_ALWAYS=/delayed-hello,/test-error
OTEL_ROUTE_DROP=/metrics
OTEL_ROUTE_RATIO=/health=0.01,/live=0.01,/ready=0.01
OTEL_ROUTE_DEFAULT=always
OTEL_ROUTE_DEFAULT_RATIO=1.0
```

**Result:**
- `/delayed-hello`, `/test-error`: **100% sampling**
- `/health`, `/live`, `/ready`: **1% sampling**
- `/metrics`: **0% sampling**
- Other routes (including `/hello`): **100% sampling**

### Option 2: Disable Route Policy (Sample Everything)

```env
# Disable route-based policy (sample all traces)
OTEL_ROUTE_POLICY_ENABLED=false
```

**Result:**
- All routes: **100% sampling** (default OpenTelemetry behavior)

### Restart Application

After updating `.env`, restart the application:

```bash
# Docker Compose
make docker-up-rebuild

# Local with air
make dev-run
```

---

## Summary

- **Problem:** High-throughput load tests generate millions of traces, overwhelming Jaeger/Tempo UI
- **Solution:** Apply route-based sampling policy to drop or sample at 0.1% for high-volume endpoints
- **Configuration:** Use `OTEL_ROUTE_DROP` or `OTEL_ROUTE_RATIO` for high-volume routes
- **Validation:** Verify Jaeger UI remains responsive and traces exist for error-focused endpoints
- **Revert:** After load test, revert to normal sampling policy

**Remember:** This is a **load-testing-focused** sampling policy and may differ from production sampling policies. Always adjust sampling based on your specific needs.

---

## Related Documentation

- [Route-Based Tracing Policy](./OBSERVABILITY.md#route-based-tracing-policy-alwaysratiodrop) - General guide for route-based tracing
- [k6 Load Testing Guide](./LOAD_TESTING_K6_HELLO_CONCURRENCY.md) - Guide for running k6 load tests
- [Observability Stack](./OBSERVABILITY.md) - Complete observability setup guide

---

## References

- [OpenTelemetry Sampling](https://opentelemetry.io/docs/specs/otel/trace/sdk/#sampling)
- [Jaeger Performance Tuning](https://www.jaegertracing.io/docs/latest/performance-tuning/)
- [Tempo Performance](https://grafana.com/docs/tempo/latest/operations/performance/)

