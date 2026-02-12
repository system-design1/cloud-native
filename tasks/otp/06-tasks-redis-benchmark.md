# Redis Benchmark (GET/SET) — Tasks for Cursor

هدف:
- اضافه کردن Redis به پروژه (docker compose + config).
- ساخت یک Redis client با connection pool (از ENV).
- ساخت دو endpoint ساده برای benchmark:
  - `POST /v1/redis/set`  → SET یک key/value ثابت با TTL
  - `GET  /v1/redis/get`  → GET همان key
- کمترین overhead برای load test.

---

## Prompt 01 — Add Redis service to docker-compose

TASK: Add a Redis service for local development and connect it to app-network.
INPUT:
- File: docker-compose.yml
- Add service `redis`:
  - image: redis:7.4-alpine
  - container_name: go-backend-redis
  - ports: "6379:6379"
  - command: ["redis-server", "--save", "", "--appendonly", "no"]
  - healthcheck: ["CMD", "redis-cli", "ping"] with interval 10s, timeout 5s, retries 5
  - networks: app-network
- Ensure `api` depends_on includes redis with condition service_healthy
- Do NOT remove existing services or networks
OUTPUT: Modify exactly one file: docker-compose.yml
RULES:
- No explanation
- No analysis
- No extra text
- Output only the result

---

## Prompt 02 — Add Redis ENV vars to .env and env.example

TASK: Add Redis env vars to .env and env.example.
INPUT:
- Files: .env and env.example
- Add these lines near other service configs:
  - REDIS_HOST=redis
  - REDIS_PORT=6379
  - REDIS_PASSWORD=
  - REDIS_DB=0
  - REDIS_POOL_SIZE=50
  - REDIS_MIN_IDLE_CONNS=10
  - REDIS_DIAL_TIMEOUT=2s
  - REDIS_READ_TIMEOUT=2s
  - REDIS_WRITE_TIMEOUT=2s
- Keep existing content unchanged
OUTPUT:
- Modify: .env
- Modify: env.example
RULES:
- No explanation
- No analysis
- No extra text
- Output only the result

---

## Prompt 03 — Add Redis config loader

TASK: Extend config to load Redis settings from ENV with defaults.
INPUT:
- File: internal/config/config.go
- Add `Redis RedisConfig` to `type Config`
- Add new struct:
  - type RedisConfig struct {
      Host string
      Port int
      Password string
      DB int
      PoolSize int
      MinIdleConns int
      DialTimeout time.Duration
      ReadTimeout time.Duration
      WriteTimeout time.Duration
    }
- Add loader function `loadRedisConfig(cfg *Config) error` and call it from `Load()` after database config.
- ENV + defaults:
  - REDIS_HOST default "127.0.0.1"
  - REDIS_PORT default 6379
  - REDIS_PASSWORD default ""
  - REDIS_DB default 0
  - REDIS_POOL_SIZE default 50
  - REDIS_MIN_IDLE_CONNS default 10
  - REDIS_DIAL_TIMEOUT default 2s
  - REDIS_READ_TIMEOUT default 2s
  - REDIS_WRITE_TIMEOUT default 2s
- Validate:
  - Port 1..65535
  - DB >= 0
  - PoolSize >= 1
  - MinIdleConns >= 0 and <= PoolSize
  - durations parse via time.ParseDuration
OUTPUT: Modify exactly one file: internal/config/config.go
RULES:
- No explanation
- No analysis
- No extra text
- Output only the result

---

## Prompt 04 — Create Redis client package (go-redis) with pool settings

TASK: Create a Redis client initializer using go-redis/v9 and config.Redis.
INPUT:
- File path: internal/redis/redis.go
- Use package name: redis
- Use imports:
  - "context"
  - "fmt"
  - "time"
  - cfg "go-backend-service/internal/config"
  - "github.com/redis/go-redis/v9"
- Provide function:
  - func NewClient(c cfg.RedisConfig) (*redis.Client, error)
    - Build addr host:port
    - Create redis.NewClient(&redis.Options{ ... pool settings from config ... })
    - Ping with 2s context timeout to verify connectivity
    - Return client or wrapped error
- Include English comments only.
OUTPUT: Exactly one Go file: internal/redis/redis.go
RULES:
- No explanation
- No analysis
- No extra text
- Output only the result

---

## Prompt 05 — Add repository wrapper for benchmark GET/SET

TASK: Create a small repository for Redis SET/GET operations for benchmarking.
INPUT:
- File path: internal/repository/redis_benchmark_repo.go
- Repository struct:
  - type RedisBenchmarkRepository struct { client *redis.Client }
  - constructor NewRedisBenchmarkRepository(client *redis.Client) *RedisBenchmarkRepository
- Methods:
  - SetBenchmarkKey(ctx context.Context, key string, value string, ttl time.Duration) error
  - GetBenchmarkKey(ctx context.Context, key string) (string, error)
