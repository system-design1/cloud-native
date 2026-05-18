# OTP Benchmark Platform — Architecture & System Design Document

## 1. Project Overview

This project is a multi-tenant OTP (One-Time Password) delivery platform designed not only as a production-grade SaaS backend, but also as a long-term system design and benchmarking playground.

The primary goal of the project is to implement, benchmark, evolve, and compare different system design strategies in a realistic OTP provider architecture.

The platform is intentionally designed to evolve gradually from a simple monolithic REST service into a more scalable and distributed architecture over time.

The project focuses heavily on:

- Scalability
- Observability
- Reliability
- Latency analysis
- Benchmark-driven engineering
- Architecture evolution
- Production readiness
- Operational visibility

---

## 2. Business Context

The platform acts as an OTP provider service for multiple tenants.

Each tenant is a customer who has purchased the OTP delivery service.

Tenants send requests to the platform in order to:

- Send OTP codes to mobile numbers
- Verify previously generated OTP codes
- Retrieve OTP-related reports in the future

The system is designed as a SaaS platform.

---

## 3. Initial Product Scope

Current scope intentionally remains minimal.

The following assumptions are currently valid:

- Tenants are already registered.
- Tenant configuration data already exists in PostgreSQL.
- Authentication/token validation is temporarily skipped.
- SMS delivery providers are currently mocked/simulated.
- Observability stack is already implemented and operational.

Future versions will include:

- Tenant authentication
- JWT/API key validation
- Rate limiting
- Multi-provider SMS routing
- Distributed queues
- Retry pipelines
- Analytics/reporting
- Advanced security policies
- Horizontal scalability improvements

---

## 4. High-Level System Goals

The platform must support:

| Requirement | Goal |
|---|---|
| Multi-tenancy | Support many independent customers |
| Scalability | Scale from 10K tenants to 100K+ |
| Low latency | Fast OTP generation and verification |
| High observability | Full tracing, logging, and metrics |
| Benchmark-driven design | Measure bottlenecks continuously |
| Evolvable architecture | Support future distributed evolution |
| Production readiness | Graceful shutdown, health checks, telemetry |
| Operational visibility | Enable debugging and monitoring |

---

## 5. Current Technology Stack

### Backend

- Go
- Gin

### Storage

- PostgreSQL
- Redis
- MongoDB (benchmark experiments only)

### Observability

- Prometheus
- Grafana
- Loki
- Promtail
- Tempo
- Jaeger
- OpenTelemetry

### Infrastructure

- Docker
- Docker Compose
- Makefile

### Logging

- Zerolog

### Migrations

- `github.com/rubenv/sql-migrate`

---

## 6. Current Architecture

### 6.1 Entry Point

#### `cmd/server/main.go`

Responsible for:

- Loading configuration
- Initializing logger
- Initializing tracer
- Initializing metrics
- Creating the HTTP server
- Lifecycle management
- Graceful shutdown

### 6.2 API Layer

#### `internal/api/routes.go`

Responsible for:

- Route registration
- Middleware registration
- Middleware ordering

#### `internal/api/handlers.go`

Current endpoints:

| Endpoint | Purpose |
|---|---|
| `GET /hello` | Basic health response |
| `GET /delayed-hello` | Latency simulation |
| `GET /health` | Health state |
| `GET /ready` | Readiness |
| `GET /live` | Liveness |
| `GET /test-error` | Error testing |
| `GET /metrics` | Prometheus metrics |

### 6.3 Middleware Stack

Path:

```text
internal/middleware/
```

Implemented middleware:

| Middleware | Responsibility |
|---|---|
| `prometheus.go` | HTTP metrics |
| `tracing.go` | OpenTelemetry tracing |
| `correlation.go` | Correlation IDs |
| `logging.go` | Request/response logging |
| `error_handler.go` | Centralized error handling |

Middleware order is intentionally controlled and considered critical.

### 6.4 Lifecycle Management

#### `internal/lifecycle/lifecycle.go`

Responsible for:

- Service state management
- Readiness state
- Shutdown state
- Health integration

#### `internal/server/server.go`

Responsible for:

- HTTP server wrapper
- Startup/shutdown
- Graceful shutdown handling

### 6.5 Configuration Management

#### `internal/config/config.go`

Responsible for:

- Config loading
- Validation
- Environment parsing

Environment files:

```text
.env
env.example
```

### 6.6 Logging

#### `internal/logger/logger.go`

Responsible for:

- Zerolog configuration
- Structured logging

### 6.7 Metrics

#### `internal/metrics/metrics.go`

Responsible for:

- Prometheus metric definitions
- Metric registration

### 6.8 Tracing

#### `internal/tracer/tracer.go`

Responsible for:

- OpenTelemetry setup
- Exporter initialization
- Trace propagation

