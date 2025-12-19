# راهنمای اجرای محلی (Local Development)

این راهنما نحوه اجرای برنامه به صورت محلی (بدون Docker) را توضیح می‌دهد.

## پیش‌نیازها

- Go 1.25.1 یا بالاتر
- Docker و Docker Compose (فقط برای اجرای دیتابیس)
- PostgreSQL (اختیاری - می‌توانید از Docker استفاده کنید)

## راه‌اندازی سریع

### 1. راه‌اندازی کامل محیط توسعه

```bash
make dev
```

این دستور به صورت خودکار:
- فایل `.env` را از `env.example` ایجاد می‌کند
- دیتابیس PostgreSQL را در Docker راه‌اندازی می‌کند
- راهنمای استفاده را نمایش می‌دهد

### 2. اجرای برنامه

بعد از راه‌اندازی، می‌توانید برنامه را به دو روش اجرا کنید:

#### روش 1: اجرای ساده (بدون hot reload)

```bash
make run
```

#### روش 2: اجرا با hot reload (توصیه می‌شود)

```bash
make dev-run
```

این روش از `air` استفاده می‌کند که به صورت خودکار برنامه را با هر تغییر در کد rebuild و restart می‌کند.

## دستورات مفید

### مدیریت دیتابیس

```bash
# راه‌اندازی دیتابیس
make dev-db-up

# توقف دیتابیس
make dev-db-down
```

### سایر دستورات

```bash
# دانلود وابستگی‌های Go
make deps

# Build برنامه
make build

# اجرای تست‌ها
make test

# فرمت کردن کد
make fmt
```

## تنظیمات Environment Variables

فایل `.env` شامل تنظیمات زیر است:

```env
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database Configuration (برای اجرای محلی)
DB_HOST=localhost        # مهم: باید localhost باشد نه postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=go_backend_db
DB_SSLMODE=disable

# JWT Configuration
JWT_SECRET_KEY=your-secret-key-change-in-production-min-32-chars
JWT_REFRESH_SECRET=your-refresh-secret-key-change-in-production-min-32-chars
JWT_EXPIRATION=24h

# Application Configuration
GIN_MODE=debug          # برای development از debug استفاده کنید
```

### نکته مهم

برای اجرای محلی، `DB_HOST` باید `localhost` باشد (نه `postgres` که برای Docker استفاده می‌شود).

## نصب Air (Hot Reload)

اگر `air` نصب نیست، دستور `make dev-run` به صورت خودکار آن را نصب می‌کند. یا می‌توانید به صورت دستی نصب کنید:

```bash
go install github.com/air-verse/air@latest
```

## ساختار فایل‌ها

```
.
├── .env                    # فایل تنظیمات (باید ایجاد شود)
├── .air.toml              # تنظیمات air برای hot reload
├── docker-compose.dev.yml  # فقط دیتابیس برای development
├── docker-compose.yml      # تمام سرویس‌ها برای production
└── Makefile               # دستورات خودکار
```

## عیب‌یابی

### مشکل: "Error: .env file not found"

```bash
make dev-setup
```

### مشکل: "connection refused" برای دیتابیس

```bash
# بررسی کنید که دیتابیس در حال اجرا است
make dev-db-up

# یا بررسی کنید که پورت 5432 در دسترس است
docker ps | grep postgres
```

### مشکل: "air: command not found"

```bash
# نصب air
go install github.com/air-verse/air@latest

# یا استفاده از PATH کامل
export PATH=$PATH:$(go env GOPATH)/bin
```

## تفاوت بین اجرای محلی و Docker

| ویژگی | اجرای محلی | Docker |
|-------|-----------|--------|
| سرعت build | سریع‌تر | کندتر |
| Hot reload | بله (با air) | خیر |
| نیاز به Docker | فقط برای DB | بله |
| Debugging | آسان‌تر | سخت‌تر |
| محیط | مشابه production نیست | مشابه production |

## نکات مهم

1. **همیشه از `make dev-setup` استفاده کنید** برای ایجاد فایل `.env`
2. **`DB_HOST` را به `localhost` تغییر دهید** برای اجرای محلی
3. **از `make dev-run` استفاده کنید** برای development (hot reload)
4. **دیتابیس را با `make dev-db-up` راه‌اندازی کنید** قبل از اجرای برنامه

## مثال کامل

```bash
# 1. راه‌اندازی اولیه
make dev

# 2. در ترمینال دیگر، اجرای برنامه با hot reload
make dev-run

# 3. حالا می‌توانید کد را تغییر دهید و برنامه به صورت خودکار restart می‌شود

# 4. برای توقف
# Ctrl+C برای توقف برنامه
make dev-db-down  # برای توقف دیتابیس
```

