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
Before implementing SendOTP orchestration, analyze the current OTP domain, Redis OTP store, tenant cache provider, and fake SMS provider.

Do not modify any files yet.

Goal:
Design a small, clean first implementation of otp.Service.SendOTP.

Please analyze:
- the current otp.Service structure
- current OTP interfaces and models
- current Redis OTP store behavior
- current tenant settings provider behavior
- current fake SMS provider behavior
- current config structure
- what the first SendOTP implementation should do step-by-step
- what ordering of operations is safest
- what should happen if Redis save fails
- what should happen if SMS provider fails
- whether OTP should be generated before or after tenant validation
- whether provider timeout handling should be done inside the service
- how request IDs should be handled initially
- whether OTP request logging should be included now or deferred
- whether VerifyOTP should remain stubbed in this phase
- what tests should be added
- what edge cases should be handled now vs deferred

Important constraints:
- Do not implement anything yet
- Do not modify files
- Keep the next implementation diff small
- Do not add routes or handlers yet
- Do not add metrics or tracing yet
- Do not add database request logging yet
- Do not add rate limiting yet
- Do not add retries
- Keep the implementation aligned with the current project structure
- Prefer correctness and clarity over abstraction

Return:
1. Recommended SendOTP flow
2. Recommended ordering of operations
3. Which failures should abort the flow
4. Which failures should be tolerated
5. Exact files that should change
6. Recommended tests
7. Deferred concerns that should not be implemented yet

-----------

Implement the first version of otp.Service.SendOTP orchestration.

We are implementing incrementally to avoid large diffs and context/usage limits.

Scope:
- Modify only:
  - internal/otp/service.go
  - internal/otp/service_test.go
- Modify internal/otp/errors.go only if absolutely necessary for a clear domain error.
- Do not modify routes.
- Do not modify handlers.
- Do not modify repository code.
- Do not modify sms package.
- Do not modify config loader.
- Do not create migrations.
- Do not add metrics or tracing.
- Do not add PostgreSQL request logging yet.
- Do not implement VerifyOTP.
- Do not add retries, queues, idempotency, rate limiting, or token validation.
- Keep the diff small and easy to review.

Goal:
Implement SendOTP using the already-defined ports:
- TenantSettingsProvider
- OTPStore
- SMSProvider

Recommended flow:
1. Validate request:
   - TenantID must be greater than 0
   - Phone must not be empty
2. Load tenant settings using TenantSettingsProvider.
3. Validate tenant:
   - tenant must be active
   - OTPEnabled must be true
   - if ExpiresAt is present and expired, treat tenant as disabled
4. Generate a request ID.
5. Generate OTP code using GenerateCode(config.CodeLength).
6. Hash OTP code using HashCode.
7. Build OTPState.
8. Save OTPState using OTPStore.Save with config.TTL.
9. Send SMS using SMSProvider.SendOTP.
10. Use context.WithTimeout for provider call using config.ProviderTimeout.
11. Return SendResponse with RequestID and ExpiredAt.

Important behavior:
- Tenant validation happens before OTP generation.
- Redis/store save happens before SMS sending.
- If OTPStore.Save fails, abort and do not call SMS provider.
- If SMSProvider.SendOTP fails or times out, return an error.
- Do not delete/rollback Redis OTP state on SMS failure in this step.
- Do not expose or log plaintext OTP.
- VerifyOTP must remain a stub.
- Request ID only needs to be non-empty and consistent across OTPState and SMSRequest.
- Use an existing uuid package if already present in go.mod; do not add a new dependency.

Tests:
Create focused unit tests in internal/otp/service_test.go using fake dependencies.
Do not use Redis/Postgres/SMS infrastructure.

Required tests:
1. SendOTP success:
   - returns non-empty RequestID
   - returns non-zero ExpiredAt
   - saves OTP state
   - saved CodeHash is non-empty
   - saved CodeHash is not plaintext OTP
   - saved tenant ID and phone match request
   - saved max attempts and expiration are correct
   - calls SMS provider
   - SMS request uses same RequestID, TenantID, Phone, Provider
2. Invalid request:
   - TenantID <= 0 returns error
   - empty Phone returns error
   - store and SMS are not called
3. Tenant lookup error:
   - returns error
   - store and SMS are not called
4. Tenant disabled:
   - OTPEnabled=false returns ErrTenantDisabled
   - inactive status returns ErrTenantDisabled if status model supports it
   - store and SMS are not called
5. Store save error:
   - returns error
   - SMS is not called
6. SMS provider error:
   - returns error
   - store was called before SMS
7. SMS provider timeout:
   - returns error preserving context deadline semantics if practical
   - store was called before SMS

Before modifying files:
- Briefly state the exact files you will change and why.

After implementation:
- run gofmt
- run go test -count=1 ./internal/otp -v
- run go test -count=1 ./...
- summarize changed files and test results.
----------------

Before implementing PostgreSQL OTP request logging, analyze the current repository, migrations, and otp.Service.SendOTP flow.

Do not modify any files yet.

Goal:
Design a small, incremental implementation for OTP request/provider result logging.

Please analyze:
- current migration structure and naming conventions
- current PostgreSQL repository style
- current tenant_settings repository patterns
- current otp.OTPRequestLogger interface
- current otp.OTPRequestLog and OTPProviderResultLog models
- current SendOTP flow and where logging should happen
- whether logging should be mandatory or best-effort in this phase
- what database table/schema is needed
- whether request creation and provider result update should be one table or separate tables
- how request_id, tenant_id, phone, status, provider_name, provider_response, error_message, metadata, correlation_id should be stored
- what indexes are needed
- what tests should be added
- what should be deferred

Important constraints:
- Do not implement anything yet
- Do not modify files
- Keep the next implementation diff small
- Do not modify routes or handlers
- Do not add metrics/tracing yet
- Do not implement VerifyOTP
- Do not add async queues/outbox/retries
- Keep repository code consistent with current project style
- Prefer simple PostgreSQL implementation using database/sql

Return:
1. Recommended logging design
2. Exact files that should change
3. Proposed migration/table schema
4. Repository implementation approach
5. Where logging should later be called from SendOTP
6. Recommended tests
7. Deferred concerns

----------------
Implement PostgreSQL OTP request logging repository and migration.

We are implementing incrementally to avoid large diffs and context/usage limits.

Scope:
- Add only:
  - one sql-migrate migration for otp_requests
  - internal/repository/otp_request_log_repo.go
  - internal/repository/otp_request_log_repo_test.go
- Do not modify otp.Service.SendOTP yet.
- Do not modify routes or handlers.
- Do not modify cmd/server/main.go.
- Do not modify config.
- Do not add metrics/tracing.
- Do not implement VerifyOTP.
- Do not add async queues, outbox, retries, or idempotency.
- Keep the diff small and easy to review.

Migration requirements:
- This project uses github.com/rubenv/sql-migrate.
- First inspect existing migration files and follow the exact sql-migrate format used in this project.
- First inspect the Makefile and identify the existing make targets for running migrations.
- Do not invent new migration commands.
- Create one migration for otp_requests using the existing migration directory and naming convention.
- Include both Up and Down sections if the existing sql-migrate files use them.
- After creating the migration, run the existing make migration command if available.
- If migrations cannot be run, explain exactly why and what command should be run manually.

Table requirements:
- Create table otp_requests if it does not exist.
- Columns:
  - id BIGSERIAL PRIMARY KEY
  - request_id TEXT NOT NULL
  - tenant_id BIGINT NOT NULL
  - phone TEXT NOT NULL
  - status TEXT NOT NULL
  - provider_name TEXT NOT NULL DEFAULT ''
  - provider_response JSONB NOT NULL DEFAULT '{}'::jsonb
  - error_message TEXT
  - metadata JSONB NOT NULL DEFAULT '{}'::jsonb
  - correlation_id TEXT
  - created_at TIMESTAMPTZ NOT NULL DEFAULT now()
  - updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
- Add unique index on request_id.
- Add indexes:
  - tenant_id, created_at DESC
  - phone, created_at DESC
  - status, created_at DESC
- If existing migrations use updated_at triggers, follow the same style.

Repository requirements:
- Create OTPRequestLogRepository using database/sql.
- Constructor:
  func NewOTPRequestLogRepository(db *sql.DB) *OTPRequestLogRepository
- Implement existing otp.OTPRequestLogger interface:
  func (r *OTPRequestLogRepository) CreateRequest(ctx context.Context, log otp.OTPRequestLog) error
  func (r *OTPRequestLogRepository) UpdateProviderResult(ctx context.Context, log otp.OTPProviderResultLog) error
