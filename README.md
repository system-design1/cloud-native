# Go Backend Service

یک سرویس REST API ساده و production-ready با استفاده از Go و Gin framework.

## 📋 فهرست مطالب

- [شروع سریع](#-شروع-سریع)
- [پیش‌نیازها](#-پیشنیازها)
- [ساختار پروژه](#-ساختار-پروژه)
- [راه‌اندازی محیط Development](#-راهاندازی-محیط-development)
- [راه‌اندازی محیط Production](#-راهاندازی-محیط-production)
- [API Endpoints](#-api-endpoints)
- [استفاده از Makefile](#-استفاده-از-makefile)
- [Availability Lab: Traefik Gateway](#-availability-lab-traefik-gateway)
- [Observability](#-observability)
- [مستندات بیشتر](#-مستندات-بیشتر)

---

## 🚀 شروع سریع

### برای تازه‌کارها (اولین بار)

```bash
# 1. کلون کردن پروژه (اگر از Git استفاده می‌کنید)
git clone <repository-url>
cd sdgo

# 2. ایجاد فایل .env از نمونه
cp env.example .env

# 3. راه‌اندازی با Docker (ساده‌ترین روش)
make docker-up

# 4. تست API
curl http://localhost:8080/health
```

**خروجی مورد انتظار:**
```json
{"status":"ok","state":"ready"}
```

---

## 📦 پیش‌نیازها

### حداقل نیازمندی‌ها

- **Docker** 20.10+ و **Docker Compose** 2.0+ (برای اجرای با Docker)
- **Go** 1.21+ (فقط برای development محلی)
- **Make** (اختیاری اما توصیه می‌شود)

### نصب Docker

**Linux:**
```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
```

**macOS:**
```bash
brew install docker docker-compose
# یا دانلود Docker Desktop از docker.com
```

**Windows:**
دانلود و نصب [Docker Desktop](https://www.docker.com/products/docker-desktop)

### بررسی نصب

```bash
docker --version
docker-compose --version
```

---

## 📁 ساختار پروژه

```
sdgo/
├── cmd/
│   └── server/              # Entry point اصلی برنامه
│       └── main.go          # نقطه شروع برنامه
│
├── internal/                # کدهای داخلی (غیر قابل استفاده خارجی)
│   ├── api/                 # API handlers و routes
│   │   ├── handlers.go     # Handler functions
│   │   └── routes.go        # Route definitions
│   ├── config/              # مدیریت configuration
│   │   └── config.go        # بارگذاری و validation
│   ├── lifecycle/           # مدیریت lifecycle (ready/shutdown)
│   │   └── lifecycle.go
│   ├── logger/              # Logging utilities
│   │   └── logger.go        # Zerolog setup
│   ├── metrics/             # Prometheus metrics
│   │   └── metrics.go
│   ├── middleware/          # HTTP middleware
│   │   ├── correlation.go   # Correlation ID
│   │   ├── error_handler.go # Error handling
│   │   ├── logging.go       # Request/Response logging
│   │   ├── prometheus.go    # Metrics collection
│   │   └── tracing.go       # OpenTelemetry tracing
│   ├── server/              # HTTP server wrapper
│   │   └── server.go        # Server lifecycle
│   └── tracer/              # OpenTelemetry tracer
│       └── tracer.go
│
├── pkg/                      # Packages قابل استفاده خارجی
│   └── errors/              # Error definitions
│       └── errors.go
│
├── configs/                  # فایل‌های configuration
│   ├── prometheus.yml        # Prometheus config
│   ├── tempo.yaml            # Tempo config
│   ├── loki/                 # Loki configs
│   └── promtail/             # Promtail configs
│
├── deploy/                   # سناریوهای deploy و availability lab
│   └── availability-lab/
│       └── traefik-baseline/ # آزمایش local Traefik gateway
│
├── docs/                     # مستندات پروژه
│   ├── QUICK_START.md        # راهنمای سریع
│   ├── LOCAL_DEVELOPMENT.md  # راهنمای development
│   ├── LOAD_TESTING_K6_HELLO_CONCURRENCY.md  # راهنمای تست بار همزمانی
│   ├── LOAD_TESTING_TRACING_SAMPLING.md  # راهنمای sampling برای load testing
│   ├── OBSERVABILITY.md      # راهنمای Observability
│   ├── OBSERVABILITY_RESET.md  # راهنمای reset کردن observability stack
│   ├── LOKI_GUIDE.md         # راهنمای Loki
│   ├── PROMETHEUS_GUIDE.md   # راهنمای Prometheus
│   └── ...                   # سایر مستندات
│
├── docker-compose.yml        # Docker Compose برای production
├── docker-compose.dev.yml    # Docker Compose برای development DB
├── docker-compose.observability.yml  # Observability stack
├── Dockerfile                # Multi-stage Docker build
├── Makefile                  # Build automation
├── env.example               # نمونه فایل environment variables
└── README.md                 # این فایل
```

### توضیح ساختار

- **`cmd/server/`**: نقطه ورود برنامه. اینجا `main()` قرار دارد.
- **`internal/`**: کدهای داخلی که نباید از خارج پروژه استفاده شوند.
- **`pkg/`**: کدهای قابل استفاده خارجی (مثل libraries).
- **`configs/`**: فایل‌های configuration برای ابزارهای خارجی.
- **`deploy/availability-lab/traefik-baseline/`**: آزمایش availability محلی برای اجرای Traefik به عنوان gateway جلوی backend.

---

## 🛠️ راه‌اندازی محیط Development

### روش 1: با Docker (توصیه می‌شود برای شروع)

این روش ساده‌ترین است و نیازی به نصب Go ندارد.

#### مرحله 1: آماده‌سازی

```bash
# ایجاد فایل .env از نمونه
cp env.example .env

# بررسی فایل .env (مقادیر پیش‌فرض معمولاً کافی است)
cat .env
```

#### مرحله 2: راه‌اندازی

```bash
# راه‌اندازی تمام سرویس‌ها (PostgreSQL + API)
make docker-up
```

این دستور:
- ✅ PostgreSQL container را راه‌اندازی می‌کند
- ✅ Docker image را می‌سازد (در اولین اجرا)
- ✅ API container را راه‌اندازی می‌کند
- ✅ Health check را اجرا می‌کند

#### مرحله 3: بررسی وضعیت

```bash
# مشاهده لاگ‌ها
make docker-logs

# یا فقط لاگ API
docker-compose logs -f api

# بررسی وضعیت containers
docker ps
```

#### مرحله 4: تست API

```bash
# Health check
curl http://localhost:8080/health

# Readiness probe
curl http://localhost:8080/ready

# Liveness probe
curl http://localhost:8080/live

# Hello endpoint
curl http://localhost:8080/hello
```

#### مرحله 5: توقف

```bash
# توقف تمام containers
make docker-down
```

#### 🔄 Rebuild بعد از تغییر کد

**مهم:** Docker به صورت خودکار کد را rebuild نمی‌کند. بعد از تغییر کد:

```bash
# بهترین روش (روزمره): فقط API را rebuild کن (cache-friendly)
make docker-up-api-build

# اگر چند سرویس build دارند و می‌خواهی همه را rebuild کنی (cache-friendly)
make docker-up-rebuild

# فقط اگر cache خراب شده یا build به‌هم ریخته (خیلی کند)
make docker-up-no-cache

```

#### 📋 راهنمای Rebuild: چه زمانی چه چیزی باید rebuild شود؟

| نوع تغییر | دستور Rebuild | توضیحات |
|-----------|---------------|---------|
| **تغییرات در کد Go** (مثل handlers, middleware, config) | `make docker-up-api-build` | فقط باینری API باید دوباره build شود|
|**تغییر Dockerfile**|`make docker-up-api-build`|image API تغییر می‌کند|
|**تغییر go.mod / go.sum**|`make docker-up-api-build`|دانلود deps و build مجدد لازم است
| **تغییر .env یا environment در compose** | `make docker-up-api-recreate` | build لازم نیست؛ فقط recreate برای اعمال env|
| **تغییرات در docker-compose.yml** | `make docker-up` or `make docker-up-api-recreate` | فقط container `api` باید rebuild شود |
| **تغییرات در configs/tempo.yaml** | `make observability-up-rebuild` | فقط observability stack |
| **تغییرات در configs/prometheus.yml** | `make observability-up-rebuild` | فقط observability stack |
| **تغییرات در configs/loki/** یا **configs/promtail/** | `make observability-up-rebuild` | فقط observability stack |
| **تغییرات در docker-compose.observability.yml** | `make observability-up-rebuild` | فقط observability stack |

**نکته مهم:** 
- تغییرات در کد Go **نیازی به rebuild observability stack ندارد**
- تغییرات در config files observability **نیازی به rebuild application ندارد**

### روش 2: اجرای محلی با Hot Reload (برای Development)

این روش برای development بهتر است چون با هر تغییر کد، خودکار rebuild می‌شود.

#### مرحله 1: راه‌اندازی دیتابیس

```bash
# راه‌اندازی PostgreSQL
make dev-db-up
```

#### مرحله 2: تنظیم .env

```bash
# ایجاد .env (اگر وجود ندارد)
make dev-setup

# تغییر DB_HOST به localhost
# در فایل .env:
# DB_HOST=localhost
```

#### مرحله 3: اجرای برنامه

```bash
# اجرا با hot reload (توصیه می‌شود)
make dev-run

# یا اجرای ساده (بدون hot reload)
make run
```

**نکته:** `make dev-run` از `air` استفاده می‌کند که به صورت خودکار نصب می‌شود.

#### مرحله 4: Observability (اختیاری)

برای مشاهده traces در Jaeger/Tempo:

```bash
# Terminal 1: Application (از مرحله 3)
make dev-run

# Terminal 2: Observability stack
make observability-up

# Terminal 3: Health checker (برای ایجاد traces خودکار)
make dev-health-checker
```

**توضیحات:**
- **Prometheus**: به صورت خودکار `/metrics` را scrape می‌کند (هر 5 ثانیه)
- **Health Checker**: به صورت خودکار `/health`, `/ready` و `/live` را call می‌کند (هر 10 ثانیه)
- برای توقف health checker: `Ctrl+C`

#### مرحله 5: توقف

```bash
# توقف برنامه: Ctrl+C (در terminal که make dev-run اجرا شده)
# توقف health checker: Ctrl+C (در terminal که make dev-health-checker اجرا شده)
# توقف observability: make observability-down
# توقف دیتابیس: make dev-db-down
```

### مقایسه روش‌ها

| ویژگی | Docker (`make docker-up`) | Local (`make dev-run`) |
|-------|---------------------------|------------------------|
| نیاز به Go | ❌ | ✅ |
| Hot Reload | ❌ (نیاز به rebuild) | ✅ (خودکار) |
| سرعت تغییرات | کند (نیاز به rebuild) | سریع (instant) |
| مناسب برای | Testing, Production | Development |
| پیچیدگی | ساده | متوسط |

**توصیه:**
- **شروع کار**: از `make docker-up` استفاده کنید
- **Development فعال**: از `make dev-run` استفاده کنید

---

## 🏭 راه‌اندازی محیط Production

### پیش‌نیازها

1. فایل `.env` با مقادیر production
2. `JWT_SECRET_KEY` و `JWT_REFRESH_SECRET` باید تغییر کنند
3. `GIN_MODE=release`

### مرحله 1: تنظیم Environment Variables

```bash
# کپی از نمونه
cp env.example .env

# ویرایش .env و تغییر مقادیر مهم:
# - JWT_SECRET_KEY (حداقل 32 کاراکتر)
# - JWT_REFRESH_SECRET (حداقل 32 کاراکتر)
# - GIN_MODE=release
# - LOG_LEVEL=info
```

### مرحله 2: Build Docker Image

```bash
# Build image
make docker-build

# یا force rebuild
make docker-build-rebuild
```

### مرحله 3: اجرا

```bash
# با Docker Compose
make docker-up

# یا با Docker مستقیم
docker run -d \
  --name go-backend-api \
  -p 8080:8080 \
  --env-file .env \
  go-backend-service:latest
```

### مرحله 4: بررسی Health

```bash
# Health check
curl http://localhost:8080/health

# Readiness (برای Kubernetes)
curl http://localhost:8080/ready

# Liveness (برای Kubernetes)
curl http://localhost:8080/live
```

### مرحله 5: Monitoring

```bash
# مشاهده logs
docker logs -f go-backend-api

# یا با Docker Compose
docker-compose logs -f api
```

### نکات Production

1. **Secrets Management**: از Docker Secrets یا Kubernetes Secrets استفاده کنید
2. **Logging**: Logs به `stdout` می‌روند. از log aggregation استفاده کنید
3. **Health Checks**: از `/ready` و `/live` برای Kubernetes probes استفاده کنید
4. **Graceful Shutdown**: برنامه از graceful shutdown پشتیبانی می‌کند
5. **Metrics**: از `/metrics` برای Prometheus scraping استفاده کنید

---

## 🔌 API Endpoints

### Health & Lifecycle

| Endpoint | Method | توضیحات | استفاده |
|----------|--------|---------|---------|
| `/health` | GET | Health check عمومی | Docker healthcheck |
| `/ready` | GET | Readiness probe | Kubernetes readiness |
| `/live` | GET | Liveness probe | Kubernetes liveness |

**مثال:**
```bash
curl http://localhost:8080/health
# {"status":"ok","state":"ready"}

curl http://localhost:8080/ready
# {"status":"ready","state":"ready"}

curl http://localhost:8080/live
# {"status":"alive","state":"ready"}
```

### Application Endpoints

| Endpoint | Method | توضیحات |
|----------|--------|---------|
| `/hello` | GET | پیام Hello World |
| `/delayed-hello` | GET | Hello با delay تصادفی (1-3 ثانیه) |
| `/test-error` | GET | تست error handling |
| `/metrics` | GET | Prometheus metrics |

**مثال:**
```bash
curl http://localhost:8080/hello
# {"message":"Hello, World!"}

curl http://localhost:8080/metrics
# # HELP http_request_duration_seconds Duration of HTTP requests...
```

### Versioned API (v1)

#### OTP Service

| Endpoint | Method | توضیحات |
|----------|--------|---------|
| `/v1/otp/code` | POST | Generate a 6-digit OTP code برای benchmark و تست ساده |
| `/v1/otp/send` | POST | شروع flow واقعی OTP |
| `/v1/otp/verify` | POST | بررسی OTP و پایان مصرف یک‌بارمصرف آن |

Flow فعلی OTP شامل Redis state، fake SMS provider، request logging، verification logging، resend protection و send rate limiting است. جزئیات بیشتر در [current-state.md](./docs/current-state.md) و [architecture.md](./docs/architecture.md) نگهداری می‌شود.

**Response:**
```json
{
  "code": "123456"
}
```

**مثال:**
```bash
curl -X POST http://localhost:8080/v1/otp/code
# {"code":"123456"}
```

---

## 🎯 استفاده از Makefile

### دستورات اصلی

```bash
# نمایش تمام دستورات
make help

# Development
make dev              # راه‌اندازی کامل محیط dev
make dev-setup        # ایجاد .env
make dev-db-up        # راه‌اندازی دیتابیس
make dev-run          # اجرا با hot reload
make dev-health-checker  # اجرای health checker (برای ایجاد traces خودکار)
make run              # اجرای ساده

# Docker
make docker-up        # راه‌اندازی containers
make docker-down      # توقف containers
make docker-logs      # مشاهده logs
make docker-build     # Build image
make docker-up-rebuild # Rebuild و restart
make docker-up-api-build # زمان تغییر در کد go
make docker-up-api-recreate # زمان تغییر در فایل env و ساخته شدن مجدد کانتینر
make docker-up-no-cache

# Build & Test
make build            # Build binary
make test             # اجرای تست‌ها
make deps             # دانلود dependencies
```

### Availability Lab: Traefik Gateway

این lab یک gateway ساده و local جلوی backend قرار می‌دهد:

```text
Client -> Traefik -> OTP Service
```

حالت‌های دسترسی:

| حالت | آدرس |
|------|------|
| Direct backend mode | `http://localhost:8080` |
| Traefik gateway mode | `http://localhost:8081` |
| Traefik dashboard | `http://localhost:8082/dashboard/` |

Traefik در این lab همه مسیرها را با ``PathPrefix(`/`)`` به backend می‌فرستد و برای اثبات عبور traffic از gateway، header زیر را اضافه می‌کند:

```text
X-Gateway-Node: traefik-baseline
```

دستورات اصلی:

```bash
make traefik-config      # validate کردن compose lab
make traefik-up          # اجرای فقط Traefik gateway
make traefik-down        # توقف فقط Traefik gateway
make traefik-logs        # مشاهده logs
make traefik-ps          # مشاهده containerهای lab
make traefik-stack-up    # اجرای backend stack و سپس Traefik
make traefik-stack-down  # توقف Traefik و سپس backend stack
```

راهنمای کامل این lab در [`deploy/availability-lab/traefik-baseline/README.md`](./deploy/availability-lab/traefik-baseline/README.md) قرار دارد.

### دستورات کامل

برای لیست کامل دستورات:
```bash
make help
```

---

## 📊 Observability

این پروژه شامل پشتیبانی کامل از Observability است:

- **OpenTelemetry Tracing**: Distributed tracing
- **Tempo**: Trace storage backend
- **Jaeger UI**: Visualization traces (اختیاری)
- **Prometheus**: Metrics collection
- **Loki**: Log aggregation و central logging
- **Promtail**: Log collector از Docker containers
- **Grafana**: Dashboards و visualization (traces, logs, metrics)

### راه‌اندازی سریع

```bash
# راه‌اندازی تمام stack (Tempo, Jaeger, Prometheus, Loki, Grafana)
make observability-up

# بعد از تغییر config files (tempo.yaml, prometheus.yml, loki-config.yaml):
make observability-up-rebuild

# دسترسی به UI:
# - Grafana: http://localhost:3000 (admin/admin) - برای traces, logs, metrics
# - Jaeger: http://localhost:16686 (memory storage only)
# - Prometheus: http://localhost:9090
# - Loki API: http://localhost:3100
# - Tempo API: http://localhost:3200
```

### مشاهده Logs در Grafana

1. باز کردن: http://localhost:3000
2. رفتن به **Explore** (منوی سمت چپ)
3. انتخاب **Loki** datasource
4. جستجوی logs:
   ```logql
   {container="go-backend-api"}
   ```

برای راهنمای کامل، به [OBSERVABILITY.md](./docs/OBSERVABILITY.md) و [LOKI_GUIDE.md](./docs/LOKI_GUIDE.md) مراجعه کنید.

### Route-Based Tracing Policy (Always/Ratio/Drop)

این پروژه از **Route-Based Tracing Policy** پشتیبانی می‌کند که به شما امکان کنترل sampling traces بر اساس route را می‌دهد. این قابلیت برای کاهش noise در Jaeger/Tempo و تمرکز روی traces مهم مفید است.

#### سه نوع Policy

1. **ALWAYS**: همیشه trace می‌شود
   - برای endpoints مهم که می‌خواهید همیشه trace شوند
   - مثال: `/delayed-hello`, `/test-error`

2. **RATIO**: با احتمال مشخص trace می‌شود
   - برای endpoints پرترافیک که می‌خواهید گاهی trace شوند
   - مثال: `/health=0.01` (1% از requests)
   - مقدار باید بین `0.0` و `1.0` باشد

3. **DROP**: هرگز trace نمی‌شود
   - برای endpoints پرترافیک که نمی‌خواهید trace شوند
   - مثال: `/metrics`

#### ترتیب اولویت (Precedence)

1. **DROP** (بالاترین اولویت)
2. **ALWAYS**
3. **RATIO**
4. **DEFAULT** policy

#### تنظیمات پیش‌فرض (Demo-friendly)

با تنظیمات پیش‌فرض:
- `/delayed-hello` و `/test-error`: همیشه trace می‌شوند
- `/health`, `/live`, `/ready`: 1% از requests trace می‌شوند
- `/metrics`: trace نمی‌شود (DROP)

#### مثال Configuration

```env
# فعال‌سازی route-based policy
OTEL_ROUTE_POLICY_ENABLED=true

# Routes که همیشه trace می‌شوند
OTEL_ROUTE_ALWAYS=/delayed-hello,/test-error

# Routes که هرگز trace نمی‌شوند
OTEL_ROUTE_DROP=/metrics

# Routes با sampling ratio
OTEL_ROUTE_RATIO=/health=0.01,/live=0.01,/ready=0.01

# Default policy برای routes دیگر
OTEL_ROUTE_DEFAULT=always

# Default ratio (فقط برای OTEL_ROUTE_DEFAULT=ratio)
OTEL_ROUTE_DEFAULT_RATIO=1.0
```

#### غیرفعال کردن Policy

برای غیرفعال کردن policy و استفاده از رفتار پیش‌فرض (sample همه traces):

```env
OTEL_ROUTE_POLICY_ENABLED=false
```

#### نکات مهم

- Policy فقط زمانی اعمال می‌شود که `OTEL_ROUTE_POLICY_ENABLED=true` باشد
- وقتی policy غیرفعال است، همه traces sample می‌شوند (رفتار پیش‌فرض)
- برای debugging، می‌توانید policy را غیرفعال کنید تا همه traces را ببینید
- Routes با query string هم درست کار می‌کنند (فقط path بررسی می‌شود)

برای جزئیات بیشتر، به [OBSERVABILITY.md](./docs/OBSERVABILITY.md) مراجعه کنید.

---

## 📚 مستندات بیشتر

تمام مستندات در پوشه [`docs/`](./docs/) قرار دارند:

### راهنماهای اصلی
- **[QUICK_START.md](./docs/QUICK_START.md)**: راهنمای سریع شروع کار
- **[LOCAL_DEVELOPMENT.md](./docs/LOCAL_DEVELOPMENT.md)**: راهنمای کامل development محلی
- **[TROUBLESHOOTING.md](./docs/TROUBLESHOOTING.md)**: راهنمای عیب‌یابی مشکلات رایج

### Load Testing
- **[LOAD_TESTING_K6_HELLO_CONCURRENCY.md](./docs/LOAD_TESTING_K6_HELLO_CONCURRENCY.md)**: راهنمای کامل تست بار همزمانی با k6 برای endpoint `/hello`
- **[LOAD_TESTING_TRACING_SAMPLING.md](./docs/LOAD_TESTING_TRACING_SAMPLING.md)**: راهنمای جلوگیری از overload شدن Jaeger/Tempo در طول تست‌های با بار بالا با تنظیم sampling policy
- **OTP Service – Performance Report (Phase 1, FA):** [OTP_Performance_Report_Phase1_FA.md](./docs/OTP_Performance_Report_Phase1_FA.md)

### Observability
- **[OBSERVABILITY.md](./docs/OBSERVABILITY.md)**: راهنمای کامل Observability (Tempo, Prometheus, Grafana)
- **[OBSERVABILITY_RESET.md](./docs/OBSERVABILITY_RESET.md)**: راهنمای reset کردن observability stack و بازیابی Jaeger بعد از تست‌های با بار بالا
- **[LOKI_GUIDE.md](./docs/LOKI_GUIDE.md)**: راهنمای کامل Loki و Central Logging
- **[LOGGING_GUIDE.md](./docs/LOGGING_GUIDE.md)**: راهنمای مشاهده و مدیریت لاگ‌ها
- **[PROMETHEUS_GUIDE.md](./docs/PROMETHEUS_GUIDE.md)**: راهنمای مشاهده Metrics در Grafana
- **[TEMPO_GUIDE.md](./docs/TEMPO_GUIDE.md)**: راهنمای کامل Tempo و مشاهده Traces
- **[TEMPO_QUICK_START.md](./docs/TEMPO_QUICK_START.md)**: راهنمای سریع Tempo

### Development & Debugging
- **[VSCODE_DEBUG_GUIDE.md](./docs/VSCODE_DEBUG_GUIDE.md)**: راهنمای کامل debug با VS Code
- **[VSCODE_DEBUG.md](./docs/VSCODE_DEBUG.md)**: راهنمای debug با VS Code (نسخه قدیمی)
- **[DEBUG_TIMEOUT_FIX.md](./docs/DEBUG_TIMEOUT_FIX.md)**: راهنمای رفع مشکل Timeout در Debug
- **[DEBUG_TIPS.md](./docs/DEBUG_TIPS.md)**: نکات و ترفندهای Debug

### Docker & Infrastructure
- **[DOCKER_VERSIONING.md](./docs/DOCKER_VERSIONING.md)**: راهنمای Versioning در Docker Images
- **[DOCKER_BUILD_FIX.md](./docs/DOCKER_BUILD_FIX.md)**: راهنمای رفع مشکلات Docker Build

### سایر
- **[RUN_GUIDE.md](./docs/RUN_GUIDE.md)**: راهنمای اجرا (قدیمی - برای مرجع)

---

## 🐛 عیب‌یابی (Troubleshooting)

### مشکل: Container از کد قدیمی استفاده می‌کند

**راه‌حل:**
```bash
# Rebuild container
make docker-up-rebuild

# یا
docker-compose build api
docker-compose up -d api
```

### مشکل: Port 8080 در حال استفاده است

**راه‌حل:**
```bash
# تغییر port در .env
SERVER_PORT=8081

# یا توقف برنامه استفاده‌کننده از port
sudo lsof -i :8080
kill -9 <PID>
```

### مشکل: Database connection failed

**راه‌حل:**
```bash
# بررسی وضعیت PostgreSQL
docker ps | grep postgres

# بررسی logs
docker-compose logs postgres

# Restart database
docker-compose restart postgres
```

### مشکل: `/ready` یا `/live` 404 می‌دهد

**راه‌حل:**
```bash
# Container از کد قدیمی استفاده می‌کند
make docker-up-rebuild
```

---

## 📝 Environment Variables

تمام متغیرهای محیطی در `env.example` تعریف شده‌اند:

```bash
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_READ_TIMEOUT=15s
SERVER_WRITE_TIMEOUT=15s
SERVER_IDLE_TIMEOUT=120s
SERVER_GRACEFUL_SHUTDOWN_TIMEOUT=10s

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=go_backend_db
DB_SSLMODE=disable

# JWT
JWT_SECRET_KEY=your-secret-key-change-in-production-min-32-chars
JWT_REFRESH_SECRET=your-refresh-secret-key-change-in-production-min-32-chars
JWT_EXPIRATION=24h

# Application
GIN_MODE=release  # یا debug برای development
LOG_LEVEL=info    # debug, info, warn, error

# OTP
OTP_CODE_LENGTH=6
OTP_TTL=2m
OTP_MAX_ATTEMPTS=3
OTP_TENANT_CACHE_TTL=5m
OTP_PROVIDER_TIMEOUT=2s
OTP_FAKE_SMS_MIN_DELAY=20ms
OTP_FAKE_SMS_MAX_DELAY=30ms
OTP_FAKE_SMS_DEBUG_CODE_REDIS=false
OTP_FAKE_SMS_DEBUG_CODE_TTL=60s
OTP_SEND_RATE_LIMIT_ENABLED=false
OTP_SEND_RATE_LIMIT_MAX=5
OTP_SEND_RATE_LIMIT_WINDOW=10m

# OpenTelemetry
OTEL_TRACING_ENABLED=true
OTEL_SERVICE_NAME=go-backend-service
OTEL_SERVICE_VERSION=1.0.0
OTEL_JAEGER_ENABLED=true
OTEL_JAEGER_ENDPOINT=jaeger:4318
```

---

## 🔒 امنیت

- ✅ Non-root user در Docker
- ✅ Graceful shutdown
- ✅ Health checks
- ✅ Structured logging
- ✅ Error handling
- ⚠️ **مهم**: در production، `JWT_SECRET_KEY` را تغییر دهید

---

## 📄 مجوز

MIT

---

## 🤝 مشارکت

برای مشارکت در پروژه، لطفاً:
1. Issue ایجاد کنید
2. Fork کنید
3. Branch جدید بسازید
4. تغییرات را commit کنید
5. Pull Request ارسال کنید

---

## 📞 پشتیبانی

برای سوالات و مشکلات:
- Issue در GitHub ایجاد کنید
- مستندات را بررسی کنید
- Logs را بررسی کنید: `make docker-logs`
