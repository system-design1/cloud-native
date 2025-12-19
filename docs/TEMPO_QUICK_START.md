# 🚀 راهنمای سریع: مشاهده Traces در Tempo

## ⚠️ مشکل فعلی

از logs مشخص است که:
- ✅ Tempo در حال اجرا است
- ✅ Grafana در حال اجرا است  
- ❌ اما traces به **Jaeger** ارسال می‌شوند، نه Tempo

## 🔧 راه حل: فعال‌سازی Tempo

### روش 1: تغییر در docker-compose.yml (توصیه می‌شود)

فایل `docker-compose.yml` را ویرایش کنید:

```yaml
api:
  environment:
    # فعال کردن Tempo
    OTEL_TEMPO_ENABLED: "true"
    OTEL_TEMPO_ENDPOINT: "tempo:4318"
    
    # غیرفعال کردن Jaeger (اختیاری - می‌توانید هر دو را فعال نگه دارید)
    OTEL_JAEGER_ENABLED: "false"
```

سپس container را restart کنید:

```bash
docker-compose restart api
```

### روش 2: استفاده از .env

فایل `.env` را ویرایش کنید:

```env
# فعال کردن Tempo
OTEL_TEMPO_ENABLED=true
OTEL_TEMPO_ENDPOINT=tempo:4318

# غیرفعال کردن Jaeger (اختیاری)
OTEL_JAEGER_ENABLED=false
```

سپس:

```bash
docker-compose restart api
```

---

## 📊 مشاهده Traces در Grafana

### مرحله 1: باز کردن Grafana Explore

1. باز کردن: http://localhost:3000
2. در منوی سمت چپ، روی **"Explore"** کلیک کنید (آیکون قطب‌نما)

### مرحله 2: انتخاب Tempo Datasource

1. در بالای صفحه، dropdown **"Data source"** را باز کنید
2. **"Tempo"** را انتخاب کنید

### مرحله 3: جستجوی Traces

#### روش ساده (Search):

1. در تب **"Search"** (نه TraceQL)
2. در فیلد **"Service name"**:
   - Dropdown را باز کنید
   - `go-backend-service` را انتخاب کنید
3. در **"Time range"** (بالای صفحه):
   - "Last 5 minutes" را انتخاب کنید
4. روی **"Run query"** کلیک کنید

#### اگر هیچ trace ای نمایش داده نشد:

1. یک request تست بفرستید:
   ```bash
   curl http://localhost:8080/hello
   ```

2. چند ثانیه صبر کنید

3. دوباره "Run query" را بزنید

### مرحله 4: مشاهده Trace Details

1. روی یک trace از لیست کلیک کنید
2. مشاهده اطلاعات:
   - **Trace ID**: شناسه یکتا
   - **Duration**: زمان کل request
   - **Spans**: لیست operations
   - **Tags**: metadata (method, URL, status code)

---

## 🔍 جستجو با TraceQL (پیشرفته)

در تب **"TraceQL"** می‌توانید query های پیشرفته بنویسید:

### مثال‌های TraceQL:

```traceql
# تمام traces از service
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

---

## ✅ بررسی اینکه Traces به Tempo ارسال می‌شوند

### 1. بررسی Logs:

```bash
docker logs go-backend-api 2>&1 | grep -i "tempo\|tracing" | tail -5
```

باید ببینید:
```
"tempo_enabled":true
```

### 2. تست ارسال Request:

```bash
# ارسال request
curl http://localhost:8080/hello

# بررسی logs برای trace_id
docker logs go-backend-api 2>&1 | grep "trace_id" | tail -1
```

### 3. بررسی Tempo API:

```bash
# جستجوی traces در Tempo
curl "http://localhost:3200/api/search?limit=10"
```

اگر traces وجود داشته باشد، JSON response دریافت می‌کنید.

---

## 🐛 Troubleshooting

### مشکل: هیچ Trace ای نمایش داده نمی‌شود

**راه حل:**

1. بررسی کنید که Tempo فعال است:
   ```bash
   docker exec go-backend-api env | grep OTEL_TEMPO
   ```
   باید ببینید: `OTEL_TEMPO_ENABLED=true`

2. بررسی logs:
   ```bash
   docker logs go-backend-api 2>&1 | grep "tempo_enabled"
   ```

3. ارسال request تست:
   ```bash
   curl http://localhost:8080/hello
   ```

4. بررسی بازه زمانی:
   - اگر request را الان فرستادید، "Last 5 minutes" را انتخاب کنید

5. بررسی Service name:
   - باید دقیقاً `go-backend-service` باشد

### مشکل: "No data" در Grafana

**راه حل:**

1. بررسی Tempo datasource:
   - Grafana > Configuration > Data sources > Tempo
   - URL باید: `http://tempo:3200`

2. Restart Grafana:
   ```bash
   docker-compose -f docker-compose.observability.yml restart grafana
   ```

---

## 📝 خلاصه مراحل

1. ✅ ویرایش `docker-compose.yml`:
   ```yaml
   OTEL_TEMPO_ENABLED: "true"
   OTEL_TEMPO_ENDPOINT: "tempo:4318"
   ```

2. ✅ Restart API:
   ```bash
   docker-compose restart api
   ```

3. ✅ ارسال Request تست:
   ```bash
   curl http://localhost:8080/hello
   ```

4. ✅ باز کردن Grafana Explore:
   - http://localhost:3000/explore

5. ✅ انتخاب Tempo datasource

6. ✅ جستجو با Service name: `go-backend-service`

7. ✅ مشاهده Traces! 🎉

---

## 🔗 لینک‌های مفید

- **Grafana**: http://localhost:3000
- **Grafana Explore**: http://localhost:3000/explore
- **Tempo API**: http://localhost:3200
- **Jaeger UI**: http://localhost:16686 (اگر از Jaeger استفاده می‌کنید)

---

## 💡 نکات مهم

1. ⚠️ **Tempo خودش UI ندارد** - همیشه از Grafana استفاده کنید
2. ⚠️ **بازه زمانی مهم است** - اگر request را 1 ساعت پیش فرستادید، باید بازه زمانی را تغییر دهید
3. ✅ **می‌توانید هم Tempo و هم Jaeger را فعال نگه دارید** - traces به هر دو ارسال می‌شوند

