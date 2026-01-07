# Task 15: Add a Make Target to Reset Tracing Data and Recover Jaeger During Load Tests

## Background / Problem Statement
When running high-throughput k6 load tests (e.g., 10k–13k RPS) against high-traffic endpoints such as `/hello`, the Jaeger UI can become unresponsive (endless loading spinner) and traces may not load.

This typically happens because trace volume is too high:
- Excessive tracing for high-volume routes overwhelms the tracing pipeline (OTel exporter/collector + trace storage).
- Jaeger UI queries can time out or become very slow under heavy ingestion and large datasets.

Data retention is **not important** for this scenario. The primary goal is to quickly return the tracing stack (Jaeger/Tempo/etc.) to a clean, working state.

---

## Goal
Add a single, easy command (preferably a Makefile target) that:
1. Stops the observability stack
2. Deletes all trace storage data (volumes)
3. Starts the observability stack again from scratch

This should allow Jaeger to load normally after heavy load tests.

---

## Required Changes

### 1) Add Makefile Target
Update `Makefile` to add a new target, following existing conventions and naming patterns in the repository.

Suggested target name (choose one that matches repo conventions):
- `observability-reset`
- `tracing-reset`
- `jaeger-reset`

The target must:
- Use the same compose file(s) used by the project to run observability (e.g., `docker-compose.observability.yml`).
- Run a full reset that removes volumes.

Example implementation (adapt to repo conventions):
```bash
docker compose -f docker-compose.observability.yml down -v
docker compose -f docker-compose.observability.yml up -d
```

If the repo uses a compose project name (e.g., `-p sdgo`), ensure the reset uses the same project name consistently.

---

### 2) Add Documentation
Create or update a documentation file (preferred: new file):

`docs/OBSERVABILITY_RESET.md`

The doc must include:
- The symptom: Jaeger UI stuck loading after high-RPS tests
- Why it happens (trace overload)
- The reset command (Makefile target)
- What the reset does (deletes all observability volumes, including trace storage)
- How to verify recovery:
  - Check containers are running
  - Open Jaeger UI
  - Run a small request and verify traces appear (for sampled routes)

Include a warning that this operation **deletes all tracing data**.

---

### 3) Ensure No Impact on Prometheus Metrics
Add a short section to the documentation clarifying:
- OTEL trace sampling changes do **not** affect Prometheus metrics
- Metrics collection is independent from trace sampling
- Dropping/low-sampling `/hello` traces is safe for Prometheus dashboards and counters

---

## Validation Steps (Required)
1. Start observability stack (using existing make target or compose command).
2. Run a high-load k6 test against `/hello` until Jaeger UI becomes slow/unresponsive.
3. Run the new reset target, e.g.:
   ```bash
   make observability-reset
   ```
4. Confirm:
   - All observability containers start successfully
   - Jaeger UI loads normally
   - Traces appear again for routes that are sampled (e.g., `/test-error` if sampled at 100%)

---

## Acceptance Criteria
- Makefile contains a reset target that wipes observability volumes and restarts the stack
- `docs/OBSERVABILITY_RESET.md` exists (or existing docs updated) with clear instructions
- Documentation explains why Jaeger breaks under load and how reset fixes it
- Documentation states trace sampling does not break Prometheus metrics
- Steps are reproducible and aligned with existing repo patterns
