
# Security Layer Prompt — Auth, Rate Limiting, Headers, CORS

## Overview
You are a senior Go backend engineer. Implement or review the **security layer** for this Go (Gin) REST API.

> **Important:** Some or all security features may already be implemented.  
> First **audit the existing codebase** and **only add or improve what is missing or incomplete**.  
> Do NOT re-implement anything that is already correct.

---

## Security Goals
The service must be **production-grade, cloud-native, and secure by default**, following common backend security best practices.

---

## Task

### 1) Authentication (JWT-based)
If not already implemented, add a **JWT authentication middleware**:

- Responsibilities:
  - Extract JWT from `Authorization: Bearer <token>` header
  - Validate signature, expiration, issuer, and audience (configurable)
  - Attach authenticated user info (e.g., userID, roles) to request context
- Requirements:
  - JWT secret / public key must come from environment variables
  - No hardcoded secrets
  - Clear error responses for unauthorized requests
- Structure:
  - Middleware should be reusable and testable
  - Do not couple auth logic directly to handlers

---

### 2) Authorization Pattern
Implement a clean authorization pattern (if missing):

- Role-based or permission-based checks
- Example:
  - Middleware or helper that checks required roles per route
- Keep it simple but extensible

---

### 3) Rate Limiting
If not already present, add **rate limiting middleware**:

- Apply per-IP or per-client
- Configurable via environment variables:
  - Requests per second/minute
  - Burst size
- Behavior:
  - Return HTTP `429 Too Many Requests` when limit is exceeded
- Ensure it works correctly behind reverse proxies (X-Forwarded-For handling)

---

### 4) CORS Configuration
Add or review **CORS middleware**:

- Configurable allowed:
  - Origins
  - Methods
  - Headers
- Support credentials if needed
- Do NOT use wildcard origins in production unless explicitly configured

---

### 5) Security Headers
Ensure the following HTTP security headers are set (via middleware):

- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection`
- `Referrer-Policy`
- `Content-Security-Policy` (basic, configurable)
- `Strict-Transport-Security` (if HTTPS is assumed behind a proxy)

All header values must be configurable.

---

### 6) Input Validation & Sanitization (Example)
Provide at least one **example** of input validation:

- Validate query params or JSON body
- Reject malformed or unexpected input
- Avoid trusting client input blindly
- Use clear validation errors

---

### 7) Logging & Observability
Security-related events must be logged:

- Authentication failures
- Authorization failures
- Rate limit violations
- Suspicious input

Logs must:
- Be structured JSON (Zerolog)
- Avoid leaking sensitive data (tokens, secrets)

---

### 8) Configuration & Validation
All security-related settings must be:

- Loaded from environment variables
- Validated at startup
- Fail fast if misconfigured (e.g., missing JWT secret)

---

## Expected Output
- Secure, modular middleware for:
  - Authentication
  - Authorization
  - Rate limiting
  - CORS
  - Security headers
- Clear separation of concerns
- No hardcoded secrets or policies
- Updated documentation if new security features are added
