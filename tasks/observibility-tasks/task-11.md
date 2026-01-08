# Stage 11 Prompt — OTEL Route-Based Tracing Policy (Always / Ratio / Drop) + Config + Docs

## Context
This repository already has OpenTelemetry tracing integrated (Gin + `otelgin`) and Jaeger/Tempo support.
Right now, some endpoints are noisy (`/metrics`, `/health`, `/ready`, `/live`) while others are useful for demos and debugging (`/delayed-hello`, `/test-error`).

You must implement a **route-based tracing policy** that supports three behaviors:
1) **ALWAYS**: always create traces
2) **RATIO**: probabilistic sampling by route (e.g., 1%, 5%, 0.1%)
3) **DROP**: never create traces

> **Important workflow constraints (must follow):**
> 1) Read `README.md` first.
> 2) Review and align with previous tasks in `tasks/` (especially the prior Stage 11 draft) so naming, structure, and conventions remain consistent.
> 3) Update documentation in `README.md` and all relevant files under `docs/` after implementation.
> 4) Per project rules, **all configuration must come from `.env`**, be loaded by the config layer, and `env.example` must be kept complete and accurate.

---

## Primary Goal (Demo-friendly defaults)
Implement a policy that, by default, behaves like this:

- **ALWAYS**
  - `/delayed-hello`
  - `/test-error`

- **RATIO**
  - `/health` at **1%** (0.01)
  - `/live` at **1%** (0.01)
  - `/ready` at **1%** (0.01)

- **DROP** (preferred) or very-low ratio
  - `/metrics` should be **DROP by default** (preferred)
    - If you decide not to DROP for any reason, use **RATIO 0.1%** (0.001) by default.

The policy must be **fully configurable via `.env`**, including a master flag to enable/disable policy behavior.

---

## Requirements

### 1) Route Policy Must Be Configurable and Toggleable
Add an env-driven feature flag:

- When policy is **enabled**:
  - The route policy is applied: ALWAYS / RATIO / DROP based on configured routes.
- When policy is **disabled**:
  - **No route policy is applied**, and the system behaves like the current baseline (i.e., tracing for all routes as per existing global sampling config — typically AlwaysSample in this repo).

Required env var:
- `OTEL_ROUTE_POLICY_ENABLED=true|false`

### 2) Configuration via `.env` (Mandatory)
Implement route policy configuration via env variables.

Recommended env vars (use these unless the repo already has a strict naming scheme you must follow):
- `OTEL_ROUTE_POLICY_ENABLED=true`
- `OTEL_ROUTE_ALWAYS=/delayed-hello,/test-error`
- `OTEL_ROUTE_DROP=/metrics`
- `OTEL_ROUTE_RATIO=/health=0.01,/live=0.01,/ready=0.01`
- `OTEL_ROUTE_DEFAULT=always|ratio|drop` (default should be `always` for this repo unless there is an existing global sampler)
- `OTEL_ROUTE_DEFAULT_RATIO=1.0` (only used when `OTEL_ROUTE_DEFAULT=ratio`)

Parsing rules:
- Comma-separated lists for ALWAYS and DROP (trim spaces, ignore empty items).
- For RATIO, parse comma-separated `path=ratio` items.
- Validate ratios: `0.0 < ratio <= 1.0`. If invalid, log a warning and apply a safe fallback (document your fallback).
- Support both GET and HEAD probes; matching must be based on **path**, not only method.

Precedence rules (must implement):
1) DROP (highest priority)
2) ALWAYS
3) RATIO
4) DEFAULT policy

### 3) Sampler Implementation (OpenTelemetry Go SDK)
Implement a custom sampler that applies the route policy.

Constraints / guidance:
- Sampling decision happens at span start. With `otelgin`, span name is typically `METHOD /path`. Attributes such as `http.route` may not be available at sampling time.
- Prefer extracting the route path from:
  - span name pattern: `GET /health` → `/health`
- Ensure parent/child consistency:
  - Use `sdktrace.ParentBased(...)` with your route-policy sampler as the root sampler.
  - If a parent is sampled, keep the child sampled to preserve trace integrity.

Suggested approach:
- Build an internal matcher:
  - `alwaysRoutes` set
  - `dropRoutes` set
  - `ratioRoutes` map[path]ratio
- Implement `ShouldSample`:
  - Resolve route path
  - Apply precedence rules:
    - DROP → `sdktrace.Drop()` decision
    - ALWAYS → `sdktrace.RecordAndSample()` decision
    - RATIO → delegate to `sdktrace.TraceIDRatioBased(ratio)`
    - DEFAULT → apply configured default
- Implement `Description()` for debug clarity.

### 4) Update Config Layer + Tests
Update:
- `internal/config/config.go` (or the repo’s config domain files)
  - Add a RoutePolicy config section (fields for enabled, always/drop lists, ratio map, defaults).
  - Load from env with validation.
- `internal/config/config_test.go`
  - Add unit tests for parsing and validation:
    - enabled on/off behavior
    - precedence (DROP overrides ALWAYS, etc.)
    - ratio parsing and invalid values
    - trimming and empty entries

### 5) Update `.env` and `env.example` (Mandatory)
- Add all new variables to `.env` with demo-friendly defaults listed above.
- Add the same variables to `env.example` with:
  - clear descriptions
  - recommended ranges (e.g., 0.001–0.05 for noisy endpoints)
  - explanation of enable/disable behavior

### 6) Update Documentation (Mandatory)
After implementation, update:
- `README.md`
  - Add a section: “Tracing Route Policy (Always/Ratio/Drop)”
  - Provide an example `.env` snippet
  - Explain how to disable the policy for debugging (`OTEL_ROUTE_POLICY_ENABLED=false`)
- `docs/OBSERVABILITY.md`
  - Explain why noisy routes exist and how policy reduces Jaeger noise
  - Explain the three policy types and precedence rules
- Any other docs under `docs/` that reference tracing config must be updated to keep consistency.

---

## Acceptance Criteria (Definition of Done)

1) With `OTEL_ROUTE_POLICY_ENABLED=true` and default values:
   - `/delayed-hello` and `/test-error` always show in Jaeger.
   - `/health`, `/live`, `/ready` appear occasionally (about 1% of calls).
   - `/metrics` does not appear (DROP), OR appears very rarely if configured as 0.1% ratio.
2) With `OTEL_ROUTE_POLICY_ENABLED=false`:
   - Route policy has no effect; behavior returns to baseline (noise returns).
3) `.env` updated with defaults.
4) `env.example` updated and documented.
5) `README.md` and relevant `docs/` updated.
6) Config parsing/validation tests added and passing.

---

## Notes for PR Review
- Keep changes minimal and consistent with existing repo conventions.
- Do not introduce vendor-specific dependencies; use OpenTelemetry Go SDK only.
- Ensure the tracer init remains compatible with existing exporters (Jaeger/Tempo).
