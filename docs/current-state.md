# Current State — OTP Service Project

## Purpose

This project currently contains a working OTP flow for a Go backend service. The implementation has been built incrementally with small, reviewable slices using ChatGPT for architecture/review and Codex for focused implementation.

The OTP subsystem currently supports:

- OTP send
- OTP verify
- Redis-backed OTP state
- tenant settings lookup with Redis cache + PostgreSQL fallback
- fake SMS provider
- dev-only fake SMS OTP code capture
- PostgreSQL request logging
- PostgreSQL verification logging
- resend protection while an OTP is active
- per tenant + phone OTP send rate limiting
- HTTP handlers and routes
- runtime wiring in `cmd/server/main.go`
- env-driven OTP configuration
- focused unit and integration-style tests

## Current Architecture Snapshot

Current high-level flow:

```text
HTTP API
  -> internal/api handlers
  -> internal/otp.Service
  -> ports/interfaces
  -> Redis/PostgreSQL/SMS adapters
```

Main components:

```text
internal/api
  Thin HTTP handlers and route registration.

internal/otp
  Domain/application service for SendOTP and VerifyOTP.
  Owns core OTP orchestration, validation, hashing, retry/attempt rules, and domain errors.

internal/repository
  PostgreSQL and Redis adapters:
  - tenant settings repository
  - cached tenant settings provider
  - Redis OTP store
  - Redis send rate limiter
  - OTP request log repository
  - OTP verification log repository

internal/sms
  Fake SMS provider used for local/dev and simulated provider behavior.

internal/config
  Env-driven runtime configuration for OTP, fake SMS, and send rate limiting.

cmd/server/main.go
  Runtime wiring for database, Redis, repositories, OTP service, SMS provider, and routes.
```

## Implemented Features

### OTP Domain Foundation

Implemented:

- OTP domain request/response models
- tenant settings model
- OTP state model
- SMS request/result models
- request/provider/verification log models
- domain errors
- interfaces/ports
- OTP config defaults
- OTP hashing helpers
- configurable numeric OTP generation

### OTP Generation and Hashing

Implemented:

- dynamic numeric OTP generation
- backward-compatible 6-digit generator
- SHA-256 based hash helper
- constant-time code verification helper
- no plaintext OTP persistence in the main OTP state

Important behavior:

```text
Redis OTP state stores code_hash only.
Plaintext OTP is not stored in the main OTP key.
```

### Redis OTP Store

Implemented:

- Redis-backed OTP state store
- key format: `otp:{tenant_id}:{phone}`
- Redis Hash storage
- Save/Get/Delete
- atomic IncrementAttempts using Redis Lua
- TTL-based expiration
- malformed state detection
- integration-style Redis tests

Stored fields:

```text
request_id
tenant_id
phone
code_hash
attempt_count
max_attempts
created_at
expires_at
```

### Tenant Settings Cache Provider

Implemented:

- Redis cache-aside provider for tenant settings
- PostgreSQL fallback
- Redis key format: `tenant:{tenant_id}:settings`
- stores only OTP-domain tenant settings subset
- avoids caching sensitive/unneeded DB fields
- falls back on malformed cache
- source errors returned when PostgreSQL lookup fails

### Fake SMS Provider

Implemented:

- fake SMS provider implementing OTP SMS provider interface
- configurable latency
- default delay: `20ms to 30ms`
- context cancellation/timeout support
- safe SMS result
- no OTP code in RawResponse
- dev-only Redis debug code capture

### Dev-Only Fake SMS OTP Capture

Implemented for local/manual testing only.

Behavior:

- disabled by default
- enabled through config/env
- only active outside release mode
- stores plaintext OTP in a separate Redis debug key
- does not expose code in API response
- does not write code into `otp_requests`
- does not write code into normal OTP state
- does not log the code

Debug key format:

```text
debug:otp-code:{tenant_id}:{phone}
```

### SendOTP Service

Implemented behavior:

1. validate request
2. load tenant settings
3. validate tenant
4. check active OTP resend protection
5. check optional send rate limiter
6. generate request ID
7. generate OTP code
8. hash OTP code
9. create OTP request log
10. save OTP state in Redis
11. send SMS through provider with timeout
12. update provider result log
13. return request ID and expiration

Important details:

- tenant validation happens before Redis OTP state check
- active OTP protection happens before rate limiting
- blocked active resend does not create request log
- rate-limited send does not create request log
- SMS provider failure is mapped to domain provider failure
- request logging is mandatory for send lifecycle

### VerifyOTP Service

Implemented behavior:

1. validate request
2. load OTP state from Redis
3. handle not found
4. handle expiration
5. handle max attempts already reached
6. verify code hash
7. increment attempts only for wrong code
8. handle invalid code
9. handle max attempts after increment
10. delete OTP state strictly on success
11. return verified response

Important details:

- correct code does not increment failed attempts
- successful verification is one-time-use
- delete failure after correct code returns error
- expired/max-attempt cleanup delete is best-effort
- verification logging is best-effort

### Request Logging

Implemented PostgreSQL request logging using table `otp_requests`.

Behavior:

- create request log before Redis OTP save and SMS send
- update provider result after SMS success/failure
- request logging is mandatory in SendOTP
- provider response is safe and does not include OTP code

### Verification Logging

Implemented PostgreSQL verification logging using table `otp_verifications`.

Logged outcomes:

- success / verified
- failed / not_found
- failed / expired
- failed / invalid_code
- failed / max_attempts_exceeded

Important behavior:

- logging is best-effort
- logging failure does not change VerifyOTP response
- invalid request validation failures are not logged
- infrastructure failures are not logged
- success is logged only after Redis delete succeeds

### HTTP API

Implemented endpoints:

```http
POST /v1/otp/send
POST /v1/otp/verify
```

HTTP layer responsibilities:

- bind JSON
- validate required fields
- call service
- map domain errors to HTTP response
- keep business logic out of handlers

Error mappings include:

```text
invalid request -> 400
tenant disabled -> 403
tenant not found -> 404
OTP already active -> 429
OTP send rate limit exceeded -> 429
SMS provider failed -> 502
generic/internal errors -> 500
```

Verify business failures return `200 OK` with:

```json
{
  "verified": false,
  "reason": "..."
}
```

### Resend Protection

Implemented behavior:

```text
If an active, unexpired OTP already exists for tenant_id + phone,
a new SendOTP request is rejected.
```

This happens before the send rate limiter.

Response:

```http
429 Too Many Requests
```

Message:

```text
OTP already active
```

### Send Rate Limiting

Implemented per tenant + phone send rate limiting.

Scope:

```text
tenant_id + phone
```

Redis key format:

```text
otp:rate:send:{tenant_id}:{phone}
```

Current strategy:

```text
Fixed-window Redis rate limiter
```

Implementation details:

- Redis-backed adapter implements `otp.SendRateLimiter`
- uses Redis Lua script for atomic INCR + TTL handling
- repairs missing TTL if key exists without expiration
- maps limit exceeded to `ErrOTPRateLimited`
- rate limiter is optional and env-controlled
- disabled by default

Rate limiting runs after active OTP protection.

## Current Configuration / Env Support

Configured values include:

```text
OTP_CODE_LENGTH
OTP_TTL
OTP_MAX_ATTEMPTS
OTP_TENANT_CACHE_TTL
OTP_PROVIDER_TIMEOUT

OTP_FAKE_SMS_MIN_DELAY
OTP_FAKE_SMS_MAX_DELAY
OTP_FAKE_SMS_DEBUG_CODE_REDIS
OTP_FAKE_SMS_DEBUG_CODE_TTL

OTP_SEND_RATE_LIMIT_ENABLED
OTP_SEND_RATE_LIMIT_MAX
OTP_SEND_RATE_LIMIT_WINDOW
```

Current important defaults:

```text
OTP_CODE_LENGTH=6
OTP_TTL=2m
OTP_MAX_ATTEMPTS=3
OTP_TENANT_CACHE_TTL=5m
OTP_PROVIDER_TIMEOUT=2s

OTP_FAKE_SMS_MIN_DELAY=20ms
OTP_FAKE_SMS_MAX_DELAY=30ms
OTP_FAKE_SMS_DEBUG_CODE_REDIS=false
OTP_FAKE_SMS_DEBUG_CODE_TTL=60s

OTP_SEND_RATE_LIMIT_ENABLED=false
OTP_SEND_RATE_LIMIT_MAX=5
OTP_SEND_RATE_LIMIT_WINDOW=10m
```

## Current Redis Usage

Redis is used for:

- OTP state
- OTP attempt counter
- tenant settings cache
- send rate limiter
- dev-only fake SMS OTP debug capture

Main key patterns:

```text
otp:{tenant_id}:{phone}
tenant:{tenant_id}:settings
otp:rate:send:{tenant_id}:{phone}
debug:otp-code:{tenant_id}:{phone}
```

## Current PostgreSQL Usage

PostgreSQL is used for:

- tenant settings
- OTP request logs
- OTP verification logs

Tables involved:

```text
tenant_settings
otp_requests
otp_verifications
```

## Current Testing Status

Implemented test coverage includes:

- OTP generation tests
- hash/verify tests
- service SendOTP tests
- service VerifyOTP tests
- handler tests
- Redis OTP store tests
- Redis send rate limiter tests
- tenant cache provider tests
- request log repository tests
- verification log repository tests
- fake SMS provider tests
- config/env tests

Common verification commands:

```bash
go test -count=1 ./internal/otp -v
go test -count=1 ./internal/api -v
go test -count=1 ./internal/repository -v
go test -count=1 ./internal/config -v
go test -count=1 ./internal/sms -v
go test -count=1 ./...
```

## Current Manual Runtime Validation

Manual validation has been performed for:

- `/v1/otp/send`
- `/v1/otp/verify`
- active OTP resend protection
- dev-only debug code capture
- verification logging
- request logging
- send rate limiting

## Important Implementation Decisions

- Keep HTTP handlers thin.
- Keep OTP orchestration inside `internal/otp.Service`.
- Use interfaces for external dependencies.
- Keep Redis/PostgreSQL adapters in `internal/repository`.
- Keep fake SMS simulation in `internal/sms`.
- Do not expose OTP code in API responses.
- Do not store plaintext OTP in normal Redis OTP state.
- Use best-effort verification logging.
- Keep request logging mandatory for send lifecycle.
- Keep rate limiting optional and disabled by default.
- Keep changes incremental and small.
- Commit after each stable phase.

## Current Known Limitations

- No real SMS provider yet.
- No provider router/registry yet.
- No metrics/tracing for OTP business flows yet.
- No circuit breaker yet.
- No retry policy yet.
- Rate limiting is only per tenant + phone.
- No per-IP limiting yet.
- No tenant-wide quota yet.
- No global per-phone quota yet.
- No Retry-After header.
- No phone normalization/hashing.
- No OpenAPI documentation.
- No auth/token validation for OTP endpoints yet.
- No atomic reservation for active OTP send flow beyond current checks.