### 6.9 Shared Errors

#### `pkg/errors/errors.go`

Reusable standardized application errors.

---

## 7. Observability Stack

A dedicated observability stack already exists.

### Components

| Component | Purpose |
|---|---|
| Prometheus | Metrics collection |
| Grafana | Dashboards |
| Loki | Logs |
| Promtail | Log shipping |
| Tempo | Tracing backend |
| Jaeger | Trace visualization |

The observability stack is currently considered complete enough for the next implementation phase. Future tasks may improve dashboards, metrics taxonomy, trace attributes, alerting, and operational documentation.

---

## 8. Benchmark APIs

Several benchmark-focused APIs were intentionally implemented before business logic.

Purpose:

- Establish performance baselines
- Understand system bottlenecks
- Compare infrastructure behavior
- Support architecture decisions

### 8.1 OTP Generation Benchmark

#### `GET /otp/code`

Measures:

- Pure application baseline
- Minimal API overhead
- OTP generation cost without database/cache/provider involvement

### 8.2 PostgreSQL Read Benchmark

#### `GET /v1/otp/tenant-settings/{tenant_id}`

Measures:

- Simple `SELECT` latency
- PostgreSQL read overhead
- Tenant settings lookup cost

### 8.3 PostgreSQL Insert Benchmark

#### `POST /v1/otp/tenant-settings-insert-benchmark`

Measures:

- Insert throughput
- Write latency
- PostgreSQL audit/reporting write cost baseline

### 8.4 Redis Benchmarks

#### `POST /v1/redis/set`

#### `GET /v1/redis/get`

Measures:

- Redis write latency
- Redis read latency
- Cache operation baseline

### 8.5 MongoDB Benchmarks

#### `POST /v1/mongo/set`

#### `GET /v1/mongo/get`

Measures:

- MongoDB write/read performance
- Comparison against Redis/PostgreSQL
- Experimental persistence/cache behavior

---

## 9. OTP Flow — Functional Design

The next major implementation phase is the first real OTP flow.

This flow includes:

- Tenant lookup
- Tenant configuration caching
- OTP generation
- OTP storage
- Simulated SMS sending
- Request logging
- Provider response logging
- OTP verification

---

## 9.1 Send OTP API

### Endpoint

```http
POST /v1/otp/send
```

### Request

```json
{
  "phone": "+989121234567",
  "tenant_id": 12345,
  "token": "sfsdsf",
  "metadata": {}
}
```

### Response

```json
{
  "request_id": "uuid",
  "expired_at": "datetime"
}
```

---

## 9.2 Initial Assumptions

Currently:

- Token validation is skipped.
- Tenant exists.
- SMS provider always succeeds.
- SMS provider is simulated.
- OTP generator already exists.
- Tenant settings migrations already exist.
- The system uses existing observability middleware.

---

## 10. OTP Send Flow

### Step 1 — Receive Request

Request enters the Gin handler.

Middleware stack executes:

- Correlation ID middleware
- Tracing middleware
- Logging middleware
- Metrics middleware
- Error handling middleware

The request should receive or propagate a correlation ID and trace context.

### Step 2 — Validate Request

Initial validation should include:

- `phone` is required.
- `tenant_id` is required.
- `metadata` is optional.
- `token` is accepted but not validated yet.

Future validation may include:

- Phone format validation
- Tenant authorization
- Token/API key validation
- Rate limit checks
- Tenant quota checks

### Step 3 — Tenant Lookup

System attempts to retrieve tenant configuration.

Strategy:

```text
Redis cache
    ↓ miss
PostgreSQL
    ↓
async cache population
```

Lookup behavior:

1. Try Redis using tenant key.
2. If found, use cached tenant settings.
3. If not found, read from PostgreSQL.
4. Populate Redis for future requests.
5. Continue the send flow.

For the initial version, cache population can be synchronous or asynchronous. Synchronous is simpler; asynchronous is more scalable but introduces complexity. The preferred initial implementation may be synchronous unless current code already has safe async patterns.

### Step 4 — OTP Generation

Generate a 6-digit OTP code.

Requirements:

- Use secure randomness if possible.
- Keep OTP length configurable in the future.
- Keep expiration configurable in the future.
- Avoid predictable generation logic.

### Step 5 — Store OTP

OTP verification data should be stored in Redis because it is ephemeral and latency-sensitive.

Suggested Redis key:

```text
otp:{tenant_id}:{phone}
```

Stored data:

- OTP hash
- Expiration timestamp
- Created timestamp
- Attempt count
- Request ID
- Tenant ID
- Phone number

Important security rule:

OTP must not be stored as plaintext.

Recommended approach:

- Hash OTP before storage.
- Compare hashed values during verification.
- Avoid logging OTP values.

