# 📋 راهنمای کامل Loki - Central Logging

## 📍 مقدمه

**Loki** یک log aggregation system است که:
- ✅ لاگ‌ها را از Docker containers جمع می‌کند
- ✅ در Grafana قابل مشاهده است
- ✅ جستجو و فیلتر آسان است
- ✅ با Tempo و Prometheus یکپارچه است

**Promtail** یک log collector است که:
- ✅ لاگ‌ها را از Docker containers می‌خواند
- ✅ آن‌ها را parse می‌کند (JSON)
- ✅ Labels را extract می‌کند
- ✅ لاگ‌ها را به Loki ارسال می‌کند

---

## 🚀 راه‌اندازی

### روش 1: راه‌اندازی کامل Observability Stack

```bash
# راه‌اندازی تمام stack (شامل Loki)
make observability-up
```

این دستور تمام سرویس‌ها را راه‌اندازی می‌کند:
- Tempo
- Jaeger
- Prometheus
- **Loki**
- **Promtail**
- Grafana

### روش 2: راه‌اندازی فقط Loki

```bash
# راه‌اندازی Loki و Promtail
make loki-up

# راه‌اندازی Grafana (برای مشاهده logs)
make grafana-up
```

---

## 📊 مشاهده Logs در Grafana

### مرحله 1: باز کردن Grafana Explore

1. باز کردن: http://localhost:3000
2. در منوی سمت چپ، روی **"Explore"** کلیک کنید (آیکون قطب‌نما)

### مرحله 2: انتخاب Loki Datasource

1. در بالای صفحه، dropdown **"Data source"** را باز کنید
2. **"Loki"** را انتخاب کنید

### مرحله 3: جستجوی Logs

#### روش 1: جستجوی ساده با Labels

در فیلد query، یکی از این query ها را بنویسید:

```logql
# تمام لاگ‌های container go-backend-api
{container="go-backend-api"}

# فقط error logs
{container="go-backend-api", level="error"}

# لاگ‌های یک correlation_id خاص
{container="go-backend-api", correlation_id="abc-123-def-456"}

# لاگ‌های یک trace_id خاص
{container="go-backend-api", trace_id="1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p"}

# لاگ‌های یک endpoint خاص
{container="go-backend-api", path="/hello"}

# لاگ‌های با status code 500
{container="go-backend-api", status_code="500"}
```

سپس روی **"Run query"** کلیک کنید.

#### روش 2: جستجوی پیشرفته با LogQL

```logql
# فیلتر بر اساس message
{container="go-backend-api"} |= "error"

# فیلتر بر اساس regex
{container="go-backend-api"} |~ ".*error.*"

# فقط error و warn logs
{container="go-backend-api"} | json | level=~"error|warn"

# لاگ‌های با latency بالا
{container="go-backend-api"} | json | latency_ms > 1000

# لاگ‌های یک method خاص
{container="go-backend-api"} | json | method="GET"

# ترکیب چند شرط
{container="go-backend-api"} | json | level="error" and status_code="500"
```

---

## 🔍 مثال‌های عملی

### مثال 1: پیدا کردن Error Logs

```logql
{container="go-backend-api", level="error"}
```

یا:

```logql
{container="go-backend-api"} | json | level="error"
```

### مثال 2: Debug کردن یک Request

```bash
# 1. ارسال request با correlation ID
CORRELATION_ID="debug-$(date +%s)"
curl -H "X-Correlation-ID: $CORRELATION_ID" http://localhost:8080/hello
```

سپس در Grafana:

```logql
{container="go-backend-api", correlation_id="debug-1234567890"}
```

### مثال 3: پیدا کردن Slow Requests

```logql
{container="go-backend-api"} | json | latency_ms > 1000
```

### مثال 4: پیدا کردن لاگ‌های یک Trace

اگر Trace ID را می‌دانید:

```logql
{container="go-backend-api", trace_id="1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p"}
```

### مثال 5: پیدا کردن لاگ‌های یک Endpoint

```logql
{container="go-backend-api", path="/hello"}
```

### مثال 6: Rate of Errors

```logql
sum(rate({container="go-backend-api", level="error"}[5m]))
```

---

## 🔗 Trace-to-Logs (از Trace به Logs)

یکی از ویژگی‌های قدرتمند Grafana، امکان رفتن از trace به logs مرتبط است.

### روش 1: از Tempo Explore

1. در Grafana Explore، Tempo datasource را انتخاب کنید
2. یک trace را پیدا کنید
3. روی trace کلیک کنید
4. در بخش **"Logs"**، روی **"Show logs"** کلیک کنید
5. لاگ‌های مرتبط با این trace نمایش داده می‌شوند

### روش 2: از Trace ID

اگر Trace ID را می‌دانید:

1. در Grafana Explore، Loki datasource را انتخاب کنید
2. Query بنویسید:
   ```logql
   {container="go-backend-api", trace_id="YOUR_TRACE_ID"}
   ```

---

## 📝 LogQL Syntax

### Basic Queries

```logql
# تمام لاگ‌ها
{container="go-backend-api"}

# با چند label
{container="go-backend-api", level="error", status_code="500"}
```

