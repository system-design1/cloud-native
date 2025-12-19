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
│   └── tempo.yaml            # Tempo config
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
# روش 1: Rebuild و restart
make docker-up-rebuild

# روش 2: فقط rebuild API
docker-compose build api
docker-compose up -d api

# روش 3: Rebuild با --build flag
docker-compose up -d --build api
```

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

#### مرحله 4: توقف

```bash
# توقف دیتابیس
make dev-db-down

# توقف برنامه: Ctrl+C
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
make run              # اجرای ساده

# Docker
make docker-up        # راه‌اندازی containers
make docker-down      # توقف containers
make docker-logs      # مشاهده logs
make docker-build     # Build image
make docker-up-rebuild # Rebuild و restart

# Build & Test
make build            # Build binary
make test             # اجرای تست‌ها
make deps             # دانلود dependencies
```

### دستورات کامل

برای لیست کامل دستورات:
```bash
make help
```

---

## 📊 Observability

این پروژه شامل پشتیبانی کامل از Observability است:

- **OpenTelemetry Tracing**: Distributed tracing
- **Jaeger UI**: Visualization traces
- **Prometheus**: Metrics collection
- **Grafana**: Dashboards و visualization

### راه‌اندازی سریع

```bash
# راه‌اندازی تمام stack
make observability-up

# دسترسی به UI:
# - Jaeger: http://localhost:16686
# - Prometheus: http://localhost:9090
# - Grafana: http://localhost:3000 (admin/admin)
```

برای راهنمای کامل، به [OBSERVABILITY.md](./OBSERVABILITY.md) مراجعه کنید.

---

## 📚 مستندات بیشتر

- **[LOCAL_DEVELOPMENT.md](./LOCAL_DEVELOPMENT.md)**: راهنمای کامل development محلی
- **[VSCODE_DEBUG.md](./VSCODE_DEBUG.md)**: راهنمای debug با VS Code
- **[OBSERVABILITY.md](./OBSERVABILITY.md)**: راهنمای کامل Observability
- **[RUN_GUIDE.md](./RUN_GUIDE.md)**: راهنمای اجرا (قدیمی)

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