### Step 6 — Simulated SMS Sending

Current implementation should use a simulated SMS provider.

Behavior:

- Random delay between 20ms and 40ms
- Always returns success

Purpose:

- Simulate realistic external latency
- Benchmark asynchronous behavior
- Avoid third-party dependency noise
- Keep the system deterministic enough for local development

Recommended design:

```text
OTP service
    ↓
SMS provider interface
    ↓
Fake SMS provider implementation
```

The fake provider should be replaceable later with a real provider.

### Step 7 — Request Logging

OTP request and provider result should be stored in PostgreSQL.

Stored data should include:

- Request ID
- Tenant ID
- Phone number
- Status
- Provider name
- Provider response
- Error message if any
- Correlation ID
- Metadata
- Created timestamp
- Updated timestamp

This data is required for future reporting and tenant dashboards.

### Step 8 — Response

API returns:

```json
{
  "request_id": "uuid",
  "expired_at": "datetime"
}
```

The response should not expose the generated OTP.

---

## 11. OTP Verification Flow

### Endpoint

```http
POST /v1/otp/verify
```

### Request

```json
{
  "tenant_id": 12345,
  "phone": "+989121234567",
  "code": "123456"
}
```

### Response

Possible successful response:

```json
{
  "verified": true,
  "request_id": "uuid"
}
```

Possible failed response:

```json
{
  "verified": false,
  "request_id": "uuid",
  "reason": "invalid_code"
}
```

### Verification Process

1. Receive request.
2. Validate required fields.
3. Read OTP verification state from Redis.
4. If OTP does not exist, return expired/not found.
5. Check expiration.
6. Check attempt count.
7. Hash received OTP code.
8. Compare with stored OTP hash.
9. If valid, invalidate OTP.
10. Log verification result.
11. Return result to tenant.

---

## 12. OTP Verification Rules

Initial recommended rules:

| Rule | Value |
|---|---|
| Expiration | 2-5 minutes |
| Max attempts | 3-5 |
| Single-use | Yes |
| Reusable | No |
| Store plaintext OTP | No |
| Log OTP value | No |

Recommended behavior:

- Successful verification deletes or invalidates the OTP.
- Failed verification increments attempt count.
- Too many attempts invalidates or blocks the OTP.
- Expired OTP should not be accepted.
- Verification result should be logged for reporting.

---

## 13. Redis Design

Current recommendation:

- Use a single Redis instance for now.
- Use logical key namespaces.
- Avoid physical separation at this stage.

Example namespaces:

```text
tenant:
otp:
rate-limit:
```

Tenant settings keys:

```text
tenant:{tenant_id}:settings
```

OTP keys:

```text
otp:{tenant_id}:{phone}
```

Future rate-limit keys:

```text
rate-limit:{tenant_id}:{phone}
```

Future separation may occur later.

Possible future Redis separation:

| Redis Usage | Reason |
|---|---|
| OTP cache | Ultra-low latency and TTL-heavy |
| Tenant cache | Read-heavy configuration |
| Rate limiting | Atomic counters |
| Queue/cache hybrid | Async job coordination |

For the current phase, one Redis instance is enough.

---

## 14. PostgreSQL Design

PostgreSQL is used for strongly consistent and reportable data.

### PostgreSQL should store:

- Tenant settings
- OTP request logs
- SMS provider responses
- Verification results
- Future reporting data

### Redis should store:

- OTP verification state
- Tenant settings cache
- Future rate limiting data

### Suggested tables

Existing:

- `tenant_settings`

Future or current next-step tables:

- `otp_requests`
- `otp_verifications`

### `otp_requests` should include:

- `id`
- `request_id`
- `tenant_id`
- `phone`
- `status`
- `provider_name`
- `provider_response`
- `error_message`
- `metadata`
- `correlation_id`
- `created_at`
- `updated_at`

### `otp_verifications` should include:

- `id`
- `request_id`
- `tenant_id`
- `phone`
- `result`
- `reason`
- `attempt_count`
- `correlation_id`
- `created_at`

---

## 15. Caching Strategy

Tenant configuration caching:

```text
Redis → PostgreSQL fallback
```

Current recommendation:

- Use TTL-based invalidation.
- Accept eventual consistency for cached tenant settings.
- Keep PostgreSQL as source of truth.
- Use Redis only as a read optimization.

Reason:

Tenant settings usually change infrequently.

Cache miss behavior:

1. Query PostgreSQL.
2. If found, store in Redis.
3. Continue processing request.

Cache failure behavior:

If Redis is unavailable:

- For tenant lookup, fallback to PostgreSQL.
- Log Redis failure.
- Increment cache failure metric.
- Continue if PostgreSQL is available.

---

## 16. Async Boundaries

The system intentionally introduces async boundaries.

