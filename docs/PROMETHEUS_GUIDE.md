# 📊 راهنمای مشاهده Metrics در Grafana با Prometheus

## 📍 مقدمه

این راهنما نحوه مشاهده و query کردن metrics در Grafana با استفاده از Prometheus datasource را توضیح می‌دهد.

---

## 🚀 راه‌اندازی

### 1. اطمینان از اجرای سرویس‌ها

```bash
# بررسی وضعیت containers
docker ps | grep -E "prometheus|api"

# باید این containers در حال اجرا باشند:
# - go-backend-prometheus
# - go-backend-api
```

### 2. بررسی Metrics Endpoint

```bash
# تست metrics endpoint
curl http://localhost:8080/metrics

# باید metrics را ببینید مثل:
# http_requests_total
# http_request_duration_seconds
# http_request_errors_total
```

---

## 📊 مشاهده Metrics در Grafana

### مرحله 1: باز کردن Grafana Explore

1. باز کردن: http://localhost:3000
2. در منوی سمت چپ، روی **"Explore"** کلیک کنید (آیکون قطب‌نما)

### مرحله 2: انتخاب Prometheus Datasource

1. در بالای صفحه، dropdown **"Data source"** را باز کنید
2. **"Prometheus"** را انتخاب کنید (default است)

### مرحله 3: Query کردن Metrics

#### روش 1: استفاده از Metrics Browser

1. در query input، روی **"Metrics browser"** کلیک کنید
2. لیست metrics را ببینید
3. یک metric را انتخاب کنید (مثلاً `http_requests_total`)
4. روی **"Run query"** کلیک کنید

#### روش 2: تایپ کردن Query

در query input، یکی از این query ها را بنویسید:

```promql
# تعداد کل درخواست‌ها
http_requests_total

# تعداد درخواست‌ها بر اساس path
http_requests_total{path="/hello"}

# تعداد درخواست‌ها بر اساس method
http_requests_total{method="GET"}

# تعداد درخواست‌ها بر اساس status code
http_requests_total{status_code="200"}

# Rate درخواست‌ها (درخواست در ثانیه)
rate(http_requests_total[5m])

# Rate درخواست‌ها بر اساس path
rate(http_requests_total{path="/hello"}[5m])

# Latency (میانگین)
http_request_duration_seconds

# Latency (95th percentile)
histogram_quantile(0.95, http_request_duration_seconds_bucket)

# تعداد errors
http_request_errors_total

# Rate errors
rate(http_request_errors_total[5m])
```

سپس روی **"Run query"** کلیک کنید.

---

## 🔍 مثال‌های Query مفید

### 1. تعداد کل درخواست‌ها

```promql
http_requests_total
```

### 2. Rate درخواست‌ها (درخواست در ثانیه)

```promql
rate(http_requests_total[5m])
```

### 3. Rate درخواست‌ها بر اساس Path

```promql
rate(http_requests_total{path="/hello"}[5m])
```

### 4. Latency (میانگین)

```promql
rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m])
```

### 5. Latency (95th percentile)

```promql
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

### 6. تعداد Errors

```promql
http_request_errors_total
```

### 7. Rate Errors

```promql
rate(http_request_errors_total[5m])
```

### 8. Error Rate (درصد)

```promql
rate(http_request_errors_total[5m]) / rate(http_requests_total[5m]) * 100
```

### 9. درخواست‌ها بر اساس Status Code

```promql
sum by (status_code) (rate(http_requests_total[5m]))
```

### 10. درخواست‌ها بر اساس Method

```promql
sum by (method) (rate(http_requests_total[5m]))
```

---

## 📈 Visualization Types

### Graph (خطی)

برای نمایش trends و patterns:

```promql
rate(http_requests_total[5m])
```

### Table

برای نمایش مقادیر دقیق:

```promql
http_requests_total
```

### Stat

برای نمایش یک مقدار واحد:

```promql
sum(rate(http_requests_total[5m]))
```

---

## 🎯 مثال عملی: ایجاد Dashboard

### مرحله 1: ارسال چند Request

```bash
# ارسال چند request
for i in {1..10}; do
  curl http://localhost:8080/hello
  sleep 1
