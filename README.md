# Go Backend Service

یک سرویس REST API ساده با استفاده از Go و Gin framework.

## مشخصات پروژه

- **زبان**: Go
- **Framework**: Gin
- **نوع پروژه**: REST API
- **معماری**: Monolithic Modular
- **دیتابیس**: PostgreSQL
- **احراز هویت**: JWT
- **لاگینگ**: Zerolog (Structured JSON Logging with Correlation IDs)
- **Tracing**: OpenTelemetry با پشتیبانی از Tempo
- **Metrics**: Prometheus (آماده برای پیاده‌سازی)
- **پیکربندی**: استفاده از environment variables با validation

## ساختار پروژه

```
.
├── bin/                 # Compiled binaries (generated)
├── cmd/
│   └── server/          # Entry point
├── internal/
│   ├── config/          # Configuration management
│   └── logger/          # Logging utilities
├── pkg/                 # Shared packages
├── configs/             # Configuration files
├── Dockerfile           # Multi-stage Docker build
├── docker-compose.yml   # Docker Compose configuration
└── Makefile            # Build automation

```

## نصب و راه‌اندازی

### پیش‌نیازها

- Go 1.21 یا بالاتر
- Docker و Docker Compose (برای اجرای با Docker)

### نصب وابستگی‌ها

```bash
make deps
# یا
go mod download
```

### تنظیم Environment Variables

یک فایل `.env` در ریشه پروژه ایجاد کنید:

```env
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_READ_TIMEOUT=15s
SERVER_WRITE_TIMEOUT=15s

# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=go_backend_db
DB_SSLMODE=disable

# JWT Configuration
JWT_SECRET_KEY=your-secret-key-change-in-production
JWT_REFRESH_SECRET=your-refresh-secret-key-change-in-production
JWT_EXPIRATION=24h

# Application Configuration
GIN_MODE=release
```

### اجرای پروژه

#### اجرای محلی (توصیه می‌شود برای Development)

برای راهنمای کامل اجرای محلی، به [LOCAL_DEVELOPMENT.md](./LOCAL_DEVELOPMENT.md) مراجعه کنید.

```bash
# راه‌اندازی کامل محیط توسعه (ایجاد .env و راه‌اندازی دیتابیس)
make dev

# اجرای برنامه با hot reload (توصیه می‌شود)
make dev-run

# یا اجرای ساده
make run
```

#### Debug با VS Code

برای راهنمای کامل debug با VS Code، به [VSCODE_DEBUG.md](./VSCODE_DEBUG.md) مراجعه کنید.

**راه‌اندازی سریع:**
1. نصب Go Extension در VS Code
2. راه‌اندازی دیتابیس: `make dev-db-up`
3. فشردن F5 برای شروع debug
4. قرار دادن breakpoint و debug کردن!

#### اجرا با Docker Compose (Production)

```bash
# ساخت و راه‌اندازی سرویس‌ها
make docker-up

# مشاهده لاگ‌ها
make docker-logs

# توقف سرویس‌ها
make docker-down
```

## استفاده از Makefile

```bash
make help          # نمایش تمام دستورات موجود

# Development (Local)
make dev           # راه‌اندازی کامل محیط توسعه
make dev-setup     # ایجاد فایل .env از env.example
make dev-db-up     # راه‌اندازی دیتابیس محلی
make dev-db-down   # توقف دیتابیس محلی
make dev-run       # اجرای برنامه با hot reload (air)
make run           # اجرای ساده برنامه

# Build & Test
make build         # ساخت پروژه
make test          # اجرای تست‌ها
make deps          # دانلود وابستگی‌ها

# Docker
make docker-build  # ساخت Docker image
make docker-up     # راه‌اندازی Docker containers
make docker-down   # توقف Docker containers
make docker-logs   # مشاهده لاگ‌های Docker

# Observability (OpenTelemetry, Tempo, Prometheus, Grafana)
make observability-up    # راه‌اندازی تمام observability stack
make observability-down  # توقف observability stack
make tempo-up            # راه‌اندازی Tempo + Jaeger
make prometheus-up       # راه‌اندازی Prometheus
make grafana-up          # راه‌اندازی Grafana

# Utilities
make clean         # پاکسازی فایل‌های build
make fmt           # فرمت کردن کد
make lint          # اجرای linter
```

برای جزئیات بیشتر، به [LOCAL_DEVELOPMENT.md](./LOCAL_DEVELOPMENT.md) مراجعه کنید.

## API Endpoints

### Health Check
```
GET /health
```

### Hello World
```
GET /hello
```

## ساخت Docker Image

```bash
make docker-build
# یا
docker build -t go-backend-service:latest .
```

## ساخت Binary

```bash
make build
# Binary در bin/go-backend-service قرار می‌گیرد

# اجرای مستقیم binary
./bin/go-backend-service
```

## تست

```bash
make test
# یا
go test ./...
```

## لاگینگ

این پروژه از Zerolog برای لاگینگ استفاده می‌کند و شامل:
- Structured JSON logging
- Correlation IDs برای ردیابی درخواست‌ها
- Trace ID و Span ID در logs (با OpenTelemetry)
- Log levels قابل تنظیم

## Observability (Tracing, Metrics, Logs)

این پروژه شامل پشتیبانی کامل از observability است:

- **OpenTelemetry Tracing**: برای distributed tracing
- **Tempo**: Backend برای ذخیره traces
- **Jaeger UI**: برای visualization traces
- **Prometheus**: برای metrics collection
- **Grafana**: برای visualization و dashboards

### راهنمای کامل Observability

برای راهنمای کامل و مثال‌های کاربردی، به [OBSERVABILITY.md](./OBSERVABILITY.md) مراجعه کنید.

**راه‌اندازی سریع:**

```bash
# راه‌اندازی تمام observability stack
make observability-up

# تنظیم environment variables برای Tempo
export OTEL_TRACING_ENABLED=true
export OTEL_TEMPO_ENABLED=true
export OTEL_TEMPO_ENDPOINT=localhost:4318

# اجرای API
make run

# دسترسی به رابط‌های کاربری:
# - Jaeger UI: http://localhost:16686
# - Prometheus: http://localhost:9090
# - Grafana: http://localhost:3000
```

## مجوز

MIT