Current async candidates:

- SMS sending
- Request logging
- Provider result logging
- Cache population

Initial recommendation:

For the first real implementation, avoid introducing a queue too early.

Use interfaces that allow future async evolution:

```text
TenantSettingsProvider
OTPStore
SMSProvider
OTPRequestLogger
```

This keeps the current implementation simple while keeping future architecture flexible.

Future architecture may evolve toward:

- Workers
- Queues
- Kafka
- NATS
- Batching
- Retry pipelines
- Outbox pattern

---

## 17. Failure Scenarios

The following failure scenarios must eventually be handled:

| Failure | Expected Strategy |
|---|---|
| Redis unavailable | Fallback for tenant cache; fail verification if OTP store unavailable |
| PostgreSQL unavailable | Fail send if tenant data cannot be loaded or audit log is mandatory |
| SMS provider timeout | Timeout + retry strategy |
| Duplicate requests | Idempotency strategy |
| Partial writes | Consistency policy |
| High latency | Timeout + circuit breaker |
| Cache stale | TTL + eventual refresh |
| OTP expired | Return verification failure |
| Too many attempts | Reject and invalidate OTP |
| Provider failure | Store failed provider response and expose safe error |

Important design question:

The system must decide whether audit logging is mandatory before returning success.

Initial recommendation:

- Store OTP before sending SMS.
- Store request/provider result in PostgreSQL.
- If provider fails, return a controlled failure and persist failure result.
- Do not lose request visibility.

---

## 18. Non-Functional Requirements (NFRs)

Initial expectations:

| Requirement | Target |
|---|---|
| Initial tenant count | 10K |
| Future tenant count | 100K+ |
| OTP verification latency | Low |
| Scalability | Horizontal-ready |
| Observability | Mandatory |
| Graceful shutdown | Mandatory |
| Structured logs | Mandatory |
| Trace propagation | Mandatory |
| Redis usage | Required for latency-sensitive state |
| PostgreSQL usage | Required for source-of-truth and reporting |
| SMS provider abstraction | Required |

Open NFRs to define later:

| Requirement | Needs Decision |
|---|---|
| p95 send latency | TBD |
| p95 verify latency | TBD |
| OTP expiration | TBD |
| Max attempts | TBD |
| Provider timeout | TBD |
| Retry policy | TBD |
| Rate limit policy | TBD |
| Tenant quota policy | TBD |

---

## 19. Future Features

Planned future capabilities:

- JWT authentication
- API keys
- Rate limiting
- Tenant quotas
- Provider failover
- Multi-provider routing
- Multi-region support
- Distributed queues
- Analytics dashboards
- Fraud detection
- Retry pipelines
- Delivery tracking
- Tenant-level reports
- Admin APIs
- Provider cost optimization
- Circuit breaker
- Idempotency keys

---

## 20. Architecture Philosophy

The project intentionally evolves gradually.

The goal is not merely building an OTP service.

The real goal is:

- Learning system design
- Benchmarking architectural choices
- Understanding scalability bottlenecks
- Experimenting with production patterns
- Building operational maturity
- Practicing observability-first engineering
- Learning agentic coding on a real codebase

The project should avoid premature complexity, but it should also avoid short-term decisions that block future scalability.

Guiding principles:

- Keep the first implementation simple.
- Add abstractions only at real boundaries.
- Preserve observability from the beginning.
- Prefer benchmark-driven decisions.
- Keep PostgreSQL as source of truth.
- Use Redis for ephemeral and latency-sensitive data.
- Keep third-party provider integration behind interfaces.
- Document architectural decisions as the system evolves.

---

## 21. Current Development Stage

Current stage:

```text
Infrastructure & Observability Baseline  ✅
Micro-benchmark APIs                     ✅
First Real OTP Flow                      ← CURRENT
Scalability Evolution                    ⏳
Distributed System Challenges            ⏳
```

Immediate next step:

Create a dedicated OTP flow design document, then use it to guide Codex onboarding and phased implementation.

Recommended next implementation sequence:

1. Review existing project structure.
2. Identify existing tenant settings repository/code.
3. Identify existing Redis and PostgreSQL clients.
4. Design package boundaries for OTP flow.
5. Implement request/response DTOs.
6. Implement tenant settings lookup with Redis fallback to PostgreSQL.
7. Implement OTP Redis store.
8. Implement fake SMS provider.
9. Implement OTP request logging.
10. Implement `POST /v1/otp/send`.
11. Add focused tests.
12. Add observability metrics/traces/log fields.
13. Implement `POST /v1/otp/verify`.
14. Benchmark the complete flow.


## AI-Assisted Development Workflow

Development workflow, implementation slicing,
and ChatGPT/Codex collaboration rules are documented in:

docs/ai-workflow.md
