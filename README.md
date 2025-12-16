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
- **پیکربندی**: استفاده از environment variables با validation

## ساختار پروژه

```
.
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

#### اجرای مستقیم (بدون Docker)

```bash
make run
# یا
go run ./cmd/server
```

#### اجرا با Docker Compose

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
make build         # ساخت پروژه
make run           # اجرای پروژه
make test          # اجرای تست‌ها
make docker-build  # ساخت Docker image
make docker-up     # راه‌اندازی Docker containers
make docker-down   # توقف Docker containers
make clean         # پاکسازی فایل‌های build
```

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
- Log levels قابل تنظیم

## مجوز

MIT

