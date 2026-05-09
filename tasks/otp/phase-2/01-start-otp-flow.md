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
