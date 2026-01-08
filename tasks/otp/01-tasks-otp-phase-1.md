# OTP Service – Phase 1 (System Design & Capacity Prep)

This document contains **atomic, Cursor-friendly tasks** for implementing the **first OTP API**.
Each task is intentionally small, single-purpose, and independently executable.

Scope of this phase:
> A single API that **generates a 6-digit OTP code and returns it in the response**.  
No persistence, no SMS, no verification logic.

All tasks **must follow the attached API design guidelines**:
- Clear naming
- Proper HTTP methods
- Versioned API (`/v1`)
- Consistent error handling

---

## Task 01 — Create OTP code generator helper

**Goal**  
Implement a secure helper that generates a 6-digit numeric OTP code.

**Why this task exists**  
OTP generation logic must be isolated, testable, and reusable for later phases.

**Files / Paths**
- `internal/otp/otp.go` (new file)

**Steps**
1. Create package `internal/otp`.
2. Implement function:
   ```go
   func Generate6DigitCode() (string, error)
   ```
3. Use `crypto/rand` (NOT `math/rand`) to generate randomness.
4. Generate a number in range `[0, 999999]`.
5. Format output with leading zeros using `%06d`.
6. Return error if random generation fails.

**Done when**
- Function always returns a **string of exactly 6 digits**.
- Leading zeros are preserved (e.g. `"000123"`).

---

## Task 02 — Implement OTP generation API handler

**Goal**  
Create an HTTP handler that generates and returns a 6-digit OTP.

**API Design**
- Method: `POST`
- Path: `/v1/otp/code`
- Response:
```json
{
  "code": "123456"
}
```

**Files / Paths**
- `internal/api/otp_handlers.go` (new file or existing handlers file)

**Steps**
1. Create handler `GenerateOTPCodeHandler`.
2. Call `otp.Generate6DigitCode()`.
3. On error:
   - Return error using the project's standard error handler/middleware.
4. On success:
   - Return HTTP 200.
   - JSON body must contain only the `code` field.

**Done when**
- Calling the handler returns HTTP 200.
- Response JSON contains a 6-digit `code`.

---

## Task 03 — Register versioned OTP route

**Goal**  
Expose the OTP API under a versioned route group.

**Files / Paths**
- `internal/api/routes.go`

**Steps**
1. Ensure `/v1` route group exists.
2. Create subgroup `/otp`.
3. Register route:
   ```text
   POST /v1/otp/code
   ```
4. Bind it to `GenerateOTPCodeHandler`.

**Done when**
- API is reachable at `/v1/otp/code`.

---

## Task 04 — Add smoke-level API test for OTP endpoint

**Goal**  
Add a minimal test to verify the OTP endpoint behavior.

**Files / Paths**
- `internal/api/otp_handlers_test.go` (new file)

**Steps**
1. Create a Gin router for test only.
2. Register `/v1/otp/code` route.
3. Send a `POST` request (empty body).
4. Assert:
   - HTTP status is `200`.
   - Response contains `code`.
   - `len(code) == 6`.
   - All characters are digits.

**Done when**
- `go test ./...` passes.
- Test validates OTP format only (no business logic).

---

## Task 05 — Add minimal API documentation entry

**Goal**  
Document the new OTP endpoint for future reference and consistency.

**Files / Paths**
- `docs/` or existing API documentation file

**Steps**
1. Add a short section:
   ```text
   POST /v1/otp/code
   Response: { "code": "123456" }
   ```
2. Mention that this is **Phase 1: generation only**.

**Done when**
- Documentation clearly lists the endpoint and response shape.

---

## Execution Notes (Important for Cursor)

- Execute **one task at a time**.
- Do NOT combine tasks.
- Do NOT introduce storage, Redis, SMS, or verification logic.
- Do NOT refactor unrelated files.
- Keep changes minimal and scoped.

---

## End of Phase 1

Next phases will build on this API:
- OTP persistence
- Idempotency
- Rate limiting
- Verification
- Capacity & stress testing per component
