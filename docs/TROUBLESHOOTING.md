# 🔧 راهنمای عیب‌یابی (Troubleshooting)

این راهنما مشکلات رایج و راه‌حل‌های آن‌ها را پوشش می‌دهد.

---

## 🐛 مشکلات رایج

### 1. Port 8080 در حال استفاده است

**خطا:**
```
listen tcp 0.0.0.0:8080: bind: address already in use
```

**راه‌حل:**

```bash
# روش 1: توقف Docker containers
make docker-down

# روش 2: پیدا کردن و kill کردن process
lsof -i :8080
kill -9 <PID>

# روش 3: تغییر port در .env
# در فایل .env تغییر دهید:
SERVER_PORT=8081
```

---

### 2. Container از کد قدیمی استفاده می‌کند

**مشکل:** بعد از تغییر کد، تغییرات اعمال نمی‌شوند.

**راه‌حل:**

```bash
# سریع‌ترین راه
make docker-up-api-build

# Rebuild کامل
make docker-up-rebuild

# یا فقط rebuild API
docker-compose build api
docker-compose up -d api
```

**نکته:** Docker به صورت خودکار کد را rebuild نمی‌کند. بعد از هر تغییر کد باید rebuild کنید.

---

### 3. Database connection failed

**خطا:**
```
connection refused
dial tcp: lookup postgres
```

**راه‌حل:**

```bash
# بررسی وضعیت PostgreSQL
docker ps | grep postgres

# اگر container در حال اجرا نیست:
make dev-db-up  # برای local development
# یا
make docker-up  # برای Docker Compose

# بررسی logs
docker-compose logs postgres
```

**برای local development:**
- مطمئن شوید `DB_HOST=localhost` در `.env` است
- مطمئن شوید database container در حال اجرا است: `make dev-db-up`

**برای Docker:**
- مطمئن شوید `DB_HOST=postgres` در `.env` است
- مطمئن شوید تمام containers در حال اجرا هستند: `docker ps`

---

### 4. `/ready` یا `/live` endpoint 404 می‌دهد

**مشکل:** Endpoint‌های جدید وجود ندارند.

**راه‌حل:**

```bash
# Container از کد قدیمی استفاده می‌کند
make docker-up-rebuild
```

---

### 5. `air` نصب نمی‌شود یا پیدا نمی‌شود

**خطا:**
```
air: command not found
/bin/sh: air: not found
```

**راه‌حل:**

```bash
# نصب دستی
go install github.com/air-verse/air@latest

# بررسی نصب
which air
# یا
ls -la $(go env GOPATH)/bin/air

# اضافه کردن به PATH (اگر نیاز است)
export PATH=$PATH:$(go env GOPATH)/bin

# اضافه کردن دائمی به PATH (در ~/.bashrc یا ~/.zshrc)
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
```

**نکته:** Repository `air` از `github.com/cosmtrek/air` به `github.com/air-verse/air` منتقل شده است.

---

### 6. `make dev-run` کار نمی‌کند

**مشکلات احتمالی:**

1. **Port در حال استفاده:**
   ```bash
   make docker-down
   pkill -f "air|go-backend-service"
   ```

2. **DB_HOST اشتباه:**
   ```bash
   # برای local development باید localhost باشد
   make dev-setup  # این خودکار DB_HOST را تنظیم می‌کند
   ```

3. **Database در حال اجرا نیست:**
   ```bash
   make dev-db-up
   ```

4. **air نصب نیست:**
   ```bash
   go install github.com/air-verse/air@latest
   ```

---

### 7. Environment variables اعمال نمی‌شوند

**راه‌حل:**

```bash
# بررسی فایل .env
cat .env

# بررسی environment variables در container
docker exec go-backend-api env | grep SERVER

# Recreate container
make docker-up-api-recreate

```

---

### 8. Build با خطا مواجه می‌شود

**خطاهای رایج:**

1. **Alpine package manager:**
   ```bash
   # Option A: 
   # اگر build با خطا مواجه شد، cache را پاک کنید
   docker compose build --no-cache api
   docker compose up -d api

   # Option B: create whole stack again
   make docker-up-no-cache

   ```

2. **Go modules:**
   ```bash
   # پاک کردن cache
   go clean -modcache
   go mod download
   ```

---

### 9. Health check unhealthy است

**بررسی:**

```bash
# بررسی وضعیت container
docker ps

# بررسی logs
docker logs go-backend-api

# تست دستی health endpoint
curl http://localhost:8080/health
curl http://localhost:8080/ready
curl http://localhost:8080/live
```

**راه‌حل:**

```bash
# Restart container
make docker-up-api-recreate

# یا rebuild (if code has been changed)
make docker-up-api-build
```

---

## 🔍 دستورات مفید برای Debug

```bash
# مشاهده تمام containers
docker ps -a

# مشاهده logs
make docker-logs
docker-compose logs -f api

# بررسی network
docker network ls
docker network inspect sdgo_app-network

# بررسی volume
docker volume ls

# بررسی environment variables
docker exec go-backend-api env

# بررسی process در container
docker exec go-backend-api ps aux

# بررسی port
lsof -i :8080
netstat -tuln | grep 8080
```

---

## 📝 چک‌لیست عیب‌یابی

قبل از درخواست کمک، این موارد را بررسی کنید:

- [ ] Docker و Docker Compose نصب هستند: `docker --version`
- [ ] فایل `.env` وجود دارد: `ls -la .env`
- [ ] `DB_HOST` درست تنظیم شده (localhost برای dev، postgres برای Docker)
- [ ] Port 8080 آزاد است: `lsof -i :8080`
- [ ] Database در حال اجرا است: `docker ps | grep postgres`
- [ ] Container در حال اجرا است: `docker ps | grep go-backend-api`
- [ ] Logs را بررسی کرده‌اید: `make docker-logs`
- [ ] Health endpoint کار می‌کند: `curl http://localhost:8080/health`

---

## 🆘 درخواست کمک

اگر مشکل حل نشد:

1. **Logs را جمع‌آوری کنید:**
   ```bash
   make docker-logs > logs.txt 2>&1
   ```

2. **وضعیت سیستم را بررسی کنید:**
   ```bash
   docker ps -a > containers.txt
   docker-compose config > config.txt
   ```

3. **مشکل را در GitHub Issue گزارش دهید** با:
   - توضیح مشکل
   - دستورات اجرا شده
   - خروجی logs
   - نسخه Docker و Go

---

**موفق باشید! 🎉**

