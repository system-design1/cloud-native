# Tenant Settings – Insert Benchmark (API + Repo)

هدف: یک endpoint بسیار ساده برای **benchmark insert در Postgres** (end-to-end از طریق API) بسازیم.

فرض‌ها:
- جدول `tenant_settings_for_insert_new` در Postgres وجود دارد (مشابه `tenant_settings`).
- برای جلوگیری از overhead، داده‌ها **server-side رندم تولید نمی‌شوند**.
- برای یکتا بودن `tenant_code`، از `X-Correlation-ID` (که همین الان middleware تولید/ست می‌کند) استفاده می‌کنیم.

---

## Prompt 01 — Create repository for INSERT

TASK: Create a new repository file for inserting into Postgres table.
INPUT:
- File path: internal/repository/tenant_settings_insert_repo.go
- Table: tenant_settings_for_insert_new
- Must use database/sql with *sql.DB injected (no globals)
- Method: InsertTenantSettingsForInsertNew(ctx context.Context, tenantCode string) (int64, error)
- SQL: INSERT INTO tenant_settings_for_insert_new (tenant_code, name, status, otp_enabled, sms_provider, sms_api_key, rate_limit_per_min, signup_at, expires_at, timezone, metadata, created_at, updated_at, deleted_at)
  VALUES ($1, 'Benchmark Tenant', 'active', true, 'other', NULL, 60, now(), NULL, 'UTC', '{}'::jsonb, now(), now(), NULL)
  RETURNING id;
- On error return fmt.Errorf("...: %w", err)
OUTPUT: Exactly one Go file: internal/repository/tenant_settings_insert_repo.go
RULES:
- No explanation
- No extra text
- Output only the result

---

## Prompt 02 — Create API handler for INSERT

TASK: Create a new Gin handler file that inserts one row into tenant_settings_for_insert_new.
INPUT:
- File path: internal/api/tenant_settings_insert_handlers.go
- New handler: InsertTenantSettingsBenchmarkHandler(repo *repository.TenantSettingsInsertRepository) gin.HandlerFunc
- Route behavior:
  - Method: POST
  - Path: /v1/otp/tenant-settings-insert-benchmark
  - Tenant code source:
    - Prefer correlation_id from Gin context key "correlation_id" (set by CorrelationIDMiddleware)
    - If missing, fallback to header "X-Correlation-ID"
    - If still empty, return 500 via middleware.ErrorHandler
  - Call repo.InsertTenantSettingsForInsertNew(ctx, tenantCode)
  - On success: return 200 JSON: {"id": <returned_id>}
- Must use existing project patterns:
  - import internal/middleware for ErrorHandler
  - import pkg/errors as apperrors for typed errors (ErrInternalServerError)
OUTPUT: Exactly one Go file: internal/api/tenant_settings_insert_benchmark_handlers.go
RULES:
- No explanation
- No analysis
- No extra text
- Output only the result

---

## Prompt 03 — Wire route + update SetupRoutes signature

TASK: Wire the new insert endpoint in routes and update SetupRoutes signature.
INPUT:
- File: internal/api/routes.go
- Change function signature to:
  SetupRoutes(router *gin.Engine, lifecycleMgr *lifecycle.Manager, tenantSettingsRepo *repository.TenantSettingsRepository, tenantSettingsInsertRepo *repository.TenantSettingsInsertRepository)
- Add route under /v1/otp:
  POST /tenant-settings-insert-benchmark -> InsertTenantSettingsBenchmarkHandler(tenantSettingsInsertRepo)
- Keep all existing routes unchanged.
OUTPUT: Modify exactly one file: internal/api/routes.go
RULES:
- No explanation
- No extra text
- Output only the result

---

## Prompt 04 — Wire repository in main.go

TASK: Initialize the new insert repository and pass it to SetupRoutes.
INPUT:
- File: cmd/server/main.go
- After tenantSettingsRepo initialization, create:
  tenantSettingsInsertRepo := repository.NewTenantSettingsInsertRepository(database)
- Update SetupRoutes call to pass the new repo:
  api.SetupRoutes(router, lifecycleMgr, tenantSettingsRepo, tenantSettingsInsertRepo)
- Keep everything else unchanged.
OUTPUT: Modify exactly one file: cmd/server/main.go
RULES:
- No explanation
- No extra text
- Output only the result

---

## Quick manual run (خارج از Cursor)

1) سرویس را بالا بیاور:
- make docker-up  (یا روش معمول خودت)

2) تست دستی:
- curl -X POST http://localhost:8080/v1/otp/tenant-settings-insert

3) برای load test:
- در k6 می‌توانی check را روی status==200 بگذاری
- tenant_code به صورت خودکار از X-Correlation-ID یکتا می‌شود
