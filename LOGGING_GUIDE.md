# 📋 راهنمای مشاهده و مدیریت لاگ‌ها

## 📍 وضعیت فعلی لاگ‌ها

در حال حاضر، لاگ‌ها به صورت **structured JSON** با **Zerolog** تولید می‌شوند و به **stdout** می‌روند.

### در Docker:
- لاگ‌ها در **Docker logs** ذخیره می‌شوند
- **هیچ central logging solution** (مثل ELK یا Loki) نصب نشده است
- لاگ‌ها فقط در container logs قابل مشاهده هستند

---

## 🔍 روش‌های مشاهده لاگ‌ها

### 1. مشاهده لاگ‌های Docker Container (روش فعلی)

#### مشاهده لاگ‌های API:

```bash
# مشاهده لاگ‌های زنده (follow)
docker logs -f go-backend-api

# مشاهده آخرین 100 خط
docker logs --tail 100 go-backend-api

# مشاهده لاگ‌های از یک زمان خاص
docker logs --since 10m go-backend-api

# مشاهده لاگ‌های با timestamp
docker logs -t go-backend-api
```

#### با docker-compose:

```bash
# مشاهده لاگ‌های همه سرویس‌ها
docker-compose logs -f

# فقط API
docker-compose logs -f api

# آخرین 50 خط
docker-compose logs --tail 50 api
```

#### با Makefile:

```bash
# مشاهده لاگ‌های API
make docker-logs

# یا
make logs
```

---

### 2. فیلتر کردن لاگ‌ها

#### جستجو بر اساس Correlation ID:

```bash
# پیدا کردن لاگ‌های یک request خاص
CORRELATION_ID="abc-123-def-456"
docker logs go-backend-api 2>&1 | grep "$CORRELATION_ID"
```

#### جستجو بر اساس Trace ID:

```bash
# پیدا کردن لاگ‌های یک trace خاص
TRACE_ID="1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p"
docker logs go-backend-api 2>&1 | grep "$TRACE_ID"
```

#### جستجو بر اساس Level:

```bash
# فقط Error logs
docker logs go-backend-api 2>&1 | grep '"level":"error"'

# فقط Info logs
docker logs go-backend-api 2>&1 | grep '"level":"info"'
```

#### جستجو بر اساس Path:

```bash
# لاگ‌های یک endpoint خاص
docker logs go-backend-api 2>&1 | grep '"/hello"'
```

#### استفاده از jq برای JSON parsing:

```bash
# نصب jq (اگر نصب نیست)
sudo apt-get install jq  # Ubuntu/Debian
# یا
brew install jq  # macOS

# فیلتر کردن فقط error logs
docker logs go-backend-api 2>&1 | jq 'select(.level == "error")'

# فیلتر کردن بر اساس correlation_id
docker logs go-backend-api 2>&1 | jq 'select(.correlation_id == "abc-123")'

# نمایش فقط فیلدهای مهم
docker logs go-backend-api 2>&1 | jq '{timestamp, level, message, correlation_id, trace_id}'
```

---

### 3. ذخیره لاگ‌ها در فایل

```bash
# ذخیره لاگ‌ها در فایل
docker logs go-backend-api > api-logs.txt

# ذخیره با append
docker logs go-backend-api >> api-logs.txt

# ذخیره لاگ‌های از یک زمان خاص
docker logs --since 1h go-backend-api > last-hour-logs.txt
```

---

## 🚀 Loki - Central Logging (راه‌حل پیشنهادی)

**Loki** یک log aggregation system است که:
- ✅ لاگ‌ها را centralize می‌کند
- ✅ در Grafana قابل مشاهده است
- ✅ جستجو و فیلتر آسان است
- ✅ با Prometheus و Tempo یکپارچه است

### مزایای Loki:

