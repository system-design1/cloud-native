# Deferred Concerns — OTP Service Project

This document lists intentionally deferred items for the OTP service project. These items are not forgotten; they are deliberately postponed to keep implementation slices small, reviewable, and safe.

## 1. OTP Observability: Metrics

### Description

Add Prometheus/business metrics for OTP send and verify flows.

Potential metrics:

```text
otp_send_total{status,reason}
otp_verify_total{result,reason}
otp_provider_duration_seconds{provider,status}
otp_rate_limit_total{result}
otp_resend_block_total
otp_tenant_cache_total{result}
```

### Why Deferred

Core behavior, persistence, rate limiting, and API wiring were prioritized first.

### Current Risk

Medium.

Without metrics, operational visibility is limited.

### Complexity

Medium.

Main concerns:

- label cardinality
- tenant_id label decision
- avoiding phone labels
- choosing service vs handler instrumentation

### Recommended Priority

High.

---

## 2. OTP Tracing

### Description

Add OpenTelemetry spans and attributes around:

- SendOTP
- VerifyOTP
- tenant settings lookup
- Redis OTP state operations
- SMS provider call
- request logging
- verification logging
- rate limiter check

### Why Deferred

Business behavior was built first. Tracing is best added after stable flow boundaries exist.

### Current Risk

Medium.

Debugging production latency/failure paths is harder without traces.

### Complexity

Medium.

### Recommended Priority

High.

---

## 3. Circuit Breaker for Real SMS Provider

### Description

Add circuit breaker around real external SMS provider calls.

### Why Deferred

Current provider is fake/simulated. Circuit breaker becomes more valuable once a real SMS provider exists.

### Current Risk

Low now, High later.

### Complexity

Medium.

### Recommended Priority

Medium now, High before real provider production use.

---

## 4. Real SMS Provider Adapter

### Description

Implement real SMS provider integration, likely behind the existing `SMSProvider` interface.

### Why Deferred

The fake provider was sufficient for OTP flow development and manual verification testing.

### Current Risk

Medium.

The system cannot send real SMS yet.

### Complexity

Medium to High depending on provider.

### Recommended Priority

High if moving toward production.

---

## 5. SMS Provider Router / Registry

### Description

Route SMS requests to different providers based on tenant settings.

Example:

```text
tenant.SMSProvider = kavenegar
tenant.SMSProvider = fake
tenant.SMSProvider = provider_x
```

### Why Deferred

Only fake provider currently exists.

### Current Risk

Medium.

Tenant settings already includes provider information, but runtime does not yet use provider-specific routing.

### Complexity

Medium.

### Recommended Priority

Medium.

---

## 6. Retry Policy for SMS Provider

### Description

Add controlled retry behavior for transient SMS provider failures.

### Why Deferred

Retry behavior must be designed carefully to avoid duplicate SMS sends and provider billing issues.

### Current Risk

Medium.

### Complexity

Medium to High.

### Recommended Priority

Medium.

---

## 7. Provider Billing Protection

### Description

Add tenant/provider quota protection to avoid excessive SMS costs.

### Why Deferred

Per tenant + phone rate limiting was implemented first as the smallest useful abuse protection.

### Current Risk

Medium to High.

### Complexity

Medium.

### Recommended Priority

High before production traffic.

---

## 8. Tenant-Wide OTP Send Rate Limit

### Description

Add aggregate tenant-level send limit.

Example:

```text
tenant 202 can send at most 1000 OTPs per hour
```

### Why Deferred

The first rate limiter only covers tenant + phone.

### Current Risk

Medium.

Per-phone limiting does not prevent one tenant from sending to many different phone numbers.

### Complexity

Medium.

### Recommended Priority

High.

---

## 9. Per-IP Rate Limit

### Description

Add per-IP OTP send rate limiting.

### Why Deferred

Requires trusted client IP extraction and proxy/header policy.

### Current Risk

Medium.

### Complexity

Medium.

### Recommended Priority

Medium.

---

## 10. Global Per-Phone Rate Limit

### Description

Limit OTP sends to a phone number across all tenants.

### Why Deferred

Requires product decision: whether cross-tenant phone-level limiting is acceptable.

### Current Risk

Medium.

### Complexity

Medium.

### Recommended Priority

Medium.

---

## 11. Retry-After Header

### Description

Return `Retry-After` or structured retry metadata when rate limit or active OTP blocks a request.

### Why Deferred

Initial error mapping only returns 429 with message.

### Current Risk

Low.

### Complexity

Low to Medium.

### Recommended Priority

Medium.

---

## 12. Sliding Window or Token Bucket Rate Limiting

### Description

Replace or supplement fixed-window rate limiting with sliding window or token bucket.

### Why Deferred

Fixed window is simpler and good enough for the first phase.

### Current Risk

Low to Medium.

Fixed window allows bursts around window boundaries.

### Complexity

Medium to High.

### Recommended Priority

Low to Medium.

---

## 13. Atomic Active OTP Reservation

### Description

Make active OTP check + reservation/save atomic to avoid concurrent sends.

Current issue:

```text
Request A: Get not found
Request B: Get not found
Request A: Save/send
Request B: Save/send
```

### Why Deferred

Current resend protection handles normal repeated sends. The race exists only in concurrent edge cases.

### Current Risk

Medium.

### Complexity

Medium.

### Recommended Priority

Medium.

---

## 14. Phone Normalization

### Description