done
```

### مرحله 2: Query در Grafana

1. باز کردن Grafana Explore
2. انتخاب Prometheus datasource
3. Query:
   ```promql
   rate(http_requests_total{path="/hello"}[5m])
   ```
4. روی **"Run query"** کلیک کنید
5. باید graph را ببینید

### مرحله 3: اضافه کردن به Dashboard

1. روی **"Add to dashboard"** کلیک کنید
2. Dashboard جدید یا موجود را انتخاب کنید
3. Panel را ذخیره کنید

---

## 🔧 Troubleshooting

### مشکل: "No data" نمایش داده می‌شود

**راه حل:**

1. بررسی کنید که Prometheus در حال اجرا است:
   ```bash
   docker ps | grep prometheus
   ```

2. بررسی کنید که API در حال اجرا است:
   ```bash
   docker ps | grep go-backend-api
   ```

3. بررسی کنید که metrics endpoint کار می‌کند:
   ```bash
   curl http://localhost:8080/metrics
   ```

4. بررسی Prometheus targets:
   - باز کردن: http://localhost:9090
   - رفتن به **Status** > **Targets**
   - بررسی وضعیت `go-backend-service` target

5. ارسال یک request:
   ```bash
   curl http://localhost:8080/hello
   ```

6. بررسی بازه زمانی:
   - اگر request را الان فرستادید، "Last 5 minutes" را انتخاب کنید

### مشکل: Metrics نمایش داده نمی‌شوند

**راه حل:**

1. بررسی Prometheus config:
   ```bash
   cat configs/prometheus.yml
   ```

2. بررسی scrape interval:
   - Prometheus هر 15 ثانیه scrape می‌کند
   - ممکن است نیاز باشد چند ثانیه صبر کنید

3. بررسی labels:
   - از Metrics browser استفاده کنید تا ببینید چه labels موجود هستند

### مشکل: Query syntax error

**راه حل:**

1. بررسی syntax PromQL
2. استفاده از Metrics browser برای پیدا کردن metric names
3. استفاده از **Explain** toggle برای دیدن query details

---

## 📝 Metrics موجود

### HTTP Request Metrics

- `http_requests_total`: تعداد کل درخواست‌ها
- `http_request_duration_seconds`: زمان پاسخ درخواست‌ها
- `http_request_errors_total`: تعداد errors
- `http_request_size_bytes`: اندازه request
- `http_response_size_bytes`: اندازه response

### Labels

- `method`: HTTP method (GET, POST, etc.)
- `path`: HTTP path (/hello, /delayed-hello, etc.)
- `status_code`: HTTP status code (200, 404, 500, etc.)

---

## 🎯 Query های پیشنهادی برای Dashboard

### 1. Request Rate

```promql
sum(rate(http_requests_total[5m]))
```

### 2. Request Rate by Path

```promql
sum by (path) (rate(http_requests_total[5m]))
```

### 3. Error Rate

```promql
sum(rate(http_request_errors_total[5m]))
```

### 4. Latency (P95)

```promql
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

### 5. Requests by Status Code

```promql
sum by (status_code) (rate(http_requests_total[5m]))
```

---

## 💡 نکات مهم

1. ⚠️ **بازه زمانی مهم است** - اگر request را 1 ساعت پیش فرستادید، باید بازه زمانی را تغییر دهید
2. ✅ **Rate functions** - برای نمایش trends از `rate()` استفاده کنید
3. ✅ **Labels** - از labels برای فیلتر کردن استفاده کنید
4. ✅ **Metrics browser** - برای پیدا کردن metric names از Metrics browser استفاده کنید
5. ✅ **Explain** - برای دیدن query details از Explain toggle استفاده کنید

---

## 🔗 لینک‌های مفید

- **Grafana**: http://localhost:3000
- **Grafana Explore**: http://localhost:3000/explore
- **Prometheus UI**: http://localhost:9090
- **Prometheus Targets**: http://localhost:9090/targets

---

## 📚 منابع بیشتر

- [PromQL Documentation](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Grafana Explore Documentation](https://grafana.com/docs/grafana/latest/explore/)

