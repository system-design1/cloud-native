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


-----------------------------

Before implementing IncrementAttempts, analyze the current RedisOTPStore implementation and tests.

Do not modify any files yet.

Goal:
Prepare a small, safe implementation plan for atomic OTP attempt incrementing.

Please analyze:
- current RedisOTPStore structure
- current OTPStore interface contract
- current Redis hash field names
- current error handling style
- current test style in otp_store_redis_test.go
- how IncrementAttempts is currently stubbed
- how to implement IncrementAttempts atomically without creating missing keys
- whether Redis Lua script is the right approach here
- what edge cases should be tested

Important constraints:
- Do not implement anything yet
- Do not modify files
- Keep the next implementation diff small
- Do not refactor Save/Get/Delete
- Do not introduce new dependencies
- Keep behavior consistent with internal/otp domain errors

Return:
1. Recommended implementation approach
2. Exact files that should change
3. Test cases to add
4. Risks or edge cases
----------------------------

Implement IncrementAttempts for RedisOTPStore using Redis Lua.

We are implementing incrementally to avoid large diffs and context/usage limits.

Scope:
- Modify only:
  - internal/repository/otp_store_redis.go
  - internal/repository/otp_store_redis_test.go
- Do not modify Save, Get, or Delete unless absolutely necessary.
- Do not modify internal/otp.
- Do not modify routes, cmd/server/main.go, config, migrations, or docs.
- Do not introduce new dependencies.
- Keep the diff small and easy to review.

Goal:
Replace the current IncrementAttempts stub with a real atomic Redis implementation.

Implementation requirements:
- Use a small Redis Lua script.
- The script must:
  1. Check if the OTP key exists.
  2. Return -1 if the key does not exist.
  3. Check if the attempt_count field exists.
  4. Return -2 if attempt_count is missing.
  5. Increment attempt_count using HINCRBY.
  6. Return the new attempt count.
- Map -1 to otp.ErrOTPNotFound.
- Map -2 to a wrapped malformed-state error.
- Redis/Lua failures should return wrapped errors.
- Do not reset or modify TTL.
- Do not create missing OTP keys.
- Do not store plaintext OTP codes.

Tests:
- Replace the current IncrementAttempts not-implemented test.
- Add focused Redis integration tests that skip cleanly when Redis is unavailable.

Required test cases:
1. IncrementAttempts increments from 0 to 1 and then to 2.
2. After incrementing, Get returns OTPState with AttemptCount = 2.
3. IncrementAttempts on a missing key returns otp.ErrOTPNotFound.
4. IncrementAttempts on an existing hash without attempt_count returns an error that is not otp.ErrOTPNotFound.
5. IncrementAttempts on a hash with non-integer attempt_count returns an error that is not otp.ErrOTPNotFound.

Do not add a concurrent test in this step.

Before modifying files:
- Briefly state the exact files you will change and why.

After implementation:
- run gofmt
- run go test -count=1 ./internal/repository -v
- summarize changed files and test results.
-------------

Before implementing tenant settings cache-aside behavior, analyze the current tenant settings repository and Redis integration patterns.

Do not modify any files yet.

Goal:
Design a small, clean cache-aside tenant settings provider for the OTP domain.

Please analyze:
- current tenant settings repository structure
- how tenant settings are modeled today
- current Redis integration style
- where the cache-aside implementation should live
- whether it should live in internal/repository or internal/otp
- how cache keys should be structured
- how tenant settings should be serialized in Redis
- whether TTL handling should be owned by the cache layer
- how cache miss fallback should work
- how cache failures should behave
- whether stale cache handling is needed yet
- what tests should be added

Important constraints:
- Do not implement anything yet
- Do not modify files
- Keep the next implementation diff small
- Do not refactor existing repository code
- Do not introduce generic caching abstractions
- Keep recommendations aligned with the current project structure

Return:
1. Recommended implementation approach
2. Exact files that should change
3. Cache key strategy
4. Cache serialization strategy
5. Error-handling strategy
6. Recommended test cases
7. Risks or edge cases
--------------------

Implement tenant settings cache-aside provider for the OTP flow.

We are implementing incrementally to avoid large diffs and context/usage limits.

Scope:
- Create only:
  - internal/repository/tenant_settings_cache_provider.go
  - internal/repository/tenant_settings_cache_provider_test.go
- Do not modify internal/otp unless there is a compile-time mismatch that must be fixed.
- Do not modify routes.
- Do not modify cmd/server/main.go.
- Do not modify config.
- Do not modify migrations.
- Do not refactor existing repositories.
- Do not introduce generic cache abstractions.
- Do not add new dependencies.
- Keep the diff small and easy to review.

Goal:
Add a cache-aside tenant settings provider that implements otp.TenantSettingsProvider.

Recommended design:
- Define a small unexported source interface in the new provider file:

  type tenantSettingsSource interface {
      GetTenantSettingsByID(ctx context.Context, tenantID int64) (*TenantSettings, error)
  }

- Implement:

  type CachedTenantSettingsProvider struct {
      client *redis.Client
      source tenantSettingsSource
      ttl    time.Duration
  }

- Add constructor:

  func NewCachedTenantSettingsProvider(
      client *redis.Client,
      source tenantSettingsSource,
      ttl time.Duration,
  ) *CachedTenantSettingsProvider

- Add method:

  func (p *CachedTenantSettingsProvider) GetTenantSettings(ctx context.Context, tenantID int64) (*otp.TenantSettings, error)

