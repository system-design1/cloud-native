# Task 14: Prevent Jaeger UI Hang During High-Load k6 Tests by Adjusting OTEL Sampling

## Background / Problem Statement
During high-throughput k6 load tests against the `/hello` endpoint (e.g., ~10k–13k RPS on a local machine), the Jaeger UI may become unresponsive (endless loading spinner) and traces do not load.

This is expected when tracing is sampled too aggressively for high-traffic routes:
- If `/hello` is sampled at 100% (or too high), the service generates **hundreds of thousands to millions** of traces quickly.
- The tracing pipeline (OTel exporter/collector + Tempo/Jaeger storage) becomes overwhelmed:
  - ingestion backpressure / dropped spans
  - query timeouts in Jaeger UI
  - increased disk usage for trace storage

**Result:** The load test can still look “healthy” (0% HTTP errors) while observability becomes unusable due to trace overload.

---

## Goal
Keep tracing usable during load tests by applying a route-based sampling policy:
- **Drop** or **sample at 0.1% (0.001)** for high-volume endpoints like `/hello` (and optionally `/health`, `/ready`, `/live`).
- Keep **100% sampling** for error-focused endpoints such as `/test-error`.

---

## Required Changes

### 1) Add Documentation
Create a new documentation file (or if you have file for it, update it):


The document must include:
- Why Jaeger/Tempo overload happens during high RPS tests
- Symptoms (Jaeger UI spinner / missing traces)
- Recommended sampling strategy for load tests:
  - `/hello`: DROP or 0.1%
  - `/health`, `/ready`, `/live`: DROP or 0.1%
  - `/metrics`: DROP
  - `/test-error`: ALWAYS (100%)
- How to apply the policy using the repository’s **existing** sampling configuration (Task-11 style)
- How to validate:
  - Run k6 load test
  - Jaeger UI remains responsive
  - Traces exist for `/test-error`
  - Traces for `/hello` are absent (DROP) or very rare (0.1%)

Add a short “Load Testing Checklist” section:
- Enable load-test sampling policy
- Optionally reduce log verbosity
- Monitor CPU and disk usage

---

### 2) Provide a “Load Test Sampling” Configuration Example
Based on the repository’s existing OTEL sampling mechanism, add an example policy section in the new doc like:

- `/hello` -> DROP or RATIO=0.001 (0.1%)
- `/health|/ready|/live` -> RATIO=0.001
- `/metrics` -> DROP
- `/test-error` -> ALWAYS

**Important:** Do not invent configuration keys. Use the keys/format already implemented in this repository.

**Note**: I have the sampling in this project. You must set it in .env.example and also .env
---

---

## Acceptance Criteria
- `docs/LOAD_TESTING_TRACING_SAMPLING.md` or updated file exists and is written in English
- The doc clearly explains the problem and the solution
- The doc references the project’s existing OTEL sampling implementation
- The recommended policy reduces trace volume for `/hello` and keeps full tracing for `/test-error`
- Validation steps are included

---

## Notes
This task is meant to keep observability tools usable during stress testing. It is a **load-testing-focused** sampling policy and may differ from production sampling policies.
