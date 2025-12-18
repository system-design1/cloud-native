# راهنمای Observability (OpenTelemetry, Tempo, Prometheus, Grafana)

این سند شامل راهنمای کامل برای اجرا و استفاده از ابزارهای observability در پروژه است.

## فهرست مطالب

- [OpenTelemetry Tracing](#opentelemetry-tracing)
- [Tempo - Distributed Tracing](#tempo---distributed-tracing)
- [Prometheus - Metrics Collection](#prometheus---metrics-collection)
- [Grafana - Visualization](#grafana---visualization)
- [نمایش Logs و Traces](#نمایش-logs-و-traces)
- [مثال‌های کاربردی](#مثال‌های-کاربردی)

---

## OpenTelemetry Tracing

OpenTelemetry برای tracing درخواست‌های HTTP و ایجاد spans استفاده می‌شود.

### تنظیمات Environment Variables

```env
# فعال/غیرفعال کردن OpenTelemetry tracing
OTEL_TRACING_ENABLED=true

# نام سرویس
OTEL_SERVICE_NAME=go-backend-service

# نسخه سرویس
OTEL_SERVICE_VERSION=1.0.0

# فعال کردن Tempo exporter
OTEL_TEMPO_ENABLED=true

# آدرس Tempo endpoint (برای ارسال traces)
OTEL_TEMPO_ENDPOINT=tempo:4318
```

### اجرای OpenTelemetry

OpenTelemetry به صورت خودکار در کد فعال است. فقط کافی است متغیرهای محیطی را تنظیم کنید:

```bash
# برای اجرای محلی با Tempo
export OTEL_TRACING_ENABLED=true
export OTEL_TEMPO_ENABLED=true
export OTEL_TEMPO_ENDPOINT=localhost:4318
make run
```

---

## Tempo - Distributed Tracing

Tempo یک backend برای ذخیره و query کردن distributed traces است.

### اجرای Tempo با Makefile

```bash
# راه‌اندازی Tempo و Jaeger UI
make tempo-up

# توقف Tempo
make tempo-down

# مشاهده لاگ‌های Tempo
docker-compose -f docker-compose.observability.yml logs -f tempo
```

### دسترسی به رابط‌های کاربری

پس از اجرای `make tempo-up`:

- **Jaeger UI**: http://localhost:16686
  - برای مشاهده و جستجوی traces
  - قابلیت جستجو بر اساس service name، operation name، tags و غیره
  - **نکته**: Jaeger با memory storage اجرا می‌شود. برای مشاهده traces از Tempo، از Grafana استفاده کنید.

- **Tempo API**: http://localhost:3200
  - API endpoint برای query کردن traces (نه UI)
  - **نکته مهم**: Tempo خودش UI ندارد! برای مشاهده traces از **Grafana** استفاده کنید.
  - Endpoint های مفید:
    - `GET /api/search` - جستجوی traces
    - `GET /api/traces/{traceID}` - دریافت trace با ID
    - `GET /ready` - بررسی وضعیت Tempo

- **Grafana**: http://localhost:3000 (بهترین گزینه برای مشاهده traces از Tempo)
  - Tempo datasource از پیش تنظیم شده است
  - می‌توانید traces را در Grafana مشاهده کنید

### تنظیمات Tempo

فایل تنظیمات Tempo در `configs/tempo.yaml` قرار دارد.

پورت‌های مهم:
- **4317**: OTLP gRPC receiver
- **4318**: OTLP HTTP receiver (این پورت را در `OTEL_TEMPO_ENDPOINT` استفاده کنید)
- **3200**: Tempo API

### مثال: ارسال Traces به Tempo

1. تنظیم environment variables:
```bash
export OTEL_TRACING_ENABLED=true
export OTEL_TEMPO_ENABLED=true
export OTEL_TEMPO_ENDPOINT=localhost:4318
```

2. اجرای Tempo:
```bash
make tempo-up
```

3. اجرای برنامه:
```bash
make run
```

4. ارسال درخواست‌های تست:
```bash
curl http://localhost:8080/hello
curl http://localhost:8080/delayed-hello
```

5. مشاهده traces:
   - **روش 1 (توصیه می‌شود)**: استفاده از Grafana
     - باز کردن http://localhost:3000
     - رفتن به "Explore" (منوی سمت چپ)
     - انتخاب datasource: "Tempo"
     - جستجوی traces
   - **روش 2**: استفاده از Jaeger UI (فقط traces که مستقیماً به Jaeger ارسال می‌شوند)
     - باز کردن http://localhost:16686
     - انتخاب service: `go-backend-service`
     - کلیک روی "Find Traces"

---

## Prometheus - Metrics Collection

Prometheus برای جمع‌آوری metrics استفاده می‌شود.

### اجرای Prometheus با Makefile

```bash
# راه‌اندازی Prometheus
make prometheus-up

# توقف Prometheus
make prometheus-down

# مشاهده لاگ‌های Prometheus
docker-compose -f docker-compose.observability.yml logs -f prometheus
```

### دسترسی به Prometheus UI

پس از اجرای `make prometheus-up`:

- **Prometheus UI**: http://localhost:9090
  - برای مشاهده metrics
  - اجرای PromQL queries
  - مشاهده targets و scrape status

### تنظیمات Prometheus

فایل تنظیمات Prometheus در `configs/prometheus.yml` قرار دارد.

Prometheus به صورت خودکار metrics را از `/metrics` endpoint جمع‌آوری می‌کند.

### مثال: مشاهده Metrics در Prometheus

1. اجرای Prometheus:
```bash
make prometheus-up
```

2. اجرای برنامه (که `/metrics` endpoint را expose می‌کند):
```bash
make run
```

3. مشاهده metrics در Prometheus UI:
   - باز کردن http://localhost:9090
   - رفتن به "Status" > "Targets" برای بررسی scrape status
   - رفتن به "Graph" و اجرای queries مثل:
     - `http_requests_total` - تعداد کل درخواست‌ها
     - `http_request_duration_seconds` - زمان پاسخ درخواست‌ها

---

## Grafana - Visualization

Grafana برای visualization و dashboards استفاده می‌شود.

### اجرای Grafana با Makefile

```bash
# راه‌اندازی Grafana
make grafana-up

# توقف Grafana
make grafana-down

# مشاهده لاگ‌های Grafana
docker-compose -f docker-compose.observability.yml logs -f grafana
```

### دسترسی به Grafana UI

پس از اجرای `make grafana-up`:

- **Grafana UI**: http://localhost:3000
  - Username: `admin`
  - Password: `admin`

### Datasources در Grafana

Grafana به صورت خودکار با datasources زیر تنظیم شده است:
- **Prometheus**: http://prometheus:9090
- **Tempo**: http://tempo:3200

---

## نمایش Logs و Traces

### مشاهده Logs با Structured JSON

تمام لاگ‌ها به صورت structured JSON با Zerolog تولید می‌شوند.

#### مثال Log Output:

```json
{
  "level": "info",
  "timestamp": 1698765432,
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
  "latency_ms": 1234567890,
  "message": "HTTP request/response"
}
```

#### مشاهده Logs:

```bash
# مشاهده logs در ترمینال (برای اجرای محلی)
make run

# مشاهده logs از Docker container
docker logs -f go-backend-api

# یا با docker-compose
docker-compose logs -f api
```

#### مثال: فیلتر کردن Logs بر اساس Correlation ID

```bash
# مشاهده logs با correlation ID خاص
docker logs go-backend-api 2>&1 | grep "abc-123-def-456"

# مشاهده logs با trace ID خاص
docker logs go-backend-api 2>&1 | grep "1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p"
```

### مشاهده Traces در Jaeger

#### مثال: پیدا کردن Trace با Correlation ID

1. اجرای درخواست با Correlation ID:
```bash
curl -H "X-Correlation-ID: test-correlation-123" http://localhost:8080/hello
```

2. باز کردن Jaeger UI: http://localhost:16686

3. جستجوی trace:
   - انتخاب service: `go-backend-service`
   - در بخش "Tags" کلیک کنید و `correlation_id=test-correlation-123` را اضافه کنید
   - کلیک روی "Find Traces"

#### مثال: مشاهده Trace Details

1. در Jaeger UI، روی یک trace کلیک کنید
2. مشاهده اطلاعات:
   - **Trace ID**: شناسه یکتا برای کل request
   - **Span ID**: شناسه یکتا برای هر operation
   - **Duration**: زمان اجرای operation
   - **Tags**: شامل HTTP method، URL، status code، correlation ID و غیره
   - **Logs**: شامل log entries مرتبط با span

---

## مثال‌های کاربردی

### مثال 1: اجرای کامل Stack (API + Observability)

```bash
# 1. راه‌اندازی observability stack
make observability-up

# 2. تنظیم environment variables برای Tempo
export OTEL_TRACING_ENABLED=true
export OTEL_TEMPO_ENABLED=true
export OTEL_TEMPO_ENDPOINT=localhost:4318

# 3. راه‌اندازی API
make docker-up

# 4. ارسال درخواست‌های تست
curl http://localhost:8080/hello
curl http://localhost:8080/delayed-hello

# 5. مشاهده traces در Jaeger
# باز کردن: http://localhost:16686

# 6. مشاهده metrics در Prometheus
# باز کردن: http://localhost:9090

# 7. مشاهده dashboards در Grafana
# باز کردن: http://localhost:3000
```

### مثال 2: Debug کردن Request با Correlation ID

```bash
# 1. ارسال request با correlation ID سفارشی
CORRELATION_ID="debug-request-$(date +%s)"
curl -H "X-Correlation-ID: ${CORRELATION_ID}" \
     http://localhost:8080/delayed-hello

# 2. مشاهده logs با correlation ID
docker logs go-backend-api 2>&1 | grep "${CORRELATION_ID}"

# 3. پیدا کردن trace در Jaeger
# باز کردن Jaeger UI و جستجو با correlation_id=${CORRELATION_ID}
```

### مثال 3: مشاهده Metrics برای یک Endpoint خاص

```bash
# 1. ارسال چند درخواست
for i in {1..10}; do
  curl http://localhost:8080/hello
  sleep 1
done

# 2. باز کردن Prometheus UI: http://localhost:9090

# 3. اجرای query برای مشاهده rate درخواست‌ها:
rate(http_requests_total{path="/hello"}[5m])

# 4. اجرای query برای مشاهده latency:
histogram_quantile(0.95, http_request_duration_seconds_bucket{path="/hello"})
```

### مثال 4: مشاهده Trace Flow برای یک Request

1. ارسال request:
```bash
curl http://localhost:8080/delayed-hello
```

2. در Jaeger UI (http://localhost:16686):
   - پیدا کردن trace برای `go-backend-service`
   - مشاهده span hierarchy
   - مشاهده timing برای هر span
   - مشاهده tags شامل:
     - `http.method`: GET
     - `http.url`: /delayed-hello
     - `http.status_code`: 200
     - `correlation_id`: UUID
     - `http.response.size`: اندازه پاسخ

---

## Troubleshooting

### مشکل: Traces در Jaeger نمایش داده نمی‌شوند

**راه حل:**
1. بررسی کنید که Tempo در حال اجرا است:
```bash
docker ps | grep tempo
```

2. بررسی کنید که `OTEL_TEMPO_ENABLED=true` تنظیم شده است:
```bash
docker logs go-backend-api | grep -i "tracing"
```

3. بررسی کنید که endpoint درست است:
```bash
# برای Docker Compose: tempo:4318
# برای local: localhost:4318
```

### مشکل: Prometheus metrics scrape نمی‌شوند

**راه حل:**
1. بررسی کنید که Prometheus در حال اجرا است:
```bash
docker ps | grep prometheus
```

2. بررسی کنید که `/metrics` endpoint در دسترس است:
```bash
curl http://localhost:8080/metrics
```

3. در Prometheus UI به "Status" > "Targets" بروید و وضعیت scrape را بررسی کنید

### مشکل: Logs ساختار JSON ندارند

**راه حل:**
1. بررسی کنید که `LOG_LEVEL` تنظیم شده است
2. بررسی کنید که Gin در debug mode نیست (Gin debug output ممکن است با JSON logs مخلوط شود)

---

## خلاصه دستورات Makefile

```bash
# Observability Stack
make observability-up        # راه‌اندازی تمام stack
make observability-down      # توقف تمام stack
make observability-logs      # مشاهده logs

# Tempo
make tempo-up                # راه‌اندازی Tempo + Jaeger
make tempo-down              # توقف Tempo

# Prometheus
make prometheus-up           # راه‌اندازی Prometheus
make prometheus-down         # توقف Prometheus

# Grafana
make grafana-up              # راه‌اندازی Grafana
make grafana-down            # توقف Grafana
```

---

## لینک‌های مفید

- **Jaeger UI**: http://localhost:16686
- **Prometheus UI**: http://localhost:9090
- **Grafana UI**: http://localhost:3000
- **Tempo API**: http://localhost:3200

