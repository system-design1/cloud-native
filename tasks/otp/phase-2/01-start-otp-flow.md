We are implementing this project incrementally to avoid large diffs and context/usage limits.
Add OTP domain models, interfaces, config defaults, hashing helper, and service skeleton. Do not implement send/verify fully. 

Implement only Phase 1 of the OTP flow.

Scope:
- Create or update files only under internal/otp.
- Do not modify routes.
- Do not modify cmd/server/main.go.
- Do not modify Redis/PostgreSQL/cache/repository packages.
- Do not create migrations.
- Do not add external dependencies unless absolutely necessary.

Goal:
Add the foundational OTP domain layer.

Please implement:
1. internal/otp/models.go
   - SendRequest
   - SendResponse
   - VerifyRequest
   - VerifyResponse
   - TenantSettings
   - OTPState
   - SMSRequest
   - SMSResult
   - OTPRequestLog
   - OTPProviderResultLog
   - OTPVerificationLog
   - status and reason constants

2. internal/otp/interfaces.go
   - TenantSettingsProvider
   - OTPStore
   - SMSProvider
   - OTPRequestLogger
   - OTPVerificationLogger

3. internal/otp/config.go
   - Config struct
   - DefaultConfig function
   - fields:
     - CodeLength
     - TTL
     - MaxAttempts
     - TenantCacheTTL
     - ProviderTimeout

4. internal/otp/hash.go
   - HashCode helper
   - VerifyCode helper
   - Do not store plaintext OTP
   - Use a stable hash implementation from Go standard library only

5. internal/otp/errors.go
   - domain-level errors:
     - tenant not found
     - tenant disabled
     - otp not found
     - otp expired
     - invalid code
     - max attempts exceeded
     - sms provider failed

6. internal/otp/service.go
   - Service struct
   - NewService constructor
   - dependency fields for the interfaces
   - config field
   - method stubs:
     - SendOTP(ctx context.Context, req SendRequest) (*SendResponse, error)
     - VerifyOTP(ctx context.Context, req VerifyRequest) (*VerifyResponse, error)
   - For now, method bodies can return a clear "not implemented" error.

Important:
- Keep existing OTP benchmark generator behavior intact.
- If internal/otp/otp.go already exists, do not break it.
- Keep code idiomatic Go.
- Keep this as a small, reviewable diff.
- After changes, show the file list and summarize what changed.

------------------

Add focused unit tests only for internal/otp/hash.go.

Requirements:
- Create internal/otp/hash_test.go
- Test HashCode returns deterministic hashes for the same input
- Test different codes produce different hashes
- Test VerifyCode returns true for matching code/hash
- Test VerifyCode returns false for non-matching code/hash
- Test VerifyCode handles malformed hashes safely
- Use table-driven tests where appropriate
- Do not modify production code unless absolutely necessary
- Do not modify any files outside internal/otp
- Keep the diff very small

After implementation:
- run go test ./internal/otp -cover
- summarize the new coverage

note:
to test it, you can run these following commands:
go test ./internal/otp -cover
go test ./internal/otp -coverprofile=coverage.out
go tool cover -func=coverage.out
--------
Before implementing the Redis-backed OTP store, first analyze the current project structure and existing Redis integration patterns.

Do not modify any files yet.

I want to implement the OTPStore interface incrementally and consistently with the current codebase.

Please analyze:
- how Redis clients are currently initialized and organized
- existing Redis helper/util patterns
- whether Redis usage already has conventions for serialization, context handling, logging, or metrics
- where the Redis-backed OTPStore implementation should live
- whether internal/otp/redis_store.go is the right place
- whether Redis-specific code should stay inside internal/otp or use another package
- how IncrementAttempts should be implemented safely
- how Redis key naming should be handled consistently
- whether tests should be unit tests or integration-style tests based on the current repository style

Important:
- Do not implement anything yet
- Do not modify files
- Return only analysis and implementation recommendations
- Keep recommendations aligned with the current project style
----------------
Implement only Phase 2A-1: Redis-backed OTPStore basic operations.

We are implementing incrementally to avoid large diffs and context/usage limits.

Scope:
- Create a Redis OTP store adapter in internal/repository.
- Prefer file name: internal/repository/otp_store_redis.go
- Do not modify routes.
- Do not modify cmd/server/main.go.
- Do not modify PostgreSQL repositories.
- Do not create migrations.
- Do not implement SendOTP or VerifyOTP.
- Do not modify files outside internal/repository and internal/otp unless absolutely necessary.
- Keep the diff small.

Goal:
Implement the basic Redis-backed OTPStore adapter for the existing internal/otp.OTPStore interface.

Implement:
1. A RedisOTPStore struct wrapping *redis.Client.
2. A constructor, for example NewRedisOTPStore(client *redis.Client) *RedisOTPStore.
3. Save(ctx context.Context, state otp.OTPState, ttl time.Duration) error.
4. Get(ctx context.Context, tenantID int64, phone string) (*otp.OTPState, error).
5. Delete(ctx context.Context, tenantID int64, phone string) error.
6. A private key helper using format:
   otp:{tenant_id}:{phone}

Storage model:
- Use Redis Hash, not JSON.
- Store fields:
  - request_id
  - tenant_id
  - phone
  - code_hash
  - attempt_count
  - max_attempts
  - created_at
  - expires_at
- Use HSET + EXPIRE in Save.
- Use HGETALL in Get.
- Use DEL in Delete.
- Do not store plaintext OTP codes.

Error handling:
- Missing key should return otp.ErrOTPNotFound.
- Malformed stored values should return wrapped errors.
- Redis command failures should return wrapped errors.
- Keep low-level adapter quiet; do not add logging or metrics yet.

IncrementAttempts:
- Do not implement the real IncrementAttempts logic in this step.
- If the OTPStore interface requires it, add a stub method on RedisOTPStore that returns otp.ErrNotImplemented or a clear not implemented error.
- We will implement atomic IncrementAttempts with Lua in the next separate step.

Tests:
- Add focused tests only if they are small and consistent with the existing repository test style.
- If Redis integration tests require running Redis, make them skip cleanly when Redis is unavailable.
- Do not add new test dependencies.
- Keep existing tests passing.

After implementation:
- run gofmt
- run go test ./internal/otp
- run go test ./internal/repository
- summarize changed files and any tests added.

Important implementation constraints:
- Keep the adapter simple and idiomatic.
- Avoid introducing generic abstractions or helper packages.
- Do not add interfaces inside internal/repository.
- Do not create a shared Redis utility layer yet.
- Keep Redis field mapping explicit and readable.
- Prefer clarity over abstraction.
- Do not refactor existing Redis benchmark code in this step.
- Keep the implementation easy to review manually.

Before modifying files, briefly explain the exact files you plan to change and why.