- Use parameterized SQL.
- Marshal Metadata and ProviderResponse to JSON.
- Treat nil maps as empty JSON objects.
- Wrap errors with clear context.
- UpdateProviderResult should match by request_id.
- If UpdateProviderResult affects zero rows, return an error.

Testing requirements:
- Follow current repository integration test style.
- Tests may use real PostgreSQL and skip cleanly if unavailable.
- Do not add new dependencies.
- Do not create schema in Go tests unless existing repository tests already do that.
- Before running repository tests, ensure the otp_requests migration has been applied.
- Run:
  - gofmt
  - the relevant existing make migration command, if available
  - go test -count=1 ./internal/repository -v
  - go test -count=1 ./...

Add tests:
1. CreateRequest inserts a row and persisted fields can be queried.
2. UpdateProviderResult updates status, provider_name, provider_response, error_message.
3. UpdateProviderResult on unknown request_id returns error.
4. Optional only if small: duplicate request_id returns error.

Before modifying files:
- Briefly state the exact files you will create and why.
- Briefly state which existing migration make target you found.

After implementation:
- summarize changed files
- summarize migration command used
- summarize test results

for sql-migrate, you must use from: `https://github.com/rubenv/sql-migrate`

-----

OTP domain foundation ✅
- models
- interfaces
- config defaults
- errors
- hashing
- dynamic OTP generation

Redis OTP store ✅
- Save/Get/Delete
- Redis Hash storage
- atomic IncrementAttempts with Lua

Tenant settings cache provider ✅
- Redis cache-aside
- PostgreSQL fallback
- JSON cache
- fake-source tests

Fake SMS provider ✅
- 20ms تا 30ms delay
- context cancellation/timeout
- safe RawResponse

SendOTP service orchestration ✅
- tenant validation
- OTP generation
- hashing
- Redis save
- SMS send with timeout
- unit tests

OTP request logging repository ⏳
- migration file ساخته شده
- repository ساخته شده
- tests ساخته شده
- ولی migration هنوز روی DB اجرا نشده

------
Before wiring OTP request logging into SendOTP, analyze the current otp.Service.SendOTP implementation and the new OTPRequestLogRepository/interface.

Do not modify any files yet.

Goal:
Design a small, safe implementation plan for adding request/provider logging into SendOTP.

Please analyze:
- current otp.Service fields and constructor
- current OTPRequestLogger interface
- current OTPRequestLog and OTPProviderResultLog models
- current SendOTP ordering
- where CreateRequest should happen
- where UpdateProviderResult should happen
- what status values should be used
- how metadata and correlation_id should be handled for now
- what should happen if CreateRequest fails
- what should happen if OTPStore.Save fails after CreateRequest succeeds
- what should happen if SMSProvider.SendOTP fails after OTPStore.Save succeeds
- what should happen if UpdateProviderResult fails after SMS success
- whether request logging should be mandatory in this phase
- how tests should be updated

Important constraints:
- Do not implement anything yet
- Do not modify files
- Keep the next implementation diff small
- Modify only internal/otp/service.go and internal/otp/service_test.go in the next implementation step unless absolutely necessary
- Do not modify repository code
- Do not modify routes or handlers
- Do not modify cmd/server/main.go
- Do not add metrics/tracing
- Do not implement VerifyOTP
- Do not add async queues/outbox/retries
- Keep behavior clear and testable

Return:
1. Recommended logging flow inside SendOTP
2. Failure policy for each logging point
3. Exact files that should change
4. Recommended test updates
5. Deferred concerns

--------------
Implement OTP request/provider logging inside otp.Service.SendOTP.

Scope:
- Modify only:
  - internal/otp/service.go
  - internal/otp/service_test.go
- Do not modify repositories
- Do not modify migrations
- Do not modify routes/handlers
- Do not modify cmd/server/main.go
- Do not implement VerifyOTP
- Do not add metrics/tracing/outbox/retries

Requirements:

1. Add request logging into SendOTP flow.

Flow order must be:

1. validate request
2. load tenant settings
3. validate tenant
4. generate request ID
5. generate OTP code
6. create OTP request log with status "pending"
7. save OTP state in Redis
8. send SMS
9. update provider result log
10. return response

2. Request logging behavior

If requestLogger is nil:
- skip logging entirely
- preserve existing behavior

CreateRequest must:
- happen before Redis save
- use status "pending"

CreateRequest failure:
- abort the flow
- do not call store
- do not call SMS provider

3. Store failure behavior

If OTP store save fails:
- attempt UpdateProviderResult with:
  - status: "failed"
  - provider_name: tenant.SMSProvider
  - error_message: wrapped save error
- SMS provider must NOT be called
- return original store error

4. SMS provider failure behavior

If SMS provider fails or times out:
- attempt UpdateProviderResult with:
  - status: "failed"
  - provider_name: tenant.SMSProvider
  - error_message: provider error string
- preserve:
  - ErrSMSProviderFailed
  - context.DeadlineExceeded
  - context.Canceled
- return wrapped provider error

5. Success behavior

After SMS success:
- call UpdateProviderResult with:
  - status: "sent"
  - provider_name: tenant.SMSProvider
  - provider_response containing ONLY:
    - provider
    - status
    - message_id
    - sent_at

If UpdateProviderResult fails after SMS success:
- return the update error
- do not silently ignore it

Consistency of request audit logging is prioritized in this phase.

6. Correlation ID

- set correlation_id to empty string for now
- do not add context propagation yet

7. Tests

Add/update tests for:
- success path logging
- invalid request does not log
- tenant lookup error does not log
- disabled tenant does not log
- CreateRequest failure aborts flow
- store save failure updates failed status
- SMS provider failure updates failed status
- SMS provider timeout preserves context deadline exceeded
- success update failure returns error

Important:
- keep tests deterministic
- keep diff small
- keep code idiomatic Go
- gofmt all changed files
- after implementation, summarize:
  - changed files
  - behavior changes
  - added tests
----------
Until this section, we have completed these domains:
OTP domain models
OTP config
hashing
Redis OTP store
atomic increment attempts
tenant cache provider
fake SMS provider
SendOTP orchestration
PostgreSQL request logging
provider result logging
unit/integration tests

-----------------------------------
Analyze the next implementation step for otp.Service.VerifyOTP.

Current state:
- SendOTP is implemented and tested.
- OTP state is stored in Redis via OTPStore.
- OTP code is stored only as a hash.
- IncrementAttempts is implemented atomically with Redis Lua.
- OTP request/provider logging already exists.
- VerifyOTP is still stubbed.
- No HTTP handlers/routes yet.

I want the next step to implement VerifyOTP inside internal/otp/service.go.

Please analyze:

1. Recommended VerifyOTP flow ordering.
2. Which validations should happen before IncrementAttempts.
3. Whether IncrementAttempts should happen before or after hash verification.
4. How expired OTPs should behave.
5. How max attempts should behave.
6. Whether successful verification should delete Redis OTP state.
7. Which failures should abort immediately.
8. Which logging should happen now vs later.
9. Exact files that should change.
10. Recommended unit tests.
11. Edge cases and race conditions.
12. Recommended behavior for repeated successful verify attempts.
13. Whether VerifyOTP should be idempotent or one-time-use.
14. Whether Redis delete failure after successful verification should fail the request.

Important:
- Keep the implementation small and reviewable.
- Do not add handlers/routes yet.
- Do not add metrics/tracing.
- Do not add async workers.
- Do not add rate limiting yet beyond MaxAttempts already stored in OTPState.
- Do not change repository interfaces unless absolutely necessary.
- Prefer consistency with the current SendOTP design and existing tests.
-----

Implement the first version of otp.Service.VerifyOTP.

We are implementing incrementally to avoid large diffs and context/usage limits.

Scope:
- Modify only:
  - internal/otp/service.go
  - internal/otp/service_test.go
- Modify internal/otp/models.go or errors.go only if absolutely necessary for missing reason constants.
- Do not modify repository code.
- Do not modify sms package.
- Do not modify routes or handlers.
- Do not modify cmd/server/main.go.
- Do not add metrics/tracing.
- Do not add verification logging yet.
- Do not add new database tables/migrations.
- Do not add async workers/retries/rate limiting/idempotency.
- Keep the diff small and easy to review.

Goal:
Implement VerifyOTP using the existing OTPStore interface.

Required flow:
1. Validate request:
   - TenantID must be greater than 0
   - Phone must not be empty
   - Code must not be empty