Behavior:
1. Build Redis key using:
   tenant:{tenant_id}:settings
2. Try Redis GET.
3. If cache hit and JSON unmarshalling succeeds, return cached otp.TenantSettings.
4. If cache miss, fallback to source.GetTenantSettingsByID.
5. If cache hit has malformed JSON, fallback to source.GetTenantSettingsByID and overwrite cache if source succeeds.
6. If Redis GET fails for a reason other than redis.Nil, fallback to source.GetTenantSettingsByID.
7. Map repository TenantSettings to otp.TenantSettings.
8. Store the mapped otp.TenantSettings as JSON in Redis with TTL.
9. If Redis SET fails after source succeeds, do not fail the request.
10. Return source errors when source lookup fails.

Important:
- Cache only otp.TenantSettings, not the full repository TenantSettings.
- Do not cache SMSAPIKey or other repository-only fields.
- Do not add logging or metrics yet.
- Keep Redis errors non-fatal for tenant lookup when source succeeds.
- Keep helper functions private.
- Do not implement cache stampede protection.
- Do not implement stale refresh logic.

Tests:
- Add focused tests using the existing Redis test style.
- Tests should skip cleanly if Redis is unavailable.
- Use a small fake source implementation instead of requiring PostgreSQL.
- Test cases:
  1. Cache hit returns tenant settings and does not call source.
  2. Cache miss calls source, returns mapped tenant settings, and populates Redis.
  3. Malformed cache falls back to source and overwrites cache.
  4. Source error is returned when cache miss and source fails.
  5. Redis SET failure does not need to be simulated in this step.
- Do not add new test dependencies.

Before modifying files:
- Briefly state the exact files you will create and why.

After implementation:
- run gofmt
- run go test -count=1 ./internal/repository -v
- summarize changed files and test results.
-------------
Before implementing the fake SMS provider, analyze the current OTP domain interfaces and service direction.

Do not modify any files yet.

Goal:
Design a small, realistic fake SMS provider implementation for the OTP flow.

Please analyze:
- the current otp.SMSProvider interface
- the current SMSRequest and SMSResult models
- how the fake provider should behave
- where the provider implementation should live
- whether internal/sms is the right package
- whether the fake provider should simulate latency
- how provider request IDs/status should be modeled
- what minimal provider result fields are needed now
- how provider failures should behave
- whether context timeouts/cancellation should be respected
- what tests should be added
- whether any small model adjustments are needed before implementation

Important constraints:
- Do not implement anything yet
- Do not modify files
- Keep the next implementation diff small
- Do not add real SMS integrations
- Do not add queues/background workers
- Do not add retry logic
- Do not add generic provider frameworks
- Keep recommendations aligned with the current project structure

Return:
1. Recommended implementation approach
2. Exact files that should change
3. Whether any OTP models/interfaces need adjustment
4. Recommended fake provider behavior
5. Error-handling recommendations
6. Recommended tests
7. Risks or future extension points

-----

Implement the fake SMS provider.

We are implementing incrementally to avoid large diffs and context/usage limits.

Scope:
- Create only:
  - internal/sms/fake_provider.go
  - internal/sms/fake_provider_test.go
- Do not modify internal/otp unless there is a compile-time mismatch that must be fixed.
- Do not modify repository code.
- Do not modify routes.
- Do not modify cmd/server/main.go.
- Do not modify config.
- Do not create migrations.
- Do not add real SMS integrations.
- Do not add queues, retries, provider registry, logging, or metrics.
- Do not add new dependencies.
- Keep the diff small and easy to review.

Goal:
Add a fake SMS provider that implements otp.SMSProvider.

Implementation requirements:
- Package should be internal/sms.
- Add type FakeProvider.
- Add constructor:
  func NewFakeProvider() *FakeProvider
- Add a private/test-friendly constructor, for example:
  func newFakeProviderWithDelay(minDelay, maxDelay time.Duration) *FakeProvider
- Implement:
  func (p *FakeProvider) SendOTP(ctx context.Context, req otp.SMSRequest) (*otp.SMSResult, error)

Behavior:
- Always succeeds unless context is canceled or timed out.
- NewFakeProvider must simulate random latency between 20ms and 30ms.
- Respect context cancellation and timeout using select with a timer.
- If context is canceled/timed out, return nil and a wrapped context error while preserving errors.Is.
- Provider name should be req.Provider if provided, otherwise "fake".
- Status should be otp.RequestStatusSent if that constant exists and is suitable; otherwise use "sent".
- MessageID should be non-empty, using request ID if available, otherwise standard-library timestamp/randomness.
- RawResponse should include safe provider metadata such as:
  - provider
  - simulated: true
  - request_id
- RawResponse must not include the OTP code.
- SentAt should be time.Now().UTC().

Tests:
- Unit tests only.
- Use the private delay constructor to keep tests fast and deterministic.
- Add tests:
  1. SendOTP success returns result with provider, status, message ID, SentAt, RawResponse.
  2. RawResponse does not expose the OTP code.
  3. Request provider overrides default provider name.
  4. Canceled context returns an error where errors.Is(err, context.Canceled) is true.
  5. Timeout context returns an error where errors.Is(err, context.DeadlineExceeded) is true.
  6. NewFakeProvider is configured with default latency range 20ms to 30ms, if this can be tested without making tests flaky.

Before modifying files:
- Briefly state the exact files you will create and why.

After implementation:
- run gofmt
- run go test -count=1 ./internal/sms -v
- run go test -count=1 ./...
- summarize changed files and test results.

----------


