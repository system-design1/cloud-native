# راهنمای اجرای پروژه Go Backend Service

## پیش‌نیازها

- Go 1.21 یا بالاتر
- Docker و Docker Compose (برای اجرای با Docker)
- Make (اختیاری، اما توصیه می‌شود)

## روش‌های اجرا

### روش 1: اجرا با Docker Compose (توصیه می‌شود)

این روش هم API و هم PostgreSQL را راه‌اندازی می‌کند.

#### مرحله 1: بررسی فایل `.env`

مطمئن شوید فایل `.env` در ریشه پروژه وجود دارد. می‌توانید از `env.example` کپی کنید:

```bash
cp env.example .env
```

#### مرحله 2: راه‌اندازی سرویس‌ها

```bash
# راه‌اندازی تمام سرویس‌ها
make docker-up

# یا مستقیماً:
docker-compose up -d
```

این دستور:
- یک PostgreSQL container را راه‌اندازی می‌کند
- Docker image پروژه را می‌سازد (در اولین اجرا)
- API را روی پورت 8080 راه‌اندازی می‌کند

#### مرحله 3: بررسی وضعیت

```bash
# مشاهده لاگ‌ها
make docker-logs

# یا فقط لاگ API:
docker-compose logs -f api
```

#### مرحله 4: تست API

در ترمینال جدید:

```bash
# تست Health Check
curl http://localhost:8080/health

# تست Hello endpoint
curl http://localhost:8080/hello
```

انتظار می‌رود خروجی زیر را ببینید:

```json
# از /health:
{"status":"ok"}

# از /hello:
{"message":"Hello, World!"}
```

#### توقف سرویس‌ها

```bash
make docker-down
```

---

### روش 2: اجرای مستقیم (بدون Docker)

این روش نیاز به PostgreSQL نصب‌شده روی سیستم شما دارد.

#### مرحله 1: تنظیم فایل `.env`

برای اجرای مستقیم، `DB_HOST` باید `localhost` باشد:

```bash
# در فایل .env تغییر دهید:
DB_HOST=localhost
```

#### مرحله 2: راه‌اندازی PostgreSQL

اگر PostgreSQL نصب نیست، ابتدا نصب کنید. سپس دیتابیس را ایجاد کنید:

```bash
# وارد PostgreSQL شوید
sudo -u postgres psql

# در PostgreSQL shell:
CREATE DATABASE go_backend_db;
CREATE USER postgres WITH PASSWORD 'postgres';
GRANT ALL PRIVILEGES ON DATABASE go_backend_db TO postgres;
\q
```

#### مرحله 3: بارگذاری Environment Variables و اجرا

```bash
# بارگذاری environment variables و اجرا
export $(cat .env | grep -v '^#' | xargs) && make run
```

یا:

```bash
source <(cat .env | grep -v '^#' | sed 's/^/export /') && make run
```

#### مرحله 4: تست API

در ترمینال جدید:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/hello
```

---

## بررسی صحت پروژه

### 1. اجرای تست‌ها

```bash
# اجرای تمام تست‌ها
make test

# یا:
go test ./... -v
```

### 2. بررسی Build

```bash
# ساخت پروژه
make build

# بررسی وجود binary
ls -lh go-backend-service

# پاکسازی
make clean
```

### 3. بررسی Linter (اگر نصب باشد)

```bash
make lint
```

### 4. بررسی Docker Build

```bash
# ساخت Docker image
make docker-build

# بررسی images
docker images | grep go-backend-service
```

---

## عیب‌یابی

### مشکل: Port در حال استفاده است

```bash
# بررسی پروسس‌های در حال استفاده از پورت 8080
lsof -i :8080

# یا:
sudo netstat -tlnp | grep 8080
```

### مشکل: Docker container شروع نمی‌شود

```bash
# بررسی لاگ‌های container
docker-compose logs api

# بررسی وضعیت containers
docker-compose ps
```

### مشکل: خطای اتصال به دیتابیس

- مطمئن شوید PostgreSQL در حال اجرا است
- بررسی کنید environment variables در `.env` درست تنظیم شده‌اند
- اگر از Docker استفاده می‌کنید، مطمئن شوید `DB_HOST=postgres` باشد

### مشکل: Configuration validation failed

بررسی کنید تمام environment variables مورد نیاز در `.env` تعریف شده‌اند:
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `JWT_SECRET_KEY`

---

## دستورات مفید

```bash
# مشاهده تمام دستورات Makefile
make help

# راه‌اندازی کامل (deps + build + test)
make setup

# مشاهده لاگ‌های real-time
make docker-logs

# Restart سرویس‌ها
make docker-down && make docker-up

# پاکسازی کامل Docker (volumes را هم پاک می‌کند)
docker-compose down -v
```

---

## تست API با مثال‌های کامل

### استفاده از curl

```bash
# Health check
curl -v http://localhost:8080/health

# Hello endpoint
curl -v http://localhost:8080/hello

# با Correlation ID
curl -v -H "X-Correlation-ID: test-123" http://localhost:8080/hello
```

### استفاده از httpie (اگر نصب باشد)

```bash
http GET http://localhost:8080/health
http GET http://localhost:8080/hello
```

---

## بررسی لاگ‌ها

### در حالت Docker

```bash
# تمام لاگ‌ها
docker-compose logs

# فقط لاگ API
docker-compose logs api

# لاگ real-time
docker-compose logs -f api
```

### در حالت مستقیم

لاگ‌ها مستقیماً در console نمایش داده می‌شوند.

---

## نکات مهم

1. **Environment Variables**: فایل `.env` در `.gitignore` است و commit نمی‌شود. برای production، از روش‌های امن مدیریت secrets استفاده کنید.

2. **Security**: در production، حتماً `JWT_SECRET_KEY` و `DB_PASSWORD` را تغییر دهید.

3. **Database**: در production، `DB_SSLMODE` را `require` یا `verify-full` تنظیم کنید.

4. **GIN_MODE**: برای production، `GIN_MODE=release` تنظیم کنید.