Normalize phone numbers before Redis keys, database logs, rate limiting, and SMS sends.

### Why Deferred

Current flow assumes input phone is already usable.

### Current Risk

Medium.

Different formats may bypass rate limits or create duplicate OTP states.

### Complexity

Medium.

### Recommended Priority

High before production.

---

## 15. Phone Hashing in Redis Keys

### Description

Avoid storing raw phone numbers directly in Redis keys.

### Why Deferred

The current implementation follows existing simple key style for development speed.

### Current Risk

Medium.

Raw phone numbers in Redis keys may be sensitive.

### Complexity

Low to Medium.

### Recommended Priority

Medium.

---

## 16. Auth / Token Validation for OTP APIs

### Description

Validate tenant/client credentials before allowing OTP operations.

### Why Deferred

Core OTP behavior was implemented first.

### Current Risk

High for production.

OTP endpoints should not be public without proper authentication/authorization.

### Complexity

Medium.

### Recommended Priority

High.

---

## 17. Correlation ID Propagation

### Description

Propagate request correlation ID into:

- OTP request logs
- OTP verification logs
- traces
- service context

### Why Deferred

Core logging was implemented first with empty correlation IDs.

### Current Risk

Medium.

### Complexity

Low to Medium.

### Recommended Priority

High for observability.

---

## 18. Structured Logging

### Description

Add structured logs for key OTP events without exposing OTP codes.

### Why Deferred

Database logging and tests were prioritized first.

### Current Risk

Medium.

### Complexity

Low to Medium.

### Recommended Priority

Medium.

---

## 19. OpenAPI / Swagger Documentation

### Description

Document OTP send/verify API contracts.

### Why Deferred

API behavior was still evolving.

### Current Risk

Low to Medium.

### Complexity

Low.

### Recommended Priority

Medium.

---

## 20. Admin / Reporting APIs

### Description

Expose operational reports for:

- OTP requests
- verification outcomes
- provider failures
- rate limit hits

### Why Deferred

Tables exist, but reporting API is not part of core OTP flow.

### Current Risk

Low.

### Complexity

Medium.

### Recommended Priority

Low to Medium.

---

## 21. Migration Automation / Makefile Support

### Description

Add proper migration commands to Makefile or project tooling.

### Why Deferred

Migrations were manually applied during development.

### Current Risk

Medium.

Manual migration execution is error-prone.

### Complexity

Low to Medium.

### Recommended Priority

High.

---

## 22. End-to-End Docker/Integration Tests

### Description

Add reproducible E2E tests for:

- HTTP send
- Redis OTP state
- fake SMS debug capture
- verify
- DB logs
- rate limiting

### Why Deferred

Unit/integration-style package tests were prioritized first.

### Current Risk

Medium.

### Complexity

Medium.

### Recommended Priority

High.

---

## 23. Production Hardening for Debug OTP Capture

### Description

Ensure dev-only OTP capture cannot be accidentally enabled in production.

### Why Deferred

Basic guardrails already exist.

### Current Risk

Medium.

### Complexity

Low.

### Recommended Priority

Medium.

---

## 24. Real Provider Failure Simulation

### Description

Allow controlled fake SMS provider failures to test provider failure paths.

### Why Deferred

Fake provider currently always succeeds except context cancellation.

### Current Risk

Low to Medium.

### Complexity

Low.

### Recommended Priority

Low to Medium.

---

## 25. Updating `otp_requests` After Successful Verification

### Description

Optionally update request lifecycle status once OTP is successfully verified.

### Why Deferred

Verification logging has its own table and current request lifecycle stops at provider result.

### Current Risk

Low.

### Complexity

Medium.

### Recommended Priority

Low to Medium.

---

## 26. Strict Foreign Keys Between Logs

### Description

Link `otp_verifications.request_id` to `otp_requests.request_id`.

### Why Deferred

Some verification outcomes, such as `not_found`, may not have a request ID.

### Current Risk

Low.

### Complexity

Medium.

### Recommended Priority

Low.

---

## 27. HMAC/Pepper for OTP Hashing

### Description

Use HMAC or peppered hashing instead of plain SHA-256.

### Why Deferred

Plain SHA-256 was sufficient for initial flow development.

### Current Risk

Medium.

OTP codes are low entropy, so plain hashes are weaker if Redis is leaked.

### Complexity

Medium.

### Recommended Priority

High before production.

---

## 28. Configuration Review and Secrets Management

### Description

Review all OTP-related env config and introduce proper secret handling when real provider is added.

### Why Deferred

No real SMS provider secret is wired yet.

### Current Risk

Medium.

### Complexity

Medium.

### Recommended Priority

High before production.

---

## 29. HA / Availability Scenario Testing

### Description

Design and test high availability scenarios for the OTP service.

Planned scenarios:

1. run traffic against service
2. perform HA/failover test with Keepalived
3. repeat or compare with Nginx-based setup

### Why Deferred

Core OTP application behavior was implemented first.

### Current Risk

Medium.

### Complexity

Medium to High.

### Recommended Priority

High for infrastructure validation.

---

## 30. Nginx-Based HA / Load Balancing Scenario

### Description

Test OTP service availability behind Nginx.

Potential goals:

- load balancing
- backend failure behavior
- health checks
- retry behavior
- connection handling
- failover validation

### Why Deferred

Keepalived scenario is planned first.

### Current Risk

Medium.

### Complexity

Medium.

### Recommended Priority

High after Keepalived scenario.
