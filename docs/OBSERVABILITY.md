# راهنمای Observability (OpenTelemetry, Tempo, Prometheus, Loki, Grafana)

این سند شامل راهنمای کامل برای اجرا و استفاده از ابزارهای observability در پروژه است.

## فهرست مطالب

- [OpenTelemetry Tracing](#opentelemetry-tracing)
- [Tempo - Distributed Tracing](#tempo---distributed-tracing)
- [Prometheus - Metrics Collection](#prometheus---metrics-collection)
- [Loki - Log Aggregation](#loki---log-aggregation)
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

# Route-Based Tracing Policy (برای کنترل sampling)
# برای جزئیات کامل، به بخش "Route-Based Tracing Policy" مراجعه کنید
OTEL_ROUTE_POLICY_ENABLED=true
OTEL_ROUTE_ALWAYS=/delayed-hello,/test-error
OTEL_ROUTE_DROP=/metrics
OTEL_ROUTE_RATIO=/health=0.01,/live=0.01,/ready=0.01
OTEL_ROUTE_DEFAULT=always
OTEL_ROUTE_DEFAULT_RATIO=1.0
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

### Route-Based Tracing Policy

**Route-Based Tracing Policy** یک قابلیت پیشرفته برای کنترل sampling traces بر اساس route است. این قابلیت به شما امکان می‌دهد که:

- **کاهش noise در Jaeger/Tempo**: با drop کردن یا کاهش sampling برای endpoints پرترافیک مثل `/metrics` و `/health`
- **تمرکز روی traces مهم**: با always sampling برای endpoints مهم مثل `/delayed-hello` و `/test-error`
- **بهینه‌سازی هزینه**: با کاهش تعداد traces ارسالی به backend

#### چرا Route-Based Policy؟

برخی endpoints مثل `/metrics`، `/health`، `/ready` و `/live` بسیار پرترافیک هستند و trace کردن همه آن‌ها باعث:
- **Noise در Jaeger/Tempo**: پیدا کردن traces مهم سخت می‌شود
- **افزایش هزینه**: تعداد زیادی trace به backend ارسال می‌شود
- **کاهش کارایی**: پردازش و ذخیره‌سازی traces اضافی

با استفاده از Route-Based Policy، می‌توانید:
- `/metrics` را **DROP** کنید (هیچ trace ای sample نمی‌شود)
- `/health`، `/live`، `/ready` را با **RATIO 1%** sample کنید (فقط 1% از requests)
- `/delayed-hello` و `/test-error` را **ALWAYS** sample کنید (همیشه trace می‌شوند)

#### سه نوع Policy

##### 1. ALWAYS (همیشه trace می‌شود)

برای endpoints مهم که می‌خواهید همیشه trace شوند:

```env
OTEL_ROUTE_ALWAYS=/delayed-hello,/test-error
```

**مزایا:**
- همیشه trace می‌شود (100% sampling)
- برای debugging و demo مفید است
- برای endpoints مهم که می‌خواهید همیشه ببینید

##### 2. RATIO (با احتمال مشخص trace می‌شود)

برای endpoints پرترافیک که می‌خواهید گاهی trace شوند:

```env
OTEL_ROUTE_RATIO=/health=0.01,/live=0.01,/ready=0.01
```

**فرمت:** `path=ratio` (ratio باید بین `0.0` و `1.0` باشد)

**مثال‌ها:**
- `0.01` = 1% از requests
- `0.1` = 10% از requests
- `0.5` = 50% از requests
- `1.0` = 100% از requests (معادل ALWAYS)

**مزایا:**
- کاهش تعداد traces برای endpoints پرترافیک
- هنوز هم می‌توانید نمونه‌ای از traces را ببینید
- برای monitoring و debugging کافی است

##### 3. DROP (هرگز trace نمی‌شود)

برای endpoints پرترافیک که نمی‌خواهید trace شوند:

```env
OTEL_ROUTE_DROP=/metrics
```

**مزایا:**
- هیچ trace ای sample نمی‌شود (0% sampling)
- برای endpoints که trace کردن آن‌ها مفید نیست
- کاهش قابل توجه noise و هزینه

#### ترتیب اولویت (Precedence)

Policy ها به ترتیب زیر اعمال می‌شوند (اولی بالاترین اولویت را دارد):

1. **DROP** (بالاترین اولویت)
2. **ALWAYS**
3. **RATIO**
4. **DEFAULT** policy

**مثال:**
اگر یک route هم در `OTEL_ROUTE_DROP` و هم در `OTEL_ROUTE_ALWAYS` باشد، **DROP** اعمال می‌شود.

#### Default Policy

برای routes که در هیچ یک از لیست‌های بالا نیستند:

```env
OTEL_ROUTE_DEFAULT=always  # یا ratio یا drop
OTEL_ROUTE_DEFAULT_RATIO=1.0  # فقط برای OTEL_ROUTE_DEFAULT=ratio
```

**مقادیر ممکن:**
- `always`: همه traces را sample می‌کند (پیش‌فرض)
- `ratio`: از `OTEL_ROUTE_DEFAULT_RATIO` استفاده می‌کند
- `drop`: هیچ trace ای sample نمی‌کند

#### تنظیمات پیش‌فرض (Demo-friendly)

با تنظیمات پیش‌فرض:

```env
OTEL_ROUTE_POLICY_ENABLED=true
OTEL_ROUTE_ALWAYS=/delayed-hello,/test-error
OTEL_ROUTE_DROP=/metrics
OTEL_ROUTE_RATIO=/health=0.01,/live=0.01,/ready=0.01
OTEL_ROUTE_DEFAULT=always
OTEL_ROUTE_DEFAULT_RATIO=1.0
```

**نتیجه:**
- `/delayed-hello` و `/test-error`: **همیشه** trace می‌شوند
- `/health`، `/live`، `/ready`: **1%** از requests trace می‌شوند
- `/metrics`: **trace نمی‌شود** (DROP)
- سایر routes: **همیشه** trace می‌شوند (default)

#### غیرفعال کردن Policy

برای غیرفعال کردن policy و استفاده از رفتار پیش‌فرض (sample همه traces):

```env
OTEL_ROUTE_POLICY_ENABLED=false
```

**مزایا:**
- برای debugging مفید است
- می‌توانید همه traces را ببینید
- برای development و testing

#### نکات مهم

1. **Policy فقط زمانی اعمال می‌شود که `OTEL_ROUTE_POLICY_ENABLED=true` باشد**
   - وقتی `false` است، همه traces sample می‌شوند (رفتار پیش‌فرض)

2. **Routes با query string هم درست کار می‌کنند**
   - فقط path بررسی می‌شود، نه query string
   - مثال: `/health?check=1` و `/health?check=2` هر دو به `/health` map می‌شوند

3. **GET و HEAD هر دو پشتیبانی می‌شوند**
   - Policy بر اساس path اعمال می‌شود، نه method

4. **Parent-child consistency**
   - اگر parent span sample شده باشد، child span هم sample می‌شود
   - این برای حفظ integrity trace مهم است

5. **برای debugging، policy را غیرفعال کنید**
   - `OTEL_ROUTE_POLICY_ENABLED=false` تنظیم کنید
   - همه traces را ببینید

#### مثال‌های کاربردی

##### مثال 1: کاهش Noise برای Health Checks

```env
OTEL_ROUTE_POLICY_ENABLED=true
OTEL_ROUTE_RATIO=/health=0.01,/live=0.01,/ready=0.01
OTEL_ROUTE_DEFAULT=always
```

**نتیجه:** Health checks فقط 1% trace می‌شوند، سایر routes همیشه.

##### مثال 2: Drop کردن Metrics

```env
OTEL_ROUTE_POLICY_ENABLED=true
OTEL_ROUTE_DROP=/metrics
OTEL_ROUTE_DEFAULT=always
```

**نتیجه:** `/metrics` trace نمی‌شود، سایر routes همیشه.

##### مثال 3: Always برای Demo Endpoints

```env
OTEL_ROUTE_POLICY_ENABLED=true
OTEL_ROUTE_ALWAYS=/delayed-hello,/test-error
OTEL_ROUTE_DROP=/metrics
OTEL_ROUTE_RATIO=/health=0.01
OTEL_ROUTE_DEFAULT=ratio
OTEL_ROUTE_DEFAULT_RATIO=0.1
```

**نتیجه:**
- `/delayed-hello` و `/test-error`: همیشه
- `/metrics`: drop
- `/health`: 1%
- سایر routes: 10%

---

## Local Development با Observability

برای استفاده از observability در local development (`make dev-run`):

### Setup کامل

```bash
# Terminal 1: Database
make dev-db-up

# Terminal 2: Application با hot reload
make dev-run

# Terminal 3: Observability stack
make observability-up

# Terminal 4: Health checker (اختیاری)
make dev-health-checker
```

### چه چیزی به صورت خودکار کار می‌کند؟

1. **Prometheus**: به صورت خودکار `/metrics` را scrape می‌کند (هر 5 ثانیه)
   - Prometheus config به `host.docker.internal:8080` اشاره می‌کند
   - اگر application روی `localhost:8080` اجرا شود، Prometheus می‌تواند به آن دسترسی پیدا کند

2. **Health Checker**: به صورت خودکار `/health`, `/ready` و `/live` را call می‌کند (هر 10 ثانیه)
   - فقط اگر `make dev-health-checker` را اجرا کرده باشید
   - برای تنظیم interval: `HEALTH_CHECK_INTERVAL=5 make dev-health-checker`

### تفاوت با Docker

| ویژگی | Docker (`make docker-up`) | Local (`make dev-run`) |
|-------|---------------------------|------------------------|
| `/metrics` traces | ✅ خودکار (Prometheus scrape) | ✅ خودکار (Prometheus scrape) |
| `/health` traces | ✅ خودکار (Docker health check) | ⚠️ نیاز به `make dev-health-checker` |
| `/ready` traces | ✅ خودکار (Docker health check) | ⚠️ نیاز به `make dev-health-checker` |
| `/live` traces | ✅ خودکار (Docker health check) | ⚠️ نیاز به `make dev-health-checker` |

### نکات مهم

1. **Prometheus config**: به صورت خودکار برای local development تنظیم شده است (`host.docker.internal:8080`)
2. **Health Checker**: برای ایجاد traces خودکار برای health endpoints، باید `make dev-health-checker` را اجرا کنید
3. **Environment Variables**: در `.env` باید `OTEL_TEMPO_ENDPOINT=localhost:4318` و `OTEL_JAEGER_ENDPOINT=localhost:4320` باشد

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
# راه‌اندازی Prometheus (همراه با سایر observability tools)
make observability-up

# یا فقط Prometheus
make prometheus-up
```

### Prometheus در Local Development

وقتی از `make dev-run` استفاده می‌کنید (local development):

1. **Prometheus به صورت خودکار `/metrics` را scrape می‌کند**
   - Prometheus config به `host.docker.internal:8080` و `api:8080` اشاره می‌کند
   - اگر application روی `localhost:8080` اجرا شود، Prometheus می‌تواند به آن دسترسی پیدا کند
   - Scrape interval: هر 5 ثانیه

2. **Health Endpoints (`/health`, `/ready`, `/live`)**
   - این endpoints به صورت خودکار call نمی‌شوند (برخلاف Docker که health checks وجود دارند)
   - برای ایجاد traces خودکار، از `make dev-health-checker` استفاده کنید:
     ```bash
     # در یک terminal جداگانه
     make dev-health-checker
     ```
   - این script هر 10 ثانیه `/health`, `/ready` و `/live` را call می‌کند
   - برای تنظیم interval: `HEALTH_CHECK_INTERVAL=5 make dev-health-checker`

### تنظیمات Prometheus

فایل `configs/prometheus.yml` شامل تنظیمات زیر است:

```yaml
scrape_configs:
  - job_name: 'go-backend-service'
    static_configs:
      - targets: ['host.docker.internal:8080', 'api:8080']
    metrics_path: '/metrics'
    scrape_interval: 5s
```

**توضیحات:**
- `host.docker.internal:8080`: برای دسترسی به application از Docker container (local development)
- `api:8080`: برای دسترسی به application در Docker network (Docker Compose)
- `scrape_interval: 5s`: Prometheus هر 5 ثانیه metrics را scrape می‌کند

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

```yaml
scrape_configs:
  - job_name: 'go-backend-service'
    static_configs:
      - targets: ['host.docker.internal:8080', 'api:8080']
    metrics_path: '/metrics'
    scrape_interval: 5s
```

**توضیحات:**
- `host.docker.internal:8080`: برای دسترسی به application از Docker container (local development)
- `api:8080`: برای دسترسی به application در Docker network (Docker Compose)
- `scrape_interval: 5s`: Prometheus هر 5 ثانیه metrics را scrape می‌کند

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

4. **برای local development (`make dev-run`):**
   - بررسی کنید که Prometheus config به `host.docker.internal:8080` اشاره می‌کند
   - بررسی کنید که `extra_hosts` در `docker-compose.observability.yml` تنظیم شده است
   - در Prometheus UI، target `host.docker.internal:8080` باید status `UP` داشته باشد

### مشکل: Health endpoints traces ایجاد نمی‌شوند در Local Development

**راه حل:**
1. Health checker را اجرا کنید:
```bash
make dev-health-checker
```

2. یا به صورت دستی call کنید:
```bash
curl http://localhost:8080/health
curl http://localhost:8080/ready
curl http://localhost:8080/live
```

3. بررسی کنید که `OTEL_ROUTE_POLICY_ENABLED=false` باشد (برای sample شدن همه traces)

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

# Local Development با Observability
make dev-run                 # اجرای application با hot reload
make dev-health-checker      # اجرای health checker (برای ایجاد traces خودکار)
```

---

## لینک‌های مفید

- **Jaeger UI**: http://localhost:16686
- **Prometheus UI**: http://localhost:9090
- **Grafana UI**: http://localhost:3000
- **Tempo API**: http://localhost:3200