2. Load OTP state using OTPStore.Get.
3. If OTPStore.Get returns ErrOTPNotFound:
   - return VerifyResponse{Verified:false, Reason:not_found}, nil
   - RequestID can be empty because no state exists
4. If OTPStore.Get returns another error:
   - return nil, wrapped error
5. Check expiration:
   - if now is equal to or after state.ExpiresAt, treat as expired
   - best-effort Delete
   - return VerifyResponse{Verified:false, RequestID:state.RequestID, Reason:expired}, nil
6. Determine max attempts:
   - use state.MaxAttempts
   - if state.MaxAttempts <= 0, fall back to service config MaxAttempts
7. Check pre-existing attempts:
   - if state.AttemptCount >= maxAttempts:
     - best-effort Delete
     - return VerifyResponse{Verified:false, RequestID:state.RequestID, Reason:max_attempts_exceeded}, nil
8. Verify code:
   - use VerifyCode(req.Code, state.CodeHash)
9. If code is invalid:
   - call OTPStore.IncrementAttempts
   - if IncrementAttempts returns ErrOTPNotFound, return VerifyResponse{Verified:false, Reason:not_found}, nil
   - if IncrementAttempts returns another error, return nil, wrapped error
   - if new attempt count >= maxAttempts:
     - best-effort Delete
     - return VerifyResponse{Verified:false, RequestID:state.RequestID, Reason:max_attempts_exceeded}, nil
   - otherwise return VerifyResponse{Verified:false, RequestID:state.RequestID, Reason:invalid_code}, nil
10. If code is valid:
   - call OTPStore.Delete
   - if Delete fails, return nil, wrapped error
   - return VerifyResponse{Verified:true, RequestID:state.RequestID}, nil

Important behavior:
- IncrementAttempts must happen only for a wrong code.
- Do not increment for expired OTP.
- Do not increment for already exhausted OTP.
- Do not increment for correct code.
- Successful verification is one-time-use.
- Do not implement atomic compare-and-delete yet.
- Accept the current race condition for concurrent correct verification attempts; document nothing unless already useful.
- Do not log verification results yet.

Reason constants:
- Use existing reason constants if they exist.
- If missing, add only the minimal constants needed:
  - not_found
  - expired
  - invalid_code
  - max_attempts_exceeded
- Keep naming consistent with existing models.

Tests:
Add focused unit tests in internal/otp/service_test.go using fakeOTPStore.

Required tests:
1. success:
   - returns Verified=true
   - returns same RequestID
   - calls Delete
   - does not call IncrementAttempts
2. invalid request:
   - invalid tenant ID
   - empty phone
   - empty code
   - store Get not called
3. not found:
   - OTPStore.Get returns ErrOTPNotFound
   - returns Verified=false, Reason:not_found, nil error
4. store get error:
   - returns error
5. expired OTP:
   - returns Verified=false, Reason:expired
   - calls Delete best-effort
   - does not increment
6. max attempts already reached:
   - returns Verified=false, Reason:max_attempts_exceeded
   - calls Delete best-effort
   - does not increment
7. invalid code under max:
   - calls IncrementAttempts
   - returns Verified=false, Reason:invalid_code
   - does not delete
8. invalid code reaches max:
   - IncrementAttempts returns maxAttempts
   - returns Verified=false, Reason:max_attempts_exceeded
   - calls Delete best-effort
9. increment error:
   - returns error
10. increment returns ErrOTPNotFound:
   - returns Verified=false, Reason:not_found, nil error
11. successful delete error:
   - returns error
12. max attempts fallback:
   - if state.MaxAttempts <= 0, service config MaxAttempts is used

Before modifying files:
- Briefly state the exact files you will change and why.

After implementation:
- run gofmt
- run go test -count=1 ./internal/otp -v
- run go test -count=1 ./...
- summarize changed files and test results.

-----------
OTP domain foundation ✅
OTP generation + hashing ✅
Redis OTP store ✅
Atomic IncrementAttempts ✅
Tenant settings cache provider ✅
Fake SMS provider ✅
SendOTP service ✅
OTP request logging repository ✅
Request/provider logging داخل SendOTP ✅
VerifyOTP service ✅
HTTP handlers/routes ❌
main.go wiring ❌
config/env wiring کامل ❌
TODO: Add OTP send resend cooldown per tenant+phone to prevent repeated send requests.
verification logging repository ❌
metrics/tracing business-level ❌
manual end-to-end API test ❌
-------------------
Before implementing OTP HTTP handlers and routes, analyze the current API layer, route setup, server wiring, and OTP service implementation.

Do not modify any files yet.

Goal:
Design a small, incremental implementation plan for exposing SendOTP and VerifyOTP through HTTP endpoints.

Current state:
- otp.Service.SendOTP is implemented and tested.
- otp.Service.VerifyOTP is implemented and tested.
- Redis OTP store exists.
- tenant settings cache provider exists.
- fake SMS provider exists.
- OTP request logging repository exists.
- No HTTP handlers/routes for real OTP send/verify exist yet.

Please analyze:
1. Current internal/api package structure.
2. Current route registration style.
3. Current handler style.
4. Current validation/binding/error response patterns.
5. How dependencies are currently passed into routes/handlers.
6. Whether OTP handlers should live in internal/api or another package.
7. Which files should change for the first HTTP slice.
8. Whether cmd/server/main.go should be wired in this step or deferred.
9. How to expose:
   - POST /v1/otp/send
   - POST /v1/otp/verify
10. Request/response JSON shape for both endpoints.
11. Error mapping strategy:
   - validation errors
   - tenant disabled
   - OTP not found/expired/invalid/max attempts
   - infrastructure errors
12. Whether to use existing centralized error middleware or return responses directly.
13. What tests should be added.
14. What should be deferred.

Important constraints:
- Do not implement anything yet.
- Do not modify files.
- Keep the next implementation diff small.
- Do not add metrics/tracing yet.
- Do not add authentication/token validation yet.
- Do not add rate limiting yet.
- Do not add verification DB logging yet.
- Do not change service business logic unless analysis finds a compile-time mismatch.
- Prefer consistency with the existing API code style.

Return:
1. Recommended HTTP implementation approach.
2. Exact files that should change.
3. Whether to wire dependencies in cmd/server/main.go now or in a separate step.
4. Proposed request/response structs.
5. Error response mapping.
6. Recommended tests.
7. Deferred concerns.

------
Implement the first HTTP slice for OTP send/verify handlers and route registration.

We are implementing incrementally to avoid large diffs and context/usage limits.

Scope:
- Modify only:
  - internal/api/routes.go
  - internal/api/otp_flow_handlers.go
  - internal/api/otp_flow_handlers_test.go
- Do not modify cmd/server/main.go yet.
- Do not wire real Redis/Postgres/SMS dependencies yet.
- Do not modify otp.Service business logic.
- Do not modify repository code.
- Do not modify config.
- Do not add metrics/tracing/auth/rate limiting.
- Do not add verification DB logging.
- Keep the diff small and easy to review.

Goal:
Expose service-level SendOTP and VerifyOTP through thin Gin handlers.

Requirements:

1. Handler service abstraction
- In internal/api, define a small unexported interface for handler testing, for example:
  type otpFlowService interface {
      SendOTP(ctx context.Context, req otp.SendRequest) (*otp.SendResponse, error)
      VerifyOTP(ctx context.Context, req otp.VerifyRequest) (*otp.VerifyResponse, error)
  }
- *otp.Service should naturally satisfy this interface.
- Do not change internal/otp just for handler testing.

2. Handlers
Create:
- SendOTPHandler(service otpFlowService) gin.HandlerFunc
- VerifyOTPHandler(service otpFlowService) gin.HandlerFunc

Handler responsibilities:
- bind JSON
- validate required fields at HTTP boundary
- call service
- return JSON response
- map known service errors to HTTP responses

3. Endpoints
Register:
- POST /v1/otp/send
- POST /v1/otp/verify

4. Routes
- Update SetupRoutes minimally to accept an optional OTP flow service if consistent with current route style.
- If otp service is nil, do not register OTP send/verify routes.
- Keep existing benchmark and health routes unchanged.

5. Request DTOs
Use API-local request structs:
send:
  phone string
  tenant_id int64
  token string
  metadata map[string]interface{}

verify:
  tenant_id int64
  phone string
  code string

Map them to otp.SendRequest and otp.VerifyRequest.

6. Response behavior
Send success:
- 200 OK with otp.SendResponse

Verify success or business failure:
- 200 OK with otp.VerifyResponse
- This includes verified=false reasons such as not_found, expired, invalid_code, max_attempts_exceeded.