### Filtering

```logql
# فیلتر بر اساس message
{container="go-backend-api"} |= "error"

# فیلتر با regex
{container="go-backend-api"} |~ ".*error.*"

# فیلتر با JSON
{container="go-backend-api"} | json | level="error"
```

### Aggregations

```logql
# Count logs
count_over_time({container="go-backend-api"}[5m])

# Rate of logs
rate({container="go-backend-api"}[5m])

# Sum by label
sum by (level) (count_over_time({container="go-backend-api"}[5m]))
```

---

## 🎯 Labels موجود

Promtail به صورت خودکار این labels را از JSON logs extract می‌کند:

- `container`: نام container
- `container_name`: نام container بدون prefix
- `level`: log level (info, error, warn, etc.)
- `correlation_id`: correlation ID
- `trace_id`: trace ID
- `span_id`: span ID
- `method`: HTTP method (GET, POST, etc.)
- `path`: HTTP path
- `status_code`: HTTP status code
- `app`: نام application (go-backend-service)
- `service`: نام service (از docker-compose)
- `project`: نام project (از docker-compose)

---

## 🔧 Troubleshooting

### مشکل: هیچ Log ای نمایش داده نمی‌شود

**راه حل:**

1. بررسی کنید که Loki در حال اجرا است:
   ```bash
   docker ps | grep loki
   ```

2. بررسی کنید که Promtail در حال اجرا است:
   ```bash
   docker ps | grep promtail
   ```

3. بررسی logs Promtail:
   ```bash
   docker logs go-backend-promtail
   ```

4. بررسی logs Loki:
   ```bash
   docker logs go-backend-loki
   ```

5. بررسی کنید که container API در حال اجرا است:
   ```bash
   docker ps | grep go-backend-api
   ```

6. بررسی کنید که لاگ‌ها تولید می‌شوند:
   ```bash
   docker logs go-backend-api | tail -10
   ```

7. ارسال یک request تست:
   ```bash
   curl http://localhost:8080/hello
   ```

8. بررسی بازه زمانی:
   - اگر request را الان فرستادید، "Last 5 minutes" را انتخاب کنید

### مشکل: Labels نمایش داده نمی‌شوند

**راه حل:**

1. بررسی کنید که لاگ‌ها JSON format هستند:
   ```bash
   docker logs go-backend-api | head -1 | jq .
   ```

2. بررسی Promtail config:
   ```bash
   cat configs/promtail/promtail-config.yaml
   ```

3. Restart Promtail:
   ```bash
   docker-compose -f docker-compose.observability.yml restart promtail
   ```

### مشکل: Loki datasource پیدا نمی‌شود

**راه حل:**

1. بررسی کنید که Grafana container در حال اجرا است:
   ```bash
   docker ps | grep grafana
   ```

2. بررسی logs Grafana:
   ```bash
   docker logs go-backend-grafana | grep -i "loki\|datasource" | tail -20
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

## 📊 مثال‌های Query پیشرفته

### Rate of Errors per Minute

```logql
sum(rate({container="go-backend-api", level="error"}[1m]))
```

### Count of Logs by Level

```logql
sum by (level) (count_over_time({container="go-backend-api"}[5m]))
```

### Average Latency

```logql
avg_over_time({container="go-backend-api"} | json | latency_ms [5m])
```

### Top 10 Slowest Requests

```logql
topk(10, 
  sum by (path) (
    {container="go-backend-api"} | json | latency_ms > 1000
  )
)
```

### Error Rate by Endpoint

```logql
sum by (path) (
  rate({container="go-backend-api", level="error"}[5m])
)
```

---

## 🎯 خلاصه

### دستورات مفید:

```bash
# راه‌اندازی Loki
make loki-up

# مشاهده logs Loki
make loki-logs

# توقف Loki
make loki-down

# راه‌اندازی کامل stack
make observability-up
```

### Query های مفید:

```logql
# تمام لاگ‌ها
{container="go-backend-api"}

# فقط errors
{container="go-backend-api", level="error"}

# با correlation_id
{container="go-backend-api", correlation_id="abc-123"}

# با trace_id
{container="go-backend-api", trace_id="trace-id-here"}

# Slow requests
{container="go-backend-api"} | json | latency_ms > 1000
```

### لینک‌های مفید:

- **Grafana**: http://localhost:3000
- **Grafana Explore**: http://localhost:3000/explore
- **Loki API**: http://localhost:3100

---

## 💡 نکات مهم

1. ⚠️ **Loki خودش UI ندارد** - همیشه از Grafana استفاده کنید
2. ⚠️ **بازه زمانی مهم است** - اگر request را 1 ساعت پیش فرستادید، باید بازه زمانی را تغییر دهید
3. ✅ **Labels باید دقیق باشند** - از Explore می‌توانید labels موجود را ببینید
4. ✅ **LogQL syntax** - برای جستجوی پیشرفته از LogQL استفاده کنید
5. ✅ **Trace-to-Logs** - می‌توانید از trace به logs مرتبط بروید

