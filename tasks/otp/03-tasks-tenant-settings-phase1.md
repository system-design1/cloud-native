# Tenant Settings – Phase 1 (Lookup API + DB Wiring)

This task list is designed to be **Cursor-friendly**: small, atomic, and with clear “Done when” criteria.

Scope:
- A simple table `tenant_settings` already exists in Postgres.
- Seed ~20,000 fake rows (for local benchmarking).
- Implement an API endpoint to fetch tenant settings by `id` and return the full row as JSON.
- No OTP flow changes yet (this is a prerequisite building block).

---


## Task 01 — Add DB config/env for Postgres connection

**Goal**
Ensure the service can read Postgres connection settings from env/config.

**Files / Paths**
- Project config package(s) (e.g. `internal/config/...`)
- `.env.example` (if present)

**Steps**
1. Add config fields for: host, port, user, password, dbname, sslmode.
2. Ensure defaults are sensible for local dev.
3. Make sure config is loaded and accessible from app bootstrap.

**Done when**
- Running the app prints no DB config errors (no actual DB usage yet).

---

## Task 02 — Add Postgres connection initialization

**Goal**
Create a reusable DB connection/pool 

**NOTE**: “If config keys already exist, do not change them; only confirm they are wired and used.”

**Files / Paths**
- app bootstrap / main wiring (where dependencies are created)
- `internal/db/...` (new) if needed

**Steps**
1. Initialize a connection pool at startup using config values.
2. Verify connection with a ping/health check.
3. Ensure the pool is injected into handlers/services cleanly (no global vars).

**Done when**
- App starts successfully and confirms DB connectivity (ping OK).

---

## Task 03 — Create TenantSettings repository (GetByID)

**Goal**
Implement one query function:
- Input: `id` (integer)
- Output: a struct containing all columns of `tenant_settings`

**Files / Paths**
- `internal/repository/tenant_settings_repo.go` (new)

**Steps**
1. Create a `TenantSettings` struct matching table columns.
2. Implement `GetTenantSettingsByID(ctx, id)` using a parameterized query:
   ```sql
   SELECT ... FROM tenant_settings WHERE id = $1 AND deleted_at IS NULL;
   ```
3. Handle “not found” cleanly (return a typed error).

**Done when**
- A unit test (or a small local call) can fetch a row by id.

**NOTE**: if you need table structure, you can see this file: `internal/db/sql-query/0000001-create-tenant-setting.sql`
---

## Task 04 — Add API endpoint: GET /v1/otp/tenant-settings/:id

**Goal**
Expose an API endpoint to fetch tenant settings by id.

**API**
- Method: `GET`
- Path: `/v1/otp/tenant-settings/:id`
- Success (200): returns full tenant row as JSON
- Not found (404): standardized error format

**Files / Paths**
- `internal/api/routes.go`
- `internal/api/tenant_settings_handlers.go` (new)

**Steps**
1. Add handler `GetTenantSettingsByIDHandler`.
2. Parse `:id` from path and validate it’s a positive integer.
3. Call repository `GetByID`.
4. Return JSON response with all fields.
5. Use the project’s standard error response mechanism.

**Done when**
- `curl http://localhost:8080/v1/otp/tenant-settings/123` returns row JSON (200)
- unknown id returns 404

---

## Task 05 — Add handler tests (smoke-level)

**Goal**
Test basic API behavior: 200 / 404 / 400.

**Files / Paths**
- `internal/api/tenant_settings_handlers_test.go` (new)

**Steps**
1. Set up a test router.
2. Mock repository (preferred) or use a test DB if project already supports it.
3. Assert:
   - invalid id => 400
   - not found => 404
   - found => 200 and JSON has expected keys

**Done when**
- `go test ./...` passes.

---

## Task 06 — Add minimal docs entry

**Goal**
Document the endpoint under Load/Performance testing docs (or API docs).

**Files / Paths**
- `README.md` or `docs/...`

**Steps**
1. Add:
   - endpoint path
   - example response
2. Mention it’s used for OTP pre-check (tenant lookup) benchmarking.

**Done when**
- Docs include the endpoint reference.

---

## (Optional) Task 07 — Add k6 benchmark script for tenant lookup endpoint

**Goal**
Benchmark Postgres SELECT by hitting the API (end-to-end, realistic).

**Files / Paths**
- `k6/09-tenant-settings-get-by-id.js` (new)

**Steps**
1. Load a list of ids (or use a range).
2. `GET /v1/tenant-settings/:id`
3. Add thresholds (adjust after baseline):
   - `http_req_failed < 1%`
   - `p95 < 50ms`

**Done when**
- Script runs and produces stable metrics.

---

## Important constraints (for Cursor)
- Keep changes minimal and scoped.
- Do not refactor unrelated areas.
- Follow existing project patterns for config, DB, routing, and error responses.