7. Error mapping
- Invalid JSON or missing required fields -> 400
- otp.ErrTenantDisabled -> 403
- otp.ErrTenantNotFound -> 404
- otp.ErrSMSProviderFailed -> 502 if existing error package supports custom status; otherwise use the closest existing error style without refactoring pkg/errors
- Other errors -> 500
- Prefer consistency with existing internal/api error handling and middleware style.
- Do not refactor the error package.

8. Tests
Add focused handler tests without Redis/Postgres:
- Use fake otpFlowService.

Send handler tests:
- valid request returns 200 and JSON response
- invalid JSON returns 400
- missing tenant_id returns 400
- empty phone returns 400
- tenant disabled maps to 403
- provider failure maps to selected status, preferably 502 if supported
- generic service error maps to 500

Verify handler tests:
- valid request returns 200 and JSON response
- verified=false business response returns 200
- invalid JSON returns 400
- missing tenant_id returns 400
- empty phone returns 400
- empty code returns 400
- generic service error maps to 500

Important:
- Before modifying files, briefly state exact files you will change and why.
- After implementation, run:
  - gofmt
  - go test -count=1 ./internal/api -v
  - go test -count=1 ./...
- Summarize changed files and test results.

----------
Before wiring real OTP dependencies into cmd/server/main.go, analyze the current application startup and dependency initialization flow.

Do not modify any files yet.

Goal:
Design a small, safe wiring plan so POST /v1/otp/send and POST /v1/otp/verify are registered and usable in the running server.

Current state:
- OTP service SendOTP and VerifyOTP are implemented.
- RedisOTPStore exists.
- CachedTenantSettingsProvider exists.
- Fake SMS provider exists.
- OTPRequestLogRepository exists.
- OTP HTTP handlers/routes exist.
- SetupRoutes accepts optional OTP service.
- cmd/server/main.go has not been wired yet.

Please analyze:
1. How PostgreSQL is initialized in main.go.
2. How Redis is initialized in main.go.
3. How existing repositories are constructed.
4. How SetupRoutes is currently called.
5. Which OTP dependencies need to be constructed.
6. Where config values for OTP should come from now.
7. Whether to use otp.DefaultConfig for now or existing config fields.
8. How to pass otp.Service into SetupRoutes.
9. Whether lifecycle/shutdown needs changes.
10. What tests or manual checks should be run.
11. What should be deferred.

Important constraints:
- Do not implement anything yet.
- Do not modify files.
- Keep the next implementation diff small.
- Do not add new config/env wiring unless absolutely necessary.
- Do not modify repository implementations.
- Do not modify OTP business logic.
- Do not modify handlers unless a compile-time mismatch is found.
- Do not add metrics/tracing/auth/rate limiting.
- Do not implement verification DB logging.
- Prefer using existing clients and constructors.

Return:
1. Recommended wiring approach.
2. Exact files that should change.
3. Dependency construction order.
4. Config/default strategy.
5. Manual validation steps after implementation.
6. Risks or edge cases.

--------

Wire real OTP dependencies into cmd/server/main.go.

We are implementing incrementally to avoid large diffs and context/usage limits.

Scope:
- Modify only:
  - cmd/server/main.go
- Do not modify internal/api.
- Do not modify internal/otp.
- Do not modify internal/repository.
- Do not modify internal/sms.
- Do not modify config/env loading.
- Do not modify routes/handlers.
- Do not modify migrations.
- Do not add metrics/tracing/auth/rate limiting.
- Do not implement verification logging.
- Keep the diff small and easy to review.

Goal:
Register the real OTP send/verify HTTP endpoints in the running server by constructing otp.Service and passing it to api.SetupRoutes.

Implementation requirements:
1. Use existing PostgreSQL and Redis clients already initialized in main.go.
2. Reuse existing tenant settings repository if already constructed. Do not create duplicate variables with conflicting names.
3. Use otp.DefaultConfig() for now.
4. Construct:
   - tenant settings repository if not already available
   - CachedTenantSettingsProvider using Redis + tenant settings repository + otpConfig.TenantCacheTTL
   - RedisOTPStore using Redis
   - Fake SMS provider
   - OTPRequestLogRepository using PostgreSQL
   - otp.Service with verifyLogger nil
5. Pass otpService into api.SetupRoutes using the existing optional OTP route registration.
6. Do not change lifecycle/shutdown behavior.
7. Do not add new config fields.

Important:
- Before modifying files, briefly state the exact place in main.go where you will wire this and why.
- Keep existing benchmark routes and health routes unchanged.
- Keep existing startup behavior unchanged.
- If there is a compile-time mismatch with SetupRoutes, make the minimal fix only if absolutely necessary.

After implementation:
- run gofmt on cmd/server/main.go
- run go test -count=1 ./...
- summarize changed files and test results.

----------

I found a possible bug in the OTP send flow and I want you to debug it carefully before changing anything.

Current behavior:

1. I successfully sent OTP requests for valid tenants.
2. Then I sent a request with a tenant_id that does not exist in the database.
3. After that, every subsequent request started returning:

{
  "error": "Forbidden",
  "message": "Tenant is disabled",
  "code": 403
}

even when:
- I changed the phone number
- I used another request
- the tenant should be valid

Example request:

curl --request POST \
  --url http://localhost:8080/v1/otp/send \
  --header 'content-type: application/json' \
  --data '{
    "tenant_id": 200,
    "phone": "+989121234568",
    "metadata": {
      "source": "manual-test"
    }
}'

I do NOT want you to immediately patch the code.

First:
1. Analyze the possible root causes.
2. Inspect the current SendOTP flow carefully.
3. Inspect tenant cache behavior carefully.
4. Inspect whether invalid tenant results are being cached incorrectly.
5. Inspect whether tenant status validation logic is wrong.
6. Inspect whether Redis cache keys or cache overwrite logic can poison future requests.
7. Explain the exact root cause with code references.
8. Tell me the minimal correct fix.
9. Only after the analysis, propose the implementation plan.

Do not make broad refactors.
Do not redesign the architecture.
Focus only on the root cause and the smallest safe fix.

-------------
Before implementing anything, analyze a safe local-development way to inspect the plaintext OTP code generated by SendOTP for manual VerifyOTP testing.

Current problem:
- Fake SMS provider intentionally does not expose OTP code.
- SendOTP stores only HashCode(code) in Redis.
- For manual end-to-end testing of /v1/otp/verify, I need a temporary local/dev-only way to retrieve the generated plaintext OTP code.
- I do NOT want to expose OTP code in normal API responses.
- I do NOT want to log OTP code in normal logs.
- I do NOT want this to be unsafe for production.

Do not modify files yet.

Please analyze possible approaches:

1. Add dev-only Redis debug key
   - Example key: debug:otp-code:{tenant_id}:{phone}
   - Store plaintext code with very short TTL
   - Only enabled by explicit config/env flag

2. Add debug field to fake SMS provider result
   - Only if enabled by explicit config/env flag
   - Must not be written to normal request logs unless explicitly intended

3. Add dev-only endpoint
   - Example: GET /debug/otp-code
   - Discuss why this may be risky and whether to avoid it

4. Add database debug table
   - Discuss why this is likely too heavy/risky for this phase

5. Add local CLI/script-only helper
   - Discuss whether this is better than code changes

Please answer:
- Which approach is safest and smallest for this project right now?
- Where should it live architecturally?
- Should this be implemented inside otp.Service, fake SMS provider, or a separate debug sink/port?
- How should it be enabled/disabled?
- What default should be used?
- What Redis key and TTL should be used if Redis is chosen?
- How do we guarantee it does not run in production accidentally?
- What tests should be added?
- What files would need to change?
- What should be deferred?

Important constraints:
- Do not expose OTP code in SendOTP HTTP response.
- Do not add a public debug endpoint unless strongly justified.
- Do not store plaintext OTP in the main OTP state.
- Do not store plaintext OTP in otp_requests logs.
- Keep production behavior secure by default.
- Keep the next implementation small and reversible.
- Prefer no code changes if a safe manual workflow is enough.
- If code changes are recommended, propose the smallest safe slice only.

Return:
1. Recommended approach
2. Security risks
3. Exact implementation plan
4. Files that would change
5. Tests to add
6. Manual usage workflow

I think the best design is: 
dev-only OTP debug sink
disabled by default
stores plaintext code in Redis debug key with TTL <= OTP TTL
never returned in API response
never stored in otp_requests

------------

Review your previous recommendation for the dev-only OTP plaintext debug capture.