1. ✅ **Centralized Logging**: همه لاگ‌ها در یک جا
2. ✅ **Grafana Integration**: مشاهده لاگ‌ها در Grafana
3. ✅ **Trace-to-Logs**: می‌توانید از trace به logs بروید
4. ✅ **Powerful Queries**: LogQL برای جستجوی پیشرفته
5. ✅ **Efficient Storage**: storage بهینه‌تر از ELK

### راه‌اندازی Loki:

```bash
# راه‌اندازی Loki و Promtail
make loki-up

# یا راه‌اندازی کامل observability stack
make observability-up
```

### مشاهده Logs در Grafana:

1. باز کردن: http://localhost:3000
2. رفتن به **Explore** (منوی سمت چپ)
3. انتخاب **Loki** datasource
4. جستجوی logs:
   ```logql
   {container="go-backend-api"}
   ```

برای جزئیات کامل، به [LOKI_GUIDE.md](LOKI_GUIDE.md) مراجعه کنید.

---

## 📝 مثال‌های عملی

### مثال 1: Debug کردن یک Request

```bash
# 1. ارسال request با correlation ID
CORRELATION_ID="debug-$(date +%s)"
curl -H "X-Correlation-ID: $CORRELATION_ID" http://localhost:8080/hello

# 2. پیدا کردن لاگ‌های این request
docker logs go-backend-api 2>&1 | grep "$CORRELATION_ID"
```

### مثال 2: پیدا کردن Error ها

```bash
# تمام error logs از 1 ساعت پیش
docker logs --since 1h go-backend-api 2>&1 | grep '"level":"error"'

# با jq (بهتر)
docker logs --since 1h go-backend-api 2>&1 | jq 'select(.level == "error")'
```

### مثال 3: بررسی Performance

```bash
# پیدا کردن slow requests (latency > 1000ms)
docker logs go-backend-api 2>&1 | jq 'select(.latency_ms > 1000)'
```

---

## 🔧 تنظیمات لاگ‌ها

### تغییر Log Level:

در فایل `.env` یا `docker-compose.yml`:

```env
LOG_LEVEL=debug  # trace, debug, info, warn, error, fatal, panic
```

سپس restart کنید:

```bash
docker-compose restart api
```

### Log Levels:

- **trace**: همه چیز (خیلی verbose)
- **debug**: اطلاعات debug
- **info**: اطلاعات عمومی (default)
- **warn**: هشدارها
- **error**: فقط errors
- **fatal**: فقط fatal errors
- **panic**: فقط panic errors

---

## 📊 ساختار لاگ‌ها

لاگ‌ها به صورت JSON هستند و شامل این فیلدها می‌شوند:

```json
{
  "level": "info",
  "timestamp": 1766149771,
  "message": "HTTP request/response",
  "correlation_id": "abc-123-def-456",
  "trace_id": "1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p",
  "span_id": "7h8i9j0k1l2m3n4o5p6q",
  "method": "GET",
  "path": "/hello",
  "query": "",
  "ip": "127.0.0.1",
  "user_agent": "curl/7.68.0",
  "status_code": 200,
  "response_size": 25,
  "latency_ms": 123.456
}
```

---

## 🎯 خلاصه

### روش فعلی (بدون Central Logging):

1. ✅ لاگ‌ها در Docker logs هستند
2. ✅ مشاهده با `docker logs`
3. ✅ فیلتر با `grep` یا `jq`
4. ❌ جستجو سخت است
5. ❌ visualization نیست

### دستورات مفید:

```bash
# مشاهده لاگ‌های زنده
docker logs -f go-backend-api

# جستجو بر اساس correlation_id
docker logs go-backend-api 2>&1 | grep "correlation_id"

# فقط errors
docker logs go-backend-api 2>&1 | jq 'select(.level == "error")'

# ذخیره در فایل
docker logs go-backend-api > logs.txt
```

---

## 📚 مستندات بیشتر

- **[LOKI_GUIDE.md](LOKI_GUIDE.md)**: راهنمای کامل استفاده از Loki
- **[OBSERVABILITY.md](OBSERVABILITY.md)**: راهنمای کامل Observability