- Use go-redis/v9.
- Wrap errors with fmt.Errorf("...: %w", err)
- English comments only.
OUTPUT: Exactly one Go file: internal/repository/redis_benchmark_repo.go
RULES:
- No explanation
- No analysis
- No extra text
- Output only the result

---

## Prompt 06 — Add API handlers for Redis benchmark

TASK: Create Gin handlers for Redis OTP SET/GET benchmark endpoints.
INPUT:
- File path: internal/api/redis_benchmark_handlers.go
- Handlers must use existing middleware.ErrorHandler for errors.
- Endpoints:
  - POST /v1/redis/otp/set
    - Inputs (from query string):
      - tenant_id (required)
      - phone_number (required)
    - otp_code must be a fixed string: "123456" (no random generation)
    - Key format: "otp:{tenant_id}:{phone_number}"
    - Value format: JSON string with fields:
      - tenant_id (string)
      - phone_number (string)
      - otp_code (string)
    - TTL: 120s
    - Return 200 JSON: {"ok": true}
  - GET /v1/redis/otp/get
    - Inputs (from query string):
      - tenant_id (required)
      - phone_number (required)
    - Key format: "otp:{tenant_id}:{phone_number}"
    - If key missing (redis.Nil): return 200 JSON: {"found": false}
    - If found: return 200 JSON:
      {
        "found": true,
        "tenant_id": "<tenant_id>",
        "phone_number": "<phone_number>",
        "otp_code": "<otp_code>"
      }
- Validation:
  - If tenant_id or phone_number missing: use middleware.ErrorHandler with apperrors.ErrBadRequest("...") and return.
- Treat redis.Nil as missing key.
- English comments only.
OUTPUT: Exactly one Go file: internal/api/redis_benchmark_handlers.go
RULES:
- No explanation
- No analysis
- No extra text
- Output only the result


---

## Prompt 07 — Wire routes

TASK: Register Redis benchmark routes under /v1/redis.
INPUT:
- File: internal/api/routes.go
- Add a new route group:
  - /v1/redis
    - POST /set -> RedisSetBenchmarkHandler(redisRepo)
    - GET /get  -> RedisGetBenchmarkHandler(redisRepo)
- Update SetupRoutes signature to accept:
  - redisBenchmarkRepo *repository.RedisBenchmarkRepository
- Keep existing routes and middleware order unchanged.
OUTPUT: Modify exactly one file: internal/api/routes.go
RULES:
- No explanation
- No analysis
- No extra text
- Output only the result

---

## Prompt 08 — Wire Redis in main.go

TASK: Initialize Redis client and repository, then pass into SetupRoutes.
INPUT:
- File: cmd/server/main.go
- After config load:
  - rdb, err := redis.NewClient(cfg.Redis)
  - handle error and exit as project style
  - defer rdb.Close()
- Create repository:
  - redisRepo := repository.NewRedisBenchmarkRepository(rdb)
- Update SetupRoutes call to pass redisRepo.
- Keep everything else unchanged.
OUTPUT: Modify exactly one file: cmd/server/main.go
RULES:
- No explanation
- No analysis
- No extra text
- Output only the result

---

## Prompt 09 — Update go.mod

TASK: Add go-redis/v9 dependency.
INPUT:
- File: go.mod
- Add: github.com/redis/go-redis/v9 (latest compatible)
- Run go mod tidy changes in file content.
OUTPUT: Modify exactly one file: go.mod
RULES:
- No explanation
- No analysis
- No extra text
- Output only the result

---

## How to verify (Manual)

1) بالا آوردن سرویس‌ها:
- `make docker-up` (یا `make docker-up-api-recreate` اگر فقط api/redis می‌خوای)

2) چک سلامت Redis:
- `docker compose ps`
- `docker compose logs redis --tail=50`
- `docker compose exec redis redis-cli ping`  → باید `PONG` بده

3) چک اینکه API واقعاً env را می‌بیند:
- `docker compose exec api sh -lc 'printenv | grep -E "REDIS_|DB_MAX_"'`

4) تست دستی endpointها:
- `curl -s -X POST http://localhost:8080/v1/redis/set`
- `curl -s http://localhost:8080/v1/redis/get`

5) برای load test:
- یک k6 script ساده برای GET و یک script برای SET بساز و RPS/latency را ثبت کن.


## Prompt 10 — Convert GET result to json

TASK: Update Redis GET benchmark handler to return structured JSON instead of a JSON string.
INPUT:
- File: internal/api/redis_benchmark_handlers.go
- In GET /v1/redis/get handler:
  - When found=true, parse the returned Redis string as JSON into a struct/map with fields:
    tenant_id, phone_number, otp_code (all strings)
  - If parsing fails, return 500 using middleware.ErrorHandler with apperrors.ErrInternalServerError("invalid stored json")
  - Response on success must be:
    {"found": true, "tenant_id": "...", "phone_number": "...", "otp_code": "..."}
- Keep existing behavior for missing key: {"found": false}
OUTPUT: Modify exactly one file: internal/api/redis_benchmark_handlers.go
RULES:
- No explanation
- No analysis
- No extra text
- Output only the result