I think the proposed service-level OTPDebugCodeSink may be too large for the next slice because it touches:
- internal/otp/interfaces.go
- internal/otp/service.go
- internal/otp/service_test.go
- internal/repository
- cmd/server/main.go

Please re-analyze and compare two options:

Option A:
Service-level OTPDebugCodeSink port.

Option B:
Fake SMS provider captures the plaintext OTP code into a dev-only Redis debug key, because SMSRequest.Code already reaches the fake provider.

Context:
- This is only for local manual testing.
- We do not want to expose the OTP code in API responses.
- We do not want to store it in otp_requests.
- We do not want to log it.
- We want a very small, reversible implementation.
- Fake SMS provider is already clearly a simulation component.
- Production real providers would not include this behavior.

Please analyze:
1. Which option produces the smallest safe diff?
2. Which option better preserves otp.Service cleanliness?
3. Which option is less risky for production?
4. Which files would change for each option?
5. How to enable it safely only in local/dev?
6. Whether Redis should be injected into fake provider or whether another sink abstraction is still worth it.
7. Whether this should be implemented now or deferred.
8. If implementing now, propose the smallest safe implementation plan.

Important:
- Do not modify files yet.
- Do not implement anything.
- Prefer local/dev-only safety and small diff.
- Do not add debug HTTP endpoints.
- Do not expose OTP code in SendOTP response.
- Do not store plaintext OTP in otp_requests or main OTP Redis state.

----

Review your previous recommendation for the dev-only OTP plaintext debug capture.

I think the proposed service-level OTPDebugCodeSink may be too large for the next slice because it touches:
- internal/otp/interfaces.go
- internal/otp/service.go
- internal/otp/service_test.go
- internal/repository
- cmd/server/main.go

Please re-analyze and compare two options:

Option A:
Service-level OTPDebugCodeSink port.

Option B:
Fake SMS provider captures the plaintext OTP code into a dev-only Redis debug key, because SMSRequest.Code already reaches the fake provider.

Context:
- This is only for local manual testing.
- We do not want to expose the OTP code in API responses.
- We do not want to store it in otp_requests.
- We do not want to log it.
- We want a very small, reversible implementation.
- Fake SMS provider is already clearly a simulation component.
- Production real providers would not include this behavior.

Please analyze:
1. Which option produces the smallest safe diff?
2. Which option better preserves otp.Service cleanliness?
3. Which option is less risky for production?
4. Which files would change for each option?
5. How to enable it safely only in local/dev?
6. Whether Redis should be injected into fake provider or whether another sink abstraction is still worth it.
7. Whether this should be implemented now or deferred.
8. If implementing now, propose the smallest safe implementation plan.

Important:
- Do not modify files yet.
- Do not implement anything.
- Prefer local/dev-only safety and small diff.
- Do not add debug HTTP endpoints.
- Do not expose OTP code in SendOTP response.
- Do not store plaintext OTP in otp_requests or main OTP Redis state.

-----

Implement dev-only OTP code capture inside the fake SMS provider.

We are implementing incrementally to avoid large diffs and context/usage limits.

Scope:
- Modify only:
  - internal/sms/fake_provider.go
  - internal/sms/fake_provider_test.go
  - cmd/server/main.go
- Do not modify otp.Service.
- Do not modify internal/otp interfaces.
- Do not modify repository package.
- Do not modify API handlers/routes.
- Do not store plaintext OTP in otp_requests.
- Do not return plaintext OTP in API responses.
- Do not log plaintext OTP.
- Do not add a debug HTTP endpoint.
- Do not add generic options framework unless absolutely necessary.
- Keep the diff small and reversible.

Goal:
Allow local/development manual VerifyOTP testing by optionally storing plaintext OTP code in a separate Redis debug key from the fake SMS provider.

Design:
- Keep sms.NewFakeProvider() behavior unchanged and secure by default.
- Add a new explicit constructor:
  func NewFakeProviderWithDebugCodeCapture(client *redis.Client, ttl time.Duration) *FakeProvider
- Debug capture is enabled only when this constructor is used and client is non-nil.
- Store debug code only after simulated SMS send succeeds.
- Do not write debug code if context is canceled or timed out.

Redis debug key:
- Format:
  debug:otp-code:{tenant_id}:{phone}
- Value should be JSON with:
  - request_id
  - tenant_id
  - phone
  - code
  - provider
  - created_at
- TTL should be the provided ttl.
- If ttl <= 0, use a safe short default such as 60 seconds.
- Do not store code in the normal OTP state key.
- Do not include code in SMSResult.RawResponse.

Failure behavior:
- If debug Redis write fails, do not fail SendOTP.
- Debug capture is best-effort local tooling.

cmd/server/main.go wiring:
- Default must remain sms.NewFakeProvider().
- Enable debug capture only when:
  - environment variable OTP_FAKE_SMS_DEBUG_CODE_REDIS is true
  - and Gin mode is not release
- Use existing Redis client.
- Use a short TTL, preferably min(60s, otpConfig.TTL) or just 60s if simpler.
- Do not add full config/env struct fields in this step.

Tests:
- Existing fake provider tests must keep passing.
- Add tests:
  1. Default fake provider does not expose code in RawResponse.
  2. Debug constructor stores code in Redis under debug:otp-code:{tenant_id}:{phone}.
  3. Stored JSON contains request_id, tenant_id, phone, code, provider, created_at.
  4. Debug key has TTL.
  5. Context cancellation/timeout does not write debug key.
- Use existing Redis integration test style if Redis is needed.
- Skip Redis-dependent tests cleanly when Redis is unavailable.
- Do not add new dependencies.

Before modifying files:
- Briefly state the exact files you will change and why.

After implementation:
- run gofmt
- run go test -count=1 ./internal/sms -v
- run go test -count=1 ./...
- summarize changed files and test results.

----------
OTP config/env wiring ❌
resend cooldown / send rate protection ❌
verification logging repository ❌
business metrics/tracing ❌
better migration command/Makefile support ❌
real SMS provider abstraction/router ❌
auth/token validation ❌
OpenAPI/docs ❌


----
Before implementing OTP-related env/config wiring, analyze the current config loading system and current OTP/fake SMS hardcoded defaults.

Do not modify files yet.

Goal:
Design a small, consistent implementation plan to move OTP and fake SMS settings into environment/config files.

Current state:
- otp.DefaultConfig() is used in cmd/server/main.go.
- Fake SMS provider has default delay 20ms to 30ms.
- Dev-only fake SMS OTP capture is controlled by env var OTP_FAKE_SMS_DEBUG_CODE_REDIS directly in main.go.
- Some values are currently hardcoded/defaulted in code.
- I want these settings to be configurable through env and documented in .env / env.example if those files exist.

Please analyze:
1. Current config package structure.
2. Current env loading style.
3. Current .env and env.example files, if present.
4. How duration values are currently parsed.
5. Where OTP config should be added.
6. How to map env values into otp.Config.
7. How to configure fake SMS delay range.
8. How to configure fake SMS debug capture safely.
9. Which values should remain defaults if env is missing.
10. Validation rules for invalid env values.
11. Exact files that should change.
12. Tests that should be added.
13. What should be deferred.

Proposed env variables:
- OTP_CODE_LENGTH
- OTP_TTL
- OTP_MAX_ATTEMPTS
- OTP_TENANT_CACHE_TTL
- OTP_PROVIDER_TIMEOUT
- OTP_FAKE_SMS_MIN_DELAY
- OTP_FAKE_SMS_MAX_DELAY
- OTP_FAKE_SMS_DEBUG_CODE_REDIS
- OTP_FAKE_SMS_DEBUG_CODE_TTL

Important constraints:
- Do not implement anything yet.
- Do not modify files.
- Keep the next implementation diff small.
- Preserve current defaults when env vars are not set.
- Do not break existing config tests.
- Do not expose OTP codes in API responses or logs.
- Do not enable debug capture by default.
- Keep debug capture disabled unless explicitly enabled.
- Prefer existing config parsing patterns over new abstractions.
- Update env.example if it exists.
- Update .env only if the project already tracks and uses it for local development.
- Do not add new dependencies.

