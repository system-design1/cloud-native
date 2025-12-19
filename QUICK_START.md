# 🚀 راهنمای شروع سریع

این راهنما برای تازه‌کارها طراحی شده است. اگر اولین بار است که این پروژه را اجرا می‌کنید، اینجا شروع کنید.

## ⚡ شروع در 3 مرحله

### مرحله 1: آماده‌سازی

```bash
# اگر پروژه را از Git کلون کرده‌اید:
cd sdgo

# ایجاد فایل .env (اگر وجود ندارد)
cp env.example .env
```

### مرحله 2: راه‌اندازی

```bash
# راه‌اندازی با Docker (ساده‌ترین روش)
make docker-up
```

**این دستور چه کار می‌کند؟**
- ✅ PostgreSQL database را راه‌اندازی می‌کند
- ✅ Docker image را می‌سازد (در اولین اجرا)
- ✅ API را روی پورت 8080 راه‌اندازی می‌کند

**زمان انتظار:** حدود 1-2 دقیقه در اولین اجرا

### مرحله 3: تست

```bash
# تست Health Check
curl http://localhost:8080/health
```

**خروجی مورد انتظار:**
```json
{"status":"ok","state":"ready"}
```

اگر این خروجی را دیدید، ✅ **موفق بودید!**

---

## 📋 دستورات مفید

### مشاهده وضعیت

```bash
# مشاهده لاگ‌ها
make docker-logs

# یا فقط لاگ API
docker-compose logs -f api

# بررسی وضعیت containers
docker ps
```

### تست API

```bash
# Health check
curl http://localhost:8080/health

# Hello World
curl http://localhost:8080/hello

# Readiness probe
curl http://localhost:8080/ready

# Liveness probe
curl http://localhost:8080/live
```

### توقف

```bash
# توقف تمام containers
make docker-down
```

---

## 🔄 بعد از تغییر کد

**مهم:** Docker به صورت خودکار کد را rebuild نمی‌کند.

بعد از تغییر کد، باید rebuild کنید:

```bash
# روش 1: Rebuild و restart (توصیه می‌شود)
make docker-up-rebuild

# روش 2: فقط rebuild API
docker-compose build api
docker-compose up -d api
```

---

## 🐛 مشکلات رایج

### مشکل: Port 8080 در حال استفاده است

```bash
# تغییر port در فایل .env
SERVER_PORT=8081

# سپس restart
make docker-down
make docker-up
```

### مشکل: Container از کد قدیمی استفاده می‌کند

```bash
# Rebuild کامل
make docker-up-rebuild
```

### مشکل: Database connection failed

```bash
# بررسی وضعیت PostgreSQL
docker ps | grep postgres

# Restart database
docker-compose restart postgres
```

---

## 📚 مراحل بعدی

حالا که پروژه را راه‌اندازی کردید:

1. **مستندات کامل**: [README.md](./README.md) را بخوانید
2. **Development**: برای development فعال، [LOCAL_DEVELOPMENT.md](./LOCAL_DEVELOPMENT.md) را ببینید
3. **API Endpoints**: لیست کامل endpoints در [README.md](./README.md#-api-endpoints)

---

## 💡 نکات مهم

- ✅ همیشه از `make docker-up` برای شروع استفاده کنید
- ✅ بعد از تغییر کد، از `make docker-up-rebuild` استفاده کنید
- ✅ لاگ‌ها را با `make docker-logs` بررسی کنید
- ⚠️ فایل `.env` را commit نکنید (در `.gitignore` است)

---

## ❓ سوالات متداول

**Q: آیا نیاز به نصب Go دارم؟**
A: خیر! برای اجرا با Docker، فقط Docker و Docker Compose کافی است.

**Q: چطور کد را تغییر دهم؟**
A: کد را در editor خود تغییر دهید، سپس `make docker-up-rebuild` را اجرا کنید.

**Q: چطور با hot reload کار کنم؟**
A: برای development فعال، از `make dev-run` استفاده کنید (نیاز به Go دارد).

**Q: چطور API را تست کنم؟**
A: از `curl` یا Postman استفاده کنید. مثال‌ها در [README.md](./README.md#-api-endpoints) است.

---

**موفق باشید! 🎉**

