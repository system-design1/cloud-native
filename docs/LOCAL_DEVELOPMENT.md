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

**نکته:** برای مشاهده traces در Jaeger/Tempo، باید:
1. Observability stack را راه‌اندازی کنید: `make observability-up`
2. (اختیاری) Health checker را اجرا کنید: `make dev-health-checker` (در terminal دیگر)
   - این script به صورت خودکار `/health`, `/ready` و `/live` را هر 10 ثانیه call می‌کند
   - Prometheus به صورت خودکار `/metrics` را scrape می‌کند

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

# اجرای health checker (برای ایجاد traces خودکار)
make dev-health-checker
```

### Health Checker (برای Local Development)

وقتی از `make dev-run` استفاده می‌کنید، health endpoints (`/health`, `/ready`, `/live`) به صورت خودکار call نمی‌شوند (برخلاف Docker که health checks وجود دارند).

برای ایجاد traces خودکار برای این endpoints:

```bash
# در یک terminal دیگر (بعد از make dev-run)
make dev-health-checker
```

این script:
- هر 10 ثانیه `/health`, `/ready` و `/live` را call می‌کند
- باعث ایجاد traces در Jaeger/Tempo می‌شود
- برای توقف: `Ctrl+C`

**تنظیم interval:**
```bash
HEALTH_CHECK_INTERVAL=5 make dev-health-checker  # هر 5 ثانیه
```

**نکته:** Prometheus به صورت خودکار `/metrics` را scrape می‌کند (اگر observability stack راه‌اندازی شده باشد).

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
# Terminal 1: راه‌اندازی اولیه
make dev

# Terminal 2: اجرای برنامه با hot reload
make dev-run

# Terminal 3 (اختیاری): Observability stack (Jaeger/Tempo/Prometheus)
make observability-up

# Terminal 4 (اختیاری): Health checker (برای ایجاد traces خودکار)
make dev-health-checker

# حالا می‌توانید کد را تغییر دهید و برنامه به صورت خودکار restart می‌شود

# برای توقف:
# Terminal 2: Ctrl+C برای توقف برنامه
# Terminal 4: Ctrl+C برای توقف health checker
# Terminal 3: make observability-down
# Terminal 1: make dev-db-down  # برای توقف دیتابیس
```

## Observability در Local Development

### Prometheus

Prometheus به صورت خودکار `/metrics` را scrape می‌کند (هر 5 ثانیه) اگر:
- Observability stack راه‌اندازی شده باشد: `make observability-up`
- Prometheus config به `host.docker.internal:8080` دسترسی داشته باشد (به صورت خودکار تنظیم شده)

### Health Endpoints

برای ایجاد traces خودکار برای `/health`, `/ready` و `/live`:
- از `make dev-health-checker` استفاده کنید (در terminal جداگانه)
- یا به صورت دستی call کنید:
  ```bash
  curl http://localhost:8080/health
  curl http://localhost:8080/ready
  curl http://localhost:8080/live
  ```