Return:
1. Recommended config design.
2. Env variable names and default values.
3. Exact files that should change.
4. Validation/error behavior.
5. Required tests.
6. Deferred concerns.
```
--------------

Implement OTP and fake SMS env/config wiring.

We are implementing incrementally to avoid large diffs and context/usage limits.

Scope:
- Modify only:
  - internal/config/config.go
  - internal/config/config_test.go
  - internal/sms/fake_provider.go
  - internal/sms/fake_provider_test.go
  - cmd/server/main.go
  - env.example
- Do not modify otp.Service business logic.
- Do not modify repository code.
- Do not modify API handlers/routes.
- Do not modify migrations.
- Do not add new dependencies.
- Do not update .env.
- Keep the diff small and easy to review.

Goal:
Move OTP-related runtime values from hardcoded/default-only usage into the existing config/env system while preserving current defaults.

Config requirements:
1. Add OTPConfig to internal/config:
   - CodeLength int
   - TTL time.Duration
   - MaxAttempts int
   - TenantCacheTTL time.Duration
   - ProviderTimeout time.Duration
   - FakeSMSMinDelay time.Duration
   - FakeSMSMaxDelay time.Duration
   - FakeSMSDebugCodeRedis bool
   - FakeSMSDebugCodeTTL time.Duration

2. Add OTP OTPConfig to root Config.

3. Load env vars:
   - OTP_CODE_LENGTH default 6
   - OTP_TTL default 2m
   - OTP_MAX_ATTEMPTS default 3
   - OTP_TENANT_CACHE_TTL default 5m
   - OTP_PROVIDER_TIMEOUT default 2s
   - OTP_FAKE_SMS_MIN_DELAY default 20ms
   - OTP_FAKE_SMS_MAX_DELAY default 30ms
   - OTP_FAKE_SMS_DEBUG_CODE_REDIS default false
   - OTP_FAKE_SMS_DEBUG_CODE_TTL default 60s

4. Validation:
   - OTP_CODE_LENGTH must be between 1 and 18
   - OTP_TTL must be > 0
   - OTP_MAX_ATTEMPTS must be > 0
   - OTP_TENANT_CACHE_TTL must be > 0
   - OTP_PROVIDER_TIMEOUT must be > 0
   - OTP_FAKE_SMS_MIN_DELAY must be >= 0
   - OTP_FAKE_SMS_MAX_DELAY must be >= 0 and >= min delay
   - OTP_FAKE_SMS_DEBUG_CODE_TTL must be > 0

5. Bool parsing:
   - Keep existing style if there is one.
   - Accept true and 1 as true.
   - Missing/other values should be false unless existing config behavior says otherwise.

Fake SMS requirements:
- Keep sms.NewFakeProvider() unchanged with current default 20ms to 30ms behavior.
- Add a simple constructor for configurable delay if needed, for example:
  NewFakeProviderWithDelay(minDelay, maxDelay time.Duration)
- Do not introduce an options framework.
- Existing tests must keep passing.

main.go requirements:
- Use cfg.OTP to build otp.Config instead of raw otp.DefaultConfig().
- Use configured fake SMS delay.
- Keep debug capture disabled unless:
  - cfg.OTP.FakeSMSDebugCodeRedis is true
  - Gin mode is not release
- Debug code capture TTL should be min(cfg.OTP.FakeSMSDebugCodeTTL, cfg.OTP.TTL).
- Do not read OTP_FAKE_SMS_DEBUG_CODE_REDIS directly in main.go anymore.

env.example:
- Add all OTP env vars with defaults and short comments if the file style supports comments.

Tests:
- Add config tests for:
  - defaults
  - env overrides
  - invalid duration
  - invalid code length
  - max attempts <= 0
  - fake SMS max delay < min delay
  - debug flag true and 1
  - debug flag default false
- Add/adjust SMS tests if configurable delay constructor is added.

Before modifying files:
- Briefly state exact files you will change and why.

After implementation:
- run gofmt
- run go test -count=1 ./internal/config -v
- run go test -count=1 ./internal/sms -v
- run go test -count=1 ./...
- summarize changed files and test results.

----------

Analyze the next incremental step for OTP resend protection / active OTP prevention.

Current state:
- OTP state is stored in Redis via OTPStore.
- Key format:
  otp:{tenant_id}:{phone}
- SendOTP currently always creates a new OTP state and sends SMS.
- VerifyOTP deletes OTP state on successful verification.
- OTP TTL already exists.
- No resend protection or cooldown exists yet.
- Manual testing showed the same phone can request unlimited OTPs rapidly.

Goal:
Design the smallest safe implementation to prevent OTP spam and repeated resend while an active OTP already exists.

Important:
- Do not implement yet.
- Do not modify files yet.
- Keep the next diff small.
- Prefer reusing existing Redis OTP state.
- Avoid introducing a full generic rate limiter.
- Avoid introducing new infrastructure or dependencies.
- Avoid adding background jobs.
- Preserve current VerifyOTP behavior.

Please analyze:
1. Recommended resend protection strategy.
2. Whether SendOTP should reject when an active OTP already exists.
3. Whether cooldown should be based on Redis key existence or timestamps.
4. Recommended API response behavior and HTTP status.
5. Whether remaining TTL should be exposed to the client.
6. Race conditions and Redis consistency concerns.
7. Exact files that should change.
8. Whether OTPStore interface changes are needed.
9. Whether OTPStore.Get should be reused or a cheaper Exists API is better.
10. Required tests.
11. Edge cases:
   - expired OTP
   - malformed Redis state
   - concurrent sends
   - Redis failures
12. Whether request logging should log blocked resend attempts.
13. Deferred concerns:
   - distributed rate limiting
   - IP throttling
   - tenant quotas
   - resend-after durations
   - resend token flows
   - resend same-code behavior
   - background cleanup
   - provider billing protection

Preferred direction:
- Keep implementation small.
- Reuse existing Redis OTP state.
- Reject new sends while an active OTP exists.
- Use 429 Too Many Requests at the HTTP layer.

------

Implement the first OTP resend protection using the existing Redis OTP state.

We are implementing incrementally to avoid large diffs and context/usage limits.

Scope:
- Modify only:
  - internal/otp/errors.go
  - internal/otp/service.go
  - internal/otp/service_test.go
  - internal/api/otp_flow_handlers.go
  - internal/api/otp_flow_handlers_test.go
- Do not modify repository code.
- Do not modify OTPStore interface.
- Do not modify Redis store implementation.
- Do not modify cmd/server/main.go.
- Do not modify config.
- Do not add new Redis keys.
- Do not add generic rate limiter.
- Do not add tenant quota system.
- Do not add Retry-After header yet.
- Do not log blocked resend attempts into otp_requests.
- Keep the diff small and easy to review.

Goal:
Prevent repeated /otp/send requests for the same tenant_id + phone while an active, unexpired OTP already exists.

Requirements:

1. Add domain error:
   ErrOTPAlreadyActive

2. SendOTP flow:
   - Validate request.
   - Load tenant settings.
   - Validate tenant.
   - Check existing OTP state using OTPStore.Get(ctx, tenantID, phone).
   - If OTPStore.Get returns ErrOTPNotFound:
     continue normal send flow.
   - If OTPStore.Get returns any other error:
     abort and return wrapped error.
   - If an existing OTP state is found and ExpiresAt is in the future:
     return ErrOTPAlreadyActive.
   - If an existing OTP state is expired:
     best-effort Delete(ctx, tenantID, phone)
     continue normal send flow.
   - Do not create request log for blocked resend attempts.
   - Do not generate a new OTP for blocked resend attempts.
   - Do not call SMS provider for blocked resend attempts.

3. Expired existing OTP behavior:
   - If Delete fails, ignore the delete error and continue.
   - This matches best-effort cleanup behavior used elsewhere.

4. API mapping:
   - Map ErrOTPAlreadyActive to HTTP 429 Too Many Requests.
   - Use existing error response style.
   - Do not add Retry-After yet.

5. Tests in internal/otp:
   Add/update service tests for:
   - no existing OTP / ErrOTPNotFound: SendOTP proceeds normally
   - existing active OTP: returns ErrOTPAlreadyActive
   - existing active OTP: does not create request log
   - existing active OTP: does not save new OTP state
   - existing active OTP: does not call SMS provider
   - existing expired OTP: best-effort delete and SendOTP proceeds
   - existing expired OTP delete failure: SendOTP still proceeds
   - OTPStore.Get returns non-not-found error: SendOTP aborts
   - tenant validation still happens before active OTP check

6. Tests in internal/api:
   - Send handler maps ErrOTPAlreadyActive to 429.

Important:
- Before modifying files, briefly state exact files you will change and why.
- After implementation:
  - run gofmt
  - run go test -count=1 ./internal/otp -v
  - run go test -count=1 ./internal/api -v
  - run go test -count=1 ./...
  - summarize changed files and test results.


---------
send
verify
Redis state
PostgreSQL request logging
HTTP APIs
fake SMS
debug OTP retrieval
resend protection
expiration
max attempts
cleanup behavior
config/env wiring
handler tests
service tests
repository tests

مرحله‌های مهم بعدی احتمالاً یکی از این‌ها هستند:


verification logging (otp_verifications)
- Verification logging repository + migration
- Wire verification logger into VerifyOTP + main.go
structured rate limiting (per phone/IP/tenant)
tracing/metrics
atomic resend protection (SET NX / Lua)
real SMS provider abstraction
OpenAPI / Swagger
auth/token validation
production hardening
integration/e2e docker tests
admin/reporting APIs

-----------------

Before implementing OTP verification logging, analyze the current VerifyOTP flow, existing OTP request logging repository, and database migration style.

Do not modify files yet.

Goal:
Design a small implementation for recording OTP verification attempts/results.

Current state:
- SendOTP logs request/provider result in otp_requests.
- VerifyOTP is implemented but does not log verification attempts.
- OTPVerificationLogger interface already exists.
- OTPVerificationLog model likely exists in internal/otp.
- No otp_verifications table/repository exists yet.

Please analyze:
1. Current OTPVerificationLogger interface.
2. Current OTPVerificationLog model fields.
3. Current VerifyOTP flow and where logging should happen.
4. Whether verification logging should be mandatory or best-effort.
5. What database table/schema is needed.
6. Whether failed business verifications should be logged.
7. Whether invalid request validation failures should be logged.
8. How to handle missing OTP / expired / invalid_code / max_attempts / success.
9. Whether request_id should be stored when available.
10. Whether tenant_id, phone, result, reason, attempt_count, metadata, created_at should be stored.
11. Exact files that should change.
12. Recommended repository tests.
13. Recommended service tests.
14. What should be deferred.

Important constraints:
- Do not implement anything yet.
- Do not modify files.
- Keep the next implementation diff small.
- Do not modify HTTP handlers/routes unless absolutely necessary.
- Do not add metrics/tracing yet.
- Do not add async logging/outbox.
- Do not add rate limiting yet.
- Follow current repository and migration style.
- Prefer database/sql and existing sql-migrate format.

Return:
1. Recommended logging design.
2. Exact files that should change.
3. Proposed migration/table schema.
4. Repository implementation approach.
5. VerifyOTP wiring approach.
6. Failure policy.
7. Tests to add.
8. Deferred concerns.

---------------
we broke it to 2 separate phases:
Verification logging repository + migration
Wire verification logger into VerifyOTP + main.go
-------------
Implement OTP verification logging repository and migration only.

We are implementing incrementally to avoid large diffs and context/usage limits.

Scope:
- Add only:
  - one sql-migrate migration for otp_verifications
  - internal/repository/otp_verification_log_repo.go
  - internal/repository/otp_verification_log_repo_test.go
- Do not modify internal/otp/service.go yet.
- Do not modify internal/otp/service_test.go yet.
- Do not modify cmd/server/main.go yet.
- Do not modify API handlers/routes.
- Do not add metrics/tracing.
- Do not add async logging/outbox.
- Do not add rate limiting.
- Keep the diff small and easy to review.

Migration requirements:
- First inspect existing migration files and follow the exact naming/style used in this project.
- Use sql-migrate format with Up/Down if that is the existing style.
- Create table otp_verifications if it does not exist.
- Columns:
  - id BIGSERIAL PRIMARY KEY
  - request_id TEXT NOT NULL DEFAULT ''
  - tenant_id BIGINT NOT NULL
  - phone TEXT NOT NULL
  - result TEXT NOT NULL
  - reason TEXT NOT NULL DEFAULT ''
  - attempt_count INTEGER NOT NULL DEFAULT 0
  - correlation_id TEXT
  - created_at TIMESTAMPTZ NOT NULL DEFAULT now()
- Indexes:
  - request_id
  - tenant_id, created_at DESC
  - phone, created_at DESC
  - result, created_at DESC
- Do not add a foreign key to otp_requests yet.
- Add Down migration that drops indexes/table consistently with project style.

Repository requirements:
- Create OTPVerificationLogRepository using database/sql.
- Constructor:
  func NewOTPVerificationLogRepository(db *sql.DB) *OTPVerificationLogRepository
- Implement existing otp.OTPVerificationLogger interface:
  func (r *OTPVerificationLogRepository) LogVerification(ctx context.Context, log otp.OTPVerificationLog) error
- Use parameterized SQL.
- If CreatedAt is zero, default to time.Now().UTC().
- Store empty CorrelationID as SQL NULL if existing repository helpers support that pattern; otherwise keep it simple and consistent with existing code.
- Wrap errors with clear context.

Testing requirements:
- Follow current repository integration test style.
- Tests may use real PostgreSQL and skip cleanly if unavailable.
- Do not add new dependencies.
- Do not create schema in Go tests unless existing repository tests already do that.
- If otp_verifications table is missing because migration is not applied, skip cleanly with a clear message.

Add tests:
1. LogVerification inserts a success row.
2. LogVerification inserts a failed row with reason and attempt_count.
3. Zero CreatedAt is defaulted and stored.
4. Empty CorrelationID is stored as NULL if practical to assert.
5. Optional only if small: non-empty CorrelationID is stored.

Before modifying files:
- Briefly state the exact files you will create and why.

After implementation:
- run gofmt
- run go test -count=1 ./internal/repository -v
- run go test -count=1 ./...
- summarize changed files and test results.
---------

Before wiring OTP verification logging into VerifyOTP, analyze the current VerifyOTP service flow, service tests, and server wiring.

Do not modify files yet.

Current state:
- OTPVerificationLogger interface already exists.
- OTPVerificationLog model already exists.
- PostgreSQL OTPVerificationLogRepository and migration are implemented and committed.
- VerifyOTP is implemented and tested.
- cmd/server/main.go currently passes nil as verifyLogger to otp.NewService.
- Verification logging should now be wired as best-effort.

Goal:
Design a small, safe implementation plan to:
1. log verification outcomes from otp.Service.VerifyOTP
2. construct OTPVerificationLogRepository in cmd/server/main.go
3. pass it to otp.NewService

Please analyze:
1. Current VerifyOTP branches and all business outcomes.
2. Where LogVerification should be called in each branch.
3. Which outcomes should be logged:
   - success
   - not_found
   - expired
   - invalid_code
   - max_attempts_exceeded
4. Which outcomes should NOT be logged:
   - invalid request
   - infrastructure/store errors
   - delete failure after success
   - increment infrastructure errors
5. What result/reason values should be used.
6. How attempt_count should be populated for each outcome.
7. How request_id should be populated when state is available.
8. Whether logging should be best-effort and ignored on failure.
9. How to ensure logging does not change VerifyOTP responses.
10. Exact files that should change.
11. Required service tests.
12. Whether cmd/server/main.go should be included in the same slice.
13. Manual verification steps after implementation.

Important constraints:
- Do not implement anything yet.
- Do not modify files.
- Keep the next diff small.
- Modify only:
  - internal/otp/service.go
  - internal/otp/service_test.go
  - cmd/server/main.go
  unless a compile-time mismatch requires otherwise.
- Do not modify repository code.
- Do not modify migrations.
- Do not modify API handlers/routes.
- Do not add metrics/tracing.
- Do not add async logging/outbox.
- Verification logging must be best-effort.
- Logging failure must not change VerifyOTP response.

Return:
1. Recommended wiring approach.
2. Exact VerifyOTP logging points.
3. Result/reason mapping table.
4. Failure policy.
5. Exact files that should change.
6. Tests to add/update.
7. Manual validation steps.

---------
I have implemented these sections:
SendOTP
VerifyOTP
Redis OTP state
request logging
verification logging
resend protection
fake SMS provider
debug OTP capture
env-driven config
HTTP handlers/routes
end-to-end manual testability
repository integration tests
service-level tests
handler-level tests
-------------------
next steps:
structured rate limiting (per phone, per tenant, per IP)
tracing/metrics
atomic Redis flow
------------
We have completed the core OTP flow:
- SendOTP
- VerifyOTP
- Redis OTP state
- resend protection
- request logging
- verification logging
- HTTP handlers
- env-driven config
- fake SMS provider
- debug OTP capture

Now I want to implement the first structured rate limiting layer for OTP sends.

Please analyze and design the best small incremental implementation for this codebase.

Goals:
- prevent OTP abuse/spam
- keep implementation small and reviewable
- avoid large architecture changes
- reuse Redis when possible
- preserve current OTP resend protection behavior

I want you to analyze:
1. recommended rate limiting strategy
2. per-phone vs per-tenant vs per-IP ordering
3. Redis key design
4. fixed window vs sliding window vs token bucket
5. where the logic should live
6. whether this should be inside otp.Service or middleware
7. exact files that should change
8. config/env additions needed
9. required tests
10. HTTP/API behavior and status codes
11. race conditions and edge cases
12. interaction with existing resend protection
13. whether resend protection and rate limiting should stay separate or be merged later
14. what should explicitly be deferred

Constraints:
- keep architecture clean
- keep diff small
- do not redesign the whole app
- do not add external rate limiting libraries yet
- prefer Redis-based implementation
- keep current resend protection behavior unchanged for now
- preserve current test style
- avoid generic middleware abstractions unless clearly justified

Please provide:
- recommended design
- exact implementation plan
- exact files to change
- test plan
- deferred concerns
- rollout order
- risks/tradeoffs
----------------

Phase 1: Domain/service/API mapping با fake limiter
Phase 2: Redis limiter repository
Phase 3: config/env + main.go wiring
---------
Implement Phase 1 of OTP send rate limiting: domain/service/API integration only.

We are implementing incrementally to avoid large diffs and context/usage limits.

Scope:
- Modify only:
  - internal/otp/errors.go
  - internal/otp/interfaces.go
  - internal/otp/service.go
  - internal/otp/service_test.go
  - internal/api/otp_flow_handlers.go
  - internal/api/otp_flow_handlers_test.go
- Do not modify repository code.
- Do not add Redis rate limiter implementation yet.
- Do not modify config/env yet.
- Do not modify cmd/server/main.go yet.
- Do not modify env.example yet.
- Do not add metrics/tracing.
- Do not add generic middleware.
- Keep the diff small and easy to review.

Goal:
Prepare SendOTP to support an optional send rate limiter dependency, while preserving existing behavior when no limiter is configured.

Requirements:

1. Add domain error:
   ErrOTPRateLimited

2. Add interface in internal/otp/interfaces.go:
   type SendRateLimiter interface {
       AllowSend(ctx context.Context, tenantID int64, phone string) error
   }

3. Add optional limiter dependency to otp.Service.
   Keep existing call sites compiling.
   Prefer a small setter method such as:
   func (s *Service) SetSendRateLimiter(limiter SendRateLimiter)
   rather than changing the NewService constructor signature.

4. SendOTP flow:
   - request validation
   - tenant lookup
   - tenant validation
   - existing active OTP resend protection
   - then rate limiter check
   - then existing request generation/logging/save/SMS flow

5. Behavior:
   - If no limiter is configured, SendOTP behavior remains unchanged.
   - If active OTP already exists, return ErrOTPAlreadyActive and do not call limiter.
   - If limiter returns ErrOTPRateLimited, abort before request log/save/SMS.
   - If limiter returns any other error, abort before request log/save/SMS and wrap the error.
   - Tenant validation must happen before limiter.
   - Blocked rate-limited sends should not be logged in otp_requests yet.

6. API mapping:
   - Map ErrOTPRateLimited to HTTP 429 Too Many Requests.
   - Keep existing ErrOTPAlreadyActive mapping unchanged.
   - Do not add Retry-After yet.

7. Tests:
   Update service tests with fake limiter:
   - limiter nil: existing SendOTP success still works.
   - limiter allows send: SendOTP proceeds.
   - limiter returns ErrOTPRateLimited: SendOTP returns ErrOTPRateLimited.
   - limiter blocks: no request log, no store save, no SMS.
   - limiter returns infrastructure error: SendOTP returns error.
   - active OTP happens before limiter: limiter is not called.
   - tenant disabled happens before limiter: limiter is not called.

   Update API tests:
   - ErrOTPRateLimited maps to 429.

Before modifying files:
- Briefly state exact files you will change and why.

After implementation:
- run gofmt
- run go test -count=1 ./internal/otp -v
- run go test -count=1 ./internal/api -v
- run go test -count=1 ./...
- summarize changed files and test results.
-----------------
Analyze Phase 2 of OTP send rate limiting: Redis-backed implementation.

Current state:
- Phase 1 is done and committed.
- internal/otp has SendRateLimiter interface:
  AllowSend(ctx context.Context, tenantID int64, phone string) error
- otp.Service.SendOTP calls the limiter after active OTP resend protection and before request logging/save/SMS.
- ErrOTPRateLimited exists and maps to HTTP 429.
- No Redis limiter implementation exists yet.
- No config/main wiring exists yet for the limiter.

Goal:
Design a small Redis-backed fixed-window implementation of SendRateLimiter.

Do not modify files yet.

Please analyze:
1. Recommended Redis fixed-window strategy.
2. Redis key format.
3. Whether to use INCR+EXPIRE pipeline or Lua for atomicity.
4. How to avoid keys without TTL.
5. What constructor/config fields the repository should accept.
6. Whether this implementation should live in internal/repository.
7. How to map limit exceeded to otp.ErrOTPRateLimited.
8. How to handle Redis infrastructure errors.
9. Whether remaining quota or retry-after should be returned now or deferred.
10. Required tests.
11. Edge cases:
    - limit <= 0
    - window <= 0
    - Redis unavailable
    - TTL missing
    - different tenant/phone isolation
    - concurrent calls
12. What should be deferred to Phase 3 config/main wiring.

Constraints:
- Do not implement yet.
- Do not modify otp.Service.
- Do not modify config/env yet.
- Do not modify cmd/server/main.go yet.
- Do not modify API handlers.
- Keep diff small.
- Use existing Redis client style.
- Follow existing repository test style.
- Do not add new dependencies.

Return:
1. Recommended implementation design.
2. Exact files to change.
3. Redis command/script design.
4. Error behavior.
5. Tests to add.
6. Deferred concerns.

---------

Implement Phase 2 of OTP send rate limiting: Redis-backed fixed-window limiter.

We are implementing incrementally to avoid large diffs and context/usage limits.

Scope:
- Add only:
  - internal/repository/otp_send_rate_limiter_redis.go
  - internal/repository/otp_send_rate_limiter_redis_test.go
- Do not modify internal/otp.
- Do not modify otp.Service.
- Do not modify API handlers/routes.
- Do not modify config/env.
- Do not modify cmd/server/main.go.
- Do not modify env.example.
- Do not add new dependencies.
- Keep the diff small and easy to review.

Goal:
Add a Redis-backed implementation of otp.SendRateLimiter.

Requirements:

1. Add type:
   RedisOTPSendRateLimiter

2. Add constructor:
   func NewRedisOTPSendRateLimiter(client *redis.Client, limit int, window time.Duration) *RedisOTPSendRateLimiter

3. Implement:
   func (l *RedisOTPSendRateLimiter) AllowSend(ctx context.Context, tenantID int64, phone string) error

4. Redis key format:
   otp:rate:send:{tenant_id}:{phone}

5. Use Redis Lua script, not plain INCR + EXPIRE pipeline.

Use guarded Lua behavior:
- INCR key
- if current count == 1 OR key has no TTL, set PEXPIRE to window
- return current count

Conceptually:

  local current = redis.call("INCR", KEYS[1])
  if current == 1 or redis.call("PTTL", KEYS[1]) < 0 then
    redis.call("PEXPIRE", KEYS[1], ARGV[1])
  end
  return current

6. Error behavior:
- if client is nil, return clear configuration/infrastructure error
- if limit <= 0, return clear configuration error
- if window <= 0, return clear configuration error
- if Redis/Lua fails, return wrapped infrastructure error
- if current count > limit, return otp.ErrOTPRateLimited
- do not log inside adapter
- do not return remaining quota/retry-after yet

7. Tests:
Follow existing Redis integration test style and skip cleanly when Redis is unavailable.

Add tests:
- allows requests under limit
- blocks after limit and errors.Is(err, otp.ErrOTPRateLimited)
- sets TTL after first allowed call
- isolates different tenant IDs
- isolates different phones
- invalid limit returns non-rate-limit error
- invalid window returns non-rate-limit error
- nil Redis client returns non-rate-limit error
- repairs missing TTL:
  - manually create key without TTL
  - call AllowSend
  - assert TTL is now positive
- optional if still small:
  - concurrent calls with limit N result in exactly N allowed and remaining calls rate-limited

Important:
- Do not change service behavior in this phase.
- Do not wire this limiter in main.go yet.
- Before modifying files, briefly state exact files you will create and why.
- After implementation:
  - run gofmt
  - run go test -count=1 ./internal/repository -v
  - run go test -count=1 ./...
  - summarize changed files and test results.

  ----------

  