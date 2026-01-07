# Stage 12 Prompt — Meaningful Traces for Demo (Manual Child Spans) + Docs

## Context
This repository already has OpenTelemetry tracing integrated (Gin + `otelgin`) and Jaeger/Tempo support.
Currently, endpoints such as `/hello` or `/delayed-hello` often appear in Jaeger as a single server span (e.g., `GET /delayed-hello`), which does not demonstrate where time is spent inside handlers.

You must add **manual child spans** inside selected handlers to produce a meaningful waterfall in Jaeger for demos and practical troubleshooting.

> **Important workflow constraints (must follow):**
> 1) Read `README.md` first.
> 2) Review and align with previous tasks in `tasks/` (especially Stage 11 route policy) so naming, structure, and conventions remain consistent.
> 3) Update documentation in `README.md` and relevant files under `docs/` after implementation.
> 4) Per project rules, all configuration must come from `.env` via the config layer (if you introduce any new config).
> 5) Keep changes minimal and production-friendly (no demo hacks that cannot be justified operationally).

---

## Goals

### Primary Goal: Make Jaeger Waterfall Meaningful
Implement manual spans so that Jaeger shows a breakdown of time spent inside handlers, not only the top-level HTTP server span.

### Demo Target Endpoints
- `/delayed-hello` must show a multi-span waterfall
- `/test-error` must clearly show an error span/event and `error=true`

---

## Requirements

### 1) Manual Child Spans for `/delayed-hello`
Inside the `/delayed-hello` handler, create at least these child spans:

1) `handler.delayed_hello`  
2) `db.query.fake`  
3) `external.call.mock` (or equivalent)  
4) Optional but recommended: `sleep.random_delay`

Notes:
- The existing otelgin server span (e.g., `GET /delayed-hello`) will remain the parent; your spans must be nested under it.
- Use `c.Request.Context()` as the base context.
- Ensure spans are ended properly (`defer span.End()`).

### 2) Attributes and Events (Make spans useful)
Add meaningful attributes to spans, at minimum:

For `handler.delayed_hello`:
- `http.route=/delayed-hello` (if available)
- `app.component=handler`

For `db.query.fake`:
- `db.system=postgres` (or a sensible default)
- `db.operation=SELECT` (or similar)
- `app.component=db`

For `external.call.mock`:
- `peer.service=downstream-mock`
- `app.component=external`

For `sleep.random_delay`:
- `sleep.ms=<duration_ms>`
- `app.component=sleep`

Use OpenTelemetry attribute conventions where reasonable. Keep it consistent with existing instrumentation in the repo.

### 3) Error Visibility for `/test-error`
Enhance `/test-error` so that traces show errors clearly.

Required:
- Create a span inside the handler (e.g., `handler.test_error`) and mark it as error using:
  - span status set to error, AND/OR
  - record an exception event (preferred if already used in the repo), AND/OR
  - set attribute `error=true`

The goal is that in Jaeger you can filter or visually identify error traces.

### 4) Tracer Naming Consistency
Do NOT hardcode random tracer names.
Follow repository conventions:
- If there is already a tracer created in `internal/tracer` or a shared instrumentation helper, reuse it.
- Otherwise, create a package-level tracer with a stable name aligned with the service name (preferably from config).
- Avoid scattering tracer initialization across multiple files; keep it centralized if possible.

### 5) No New Config Unless Necessary
This stage should not require new `.env` variables.
If you feel you need configuration (e.g., to enable/disable demo spans), do NOT add it unless strictly necessary.
Prefer always-on spans for these endpoints since they are also useful operationally.

### 6) Documentation Updates (Mandatory)
Update docs after implementation:

- `README.md`
  - Add a short section: “Jaeger Demo Traces”
  - Explain what to look for in `/delayed-hello` trace (child spans and timings)
  - Explain `/test-error` trace and how error appears

- `docs/OBSERVABILITY.md`
  - Add a subsection: “Manual spans for handler breakdown”
  - Mention where spans are created and what attributes are expected
  - Provide a simple demo flow (hit `/delayed-hello`, open Jaeger, inspect waterfall)

If other files under `docs/` reference tracing behavior, keep them consistent.

---

## Acceptance Criteria (Definition of Done)

1) Calling `/delayed-hello` results in Jaeger traces with:
   - more than 1 span (at least 4 total spans including server span)
   - visible nesting:
     - `GET /delayed-hello`
       - `handler.delayed_hello`
         - `db.query.fake`
         - `external.call.mock`
         - (optional) `sleep.random_delay`

2) Calling `/test-error` results in Jaeger traces that:
   - clearly show an error
   - include an internal span (not only the server span)
   - support easy discovery (tag/attribute indicates error)

3) Docs updated (`README.md`, `docs/OBSERVABILITY.md`, and any relevant files).

4) Code follows existing repo patterns and passes tests/build.

---

## Demo Script (for team presentation)
1) Hit `/delayed-hello` several times.
2) In Jaeger, search by operation `GET /delayed-hello`.
3) Open a trace and show the waterfall:
   - identify `db.query.fake`, `external.call.mock`, and `sleep.random_delay`
   - compare durations and explain where time is spent
4) Hit `/test-error`.
5) In Jaeger, search for error traces (e.g., `error=true` or status code 500).
6) Open the trace and show where the error is recorded.

---

## Notes for PR Review
- Keep spans minimal and meaningful.
- Use consistent naming and attributes.
- Avoid demo-only hacks that reduce operational value.
