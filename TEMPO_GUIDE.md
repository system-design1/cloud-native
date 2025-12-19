# 🔍 راهنمای مشاهده Traces در Tempo با Grafana

## 📋 فهرست مطالب

- [مقدمه](#مقدمه)
- [پیش‌نیازها](#پیشنیازها)
- [فعال‌سازی ارسال Traces به Tempo](#فعالسازی-ارسال-traces-به-tempo)
- [مشاهده Traces در Grafana](#مشاهده-traces-در-grafana)
- [روش‌های جستجوی Traces](#روشهای-جستجوی-traces)
- [مشکلات رایج](#مشکلات-رایج)

---

## مقدمه

**Tempo** یک distributed tracing backend است که traces را ذخیره می‌کند. **Tempo خودش UI ندارد** و برای مشاهده traces باید از **Grafana** استفاده کنید.

---

## پیش‌نیازها

### 1. اطمینان از اجرای سرویس‌ها

```bash
# بررسی وضعیت containers
docker ps | grep -E "tempo|grafana|api"

# باید این containers در حال اجرا باشند:
# - go-backend-tempo
# - go-backend-grafana
# - go-backend-api
```

### 2. دسترسی به Grafana

- **URL**: http://localhost:3000
- **Username**: `admin` (یا anonymous اگر فعال باشد)
- **Password**: `admin`

---

## فعال‌سازی ارسال Traces به Tempo

### ⚠️ نکته مهم

در حال حاضر، traces به **Jaeger** ارسال می‌شوند، نه Tempo. برای ارسال به Tempo باید تنظیمات را تغییر دهید.

### روش 1: استفاده از Environment Variables

#### برای Docker Compose:

فایل `.env` را ویرایش کنید:

```env
# فعال کردن Tempo
OTEL_TEMPO_ENABLED=true
OTEL_TEMPO_ENDPOINT=tempo:4318

# غیرفعال کردن Jaeger (اختیاری)
OTEL_JAEGER_ENABLED=false
```

سپس container را restart کنید:

```bash
docker-compose restart api
```

#### برای Local Development:

```bash
export OTEL_TEMPO_ENABLED=true
export OTEL_TEMPO_ENDPOINT=localhost:4318
export OTEL_JAEGER_ENABLED=false

make dev-run
```

### روش 2: تغییر در docker-compose.yml

فایل `docker-compose.yml` را ویرایش کنید:

```yaml
api:
  environment:
    OTEL_TEMPO_ENABLED: "true"
    OTEL_TEMPO_ENDPOINT: "tempo:4318"
    OTEL_JAEGER_ENABLED: "false"
```

سپس rebuild کنید:

```bash
docker-compose up -d --force-recreate api
```

---

## مشاهده Traces در Grafana

### مرحله 1: باز کردن Grafana

1. باز کردن http://localhost:3000
2. Login با `admin` / `admin` (یا anonymous access)

### مرحله 2: رفتن به Explore

1. در منوی سمت چپ، روی **"Explore"** کلیک کنید (آیکون قطب‌نما)
2. یا از آدرس: http://localhost:3000/explore

### مرحله 3: انتخاب Tempo Datasource

1. در بالای صفحه، dropdown **"Data source"** را باز کنید
2. **"Tempo"** را انتخاب کنید

### مرحله 4: جستجوی Traces

#### روش 1: جستجو با Service Name

1. در تب **"Search"** (نه TraceQL)
2. در فیلد **"Service name"**، `go-backend-service` را انتخاب کنید
3. روی **"Run query"** کلیک کنید

#### روش 2: جستجو با TraceQL (پیشرفته)

1. در تب **"TraceQL"**
2. Query مثال:
   ```
   {.service.name="go-backend-service"}
   ```
3. روی **"Run query"** کلیک کنید

#### روش 3: جستجو با Trace ID

اگر Trace ID را می‌دانید (از logs):

1. در تب **"Search"**
2. در فیلد **"Trace ID"**، Trace ID را وارد کنید
3. روی **"Run query"** کلیک کنید

---

## روش‌های جستجوی Traces

### 1. جستجو بر اساس Service Name

```
Service name: go-backend-service
```

### 2. جستجو بر اساس Operation Name

```
Operation: GET /hello
```

### 3. جستجو بر اساس Tags

```
Tags: http.method=GET
Tags: http.status_code=200
```

### 4. جستجو با TraceQL

#### مثال‌های TraceQL:

```traceql
# تمام traces از یک service
{.service.name="go-backend-service"}

# traces با status code 200
{.http.status_code="200"}

# traces با method GET
{.http.method="GET"}

# ترکیب چند شرط
{.service.name="go-backend-service" && .http.method="GET"}

# جستجو بر اساس path
{.http.url="/hello"}
```

### 5. جستجو بر اساس زمان

- در بالای صفحه، بازه زمانی را انتخاب کنید (مثلاً "Last 5 minutes")
- یا بازه زمانی سفارشی انتخاب کنید

---

## مثال عملی: مشاهده Trace یک Request

### مرحله 1: ارسال Request

```bash
# ارسال request
curl http://localhost:8080/hello

# یا با delay
curl http://localhost:8080/delayed-hello
```

### مرحله 2: پیدا کردن Trace ID از Logs

```bash
# مشاهده logs
docker logs go-backend-api 2>&1 | grep "trace_id" | tail -1

# خروجی مثال:
# "trace_id":"1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p"
```

### مرحله 3: جستجو در Grafana

1. باز کردن Grafana Explore
2. انتخاب Tempo datasource
3. در تب **"Search"**:
   - **Service name**: `go-backend-service`
   - **Time range**: "Last 5 minutes"
4. کلیک روی **"Run query"**

### مرحله 4: مشاهده Trace Details

1. روی یک trace از لیست کلیک کنید
2. مشاهده اطلاعات:
   - **Trace ID**: شناسه یکتا
   - **Duration**: زمان کل request
   - **Spans**: لیست spans (operations)
   - **Tags**: metadata (method, URL, status code, etc.)

---

## مشکلات رایج

### مشکل 1: هیچ Trace ای نمایش داده نمی‌شود

**راه حل:**

1. بررسی کنید که Tempo فعال است:
   ```bash
   docker logs go-backend-tempo | tail -20
   ```

2. بررسی کنید که `OTEL_TEMPO_ENABLED=true`:
   ```bash
   docker exec go-backend-api env | grep OTEL_TEMPO
   ```

3. بررسی کنید که traces ارسال می‌شوند:
   ```bash
   docker logs go-backend-api 2>&1 | grep -i "tempo\|trace" | tail -10
   ```

4. ارسال یک request تست:
   ```bash
   curl http://localhost:8080/hello
   ```

5. بررسی Tempo API:
   ```bash
   curl http://localhost:3200/api/search?limit=10
   ```

### مشکل 2: "No data" در Grafana

**راه حل:**

1. بررسی کنید که Tempo datasource درست تنظیم شده:
   - Grafana > Configuration > Data sources > Tempo
   - URL باید: `http://tempo:3200`

2. بررسی کنید که بازه زمانی درست است:
   - اگر request را الان فرستادید، "Last 5 minutes" را انتخاب کنید

3. بررسی Service name:
   - باید دقیقاً `go-backend-service` باشد (از logs بررسی کنید)

### مشکل 3: Traces فقط در Jaeger نمایش داده می‌شوند

**راه حل:**

این یعنی traces به Jaeger ارسال می‌شوند، نه Tempo. باید:

1. `OTEL_TEMPO_ENABLED=true` تنظیم کنید
2. `OTEL_JAEGER_ENABLED=false` تنظیم کنید
3. Container را restart کنید

### مشکل 4: Tempo datasource پیدا نمی‌شود

**راه حل:**

1. بررسی کنید که Grafana container در حال اجرا است:
   ```bash
   docker ps | grep grafana
   ```

2. بررسی logs Grafana:
   ```bash
   docker logs go-backend-grafana | grep -i "tempo\|datasource" | tail -20
   ```

3. بررسی فایل provisioning:
   ```bash
   cat configs/grafana/provisioning/datasources/datasources.yml
   ```

4. Restart Grafana:
   ```bash
   docker-compose -f docker-compose.observability.yml restart grafana
   ```

---

## دستورات مفید

### بررسی وضعیت Tempo

```bash
# Health check
curl http://localhost:3200/ready

# جستجوی traces
curl "http://localhost:3200/api/search?limit=10"

# دریافت trace با ID
curl "http://localhost:3200/api/traces/{trace-id}"
```

### مشاهده Logs

```bash
# Logs Tempo
docker logs -f go-backend-tempo

# Logs Grafana
docker logs -f go-backend-grafana

# Logs API (برای دیدن trace_id)
docker logs -f go-backend-api | grep trace_id
```

### تست ارسال Traces

```bash
# ارسال چند request
for i in {1..5}; do
  curl http://localhost:8080/hello
  sleep 1
done

# سپس در Grafana Explore جستجو کنید
```

---

## خلاصه مراحل

1. ✅ اطمینان از اجرای Tempo و Grafana
2. ✅ فعال کردن `OTEL_TEMPO_ENABLED=true`
3. ✅ تنظیم `OTEL_TEMPO_ENDPOINT=tempo:4318`
4. ✅ Restart API container
5. ✅ ارسال request تست
6. ✅ باز کردن Grafana Explore
7. ✅ انتخاب Tempo datasource
8. ✅ جستجو با Service name: `go-backend-service`
9. ✅ مشاهده traces!

---

## لینک‌های مفید

- **Grafana**: http://localhost:3000
- **Grafana Explore**: http://localhost:3000/explore
- **Tempo API**: http://localhost:3200
- **Jaeger UI**: http://localhost:16686 (اگر از Jaeger استفاده می‌کنید)

---

## نکات مهم

1. ⚠️ **Tempo خودش UI ندارد** - همیشه از Grafana استفاده کنید
2. ⚠️ **Traces باید به Tempo ارسال شوند** - اگر فقط Jaeger فعال است، traces در Tempo نخواهید دید
3. ⚠️ **بازه زمانی مهم است** - اگر request را 1 ساعت پیش فرستادید، باید بازه زمانی را تغییر دهید
4. ✅ **Service name باید دقیق باشد** - از logs بررسی کنید که دقیقاً چه نامی استفاده می‌شود

