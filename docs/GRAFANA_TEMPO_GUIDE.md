# راهنمای استفاده از Tempo در Grafana

## مشکل: فقط Metrics نمایش داده می‌شود

اگر در Grafana فقط metrics را می‌بینید و traces دیگر (مثل `/hello`, `/delayed-hello`) را نمی‌بینید، احتمالاً مشکل از query است.

## راه‌حل: استفاده از Query درست در Grafana

### مرحله 1: باز کردن Grafana Explore

1. باز کردن: http://localhost:3000
2. رفتن به **Explore** (منوی سمت چپ)
3. انتخاب **Tempo** datasource از dropdown

### مرحله 2: Query درست برای مشاهده همه Traces

در Grafana Explore با Tempo، می‌توانید از query های زیر استفاده کنید:

#### Query 1: همه Traces از Service
```
{resource.service.name="go-backend-service"}
```

#### Query 2: Traces خاص (مثل /hello)
```
{resource.service.name="go-backend-service" && name="GET /hello"}
```

#### Query 3: Traces با Status Code خاص
```
{resource.service.name="go-backend-service" && http.status_code=200}
```

#### Query 4: Traces بدون /metrics
```
{resource.service.name="go-backend-service" && name!="GET /metrics"}
```

### مرحله 3: استفاده از Search

در Grafana Explore:
1. در قسمت **Search**، می‌توانید:
   - Service name را انتخاب کنید: `go-backend-service`
   - Operation name را انتخاب کنید: `GET /hello`, `GET /delayed-hello`, etc.
   - Tags را اضافه کنید: `http.status_code=200`

2. **Lookback** را تنظیم کنید: `Last 15 minutes` یا بیشتر

3. **Limit Results** را افزایش دهید: `50` یا بیشتر

4. کلیک روی **Run query**

### مرحله 4: مشاهده Trace Details

بعد از پیدا کردن trace:
1. روی trace کلیک کنید
2. در قسمت پایین، می‌توانید:
   - Timeline را ببینید
   - Spans را ببینید
   - Attributes را ببینید
   - Logs را ببینید (اگر Loki متصل باشد)

## مشکل رایج: Traces قدیمی

اگر traces جدید را نمی‌بینید:
1. **Lookback** را افزایش دهید
2. چند request جدید بزنید:
   ```bash
   curl http://localhost:8080/hello
   curl http://localhost:8080/delayed-hello
   ```
3. در Grafana، روی **Run query** کلیک کنید

## بررسی Tempo API مستقیم

برای بررسی مستقیم Tempo API:

```bash
# همه traces
curl "http://localhost:3200/api/search?limit=20"

# Traces خاص
curl "http://localhost:3200/api/search?limit=20&q={resource.service.name=\"go-backend-service\"}"
```

## نکات مهم

1. **Traces با delay نمایش داده می‌شوند**: Tempo traces را در batches ارسال می‌کند، پس ممکن است چند ثانیه delay داشته باشد.

2. **Query syntax**: در Grafana، از syntax `{key="value"}` استفاده کنید.

3. **Service name**: Service name باید دقیقاً `go-backend-service` باشد.

4. **Operation name**: Operation name به صورت `METHOD /path` است (مثل `GET /hello`).

## مثال Query های مفید

```bash
# همه traces از service
{resource.service.name="go-backend-service"}

# فقط /hello
{resource.service.name="go-backend-service" && name="GET /hello"}

# فقط /delayed-hello
{resource.service.name="go-backend-service" && name="GET /delayed-hello"}

# فقط /test-error
{resource.service.name="go-backend-service" && name="GET /test-error"}

# بدون /metrics
{resource.service.name="go-backend-service" && name!="GET /metrics"}

# با status code 200
{resource.service.name="go-backend-service" && http.status_code=200}
```

## عیب‌یابی

اگر هنوز traces را نمی‌بینید:

1. **بررسی Tempo API**:
   ```bash
   curl "http://localhost:3200/api/search?limit=10"
   ```

2. **بررسی Application Logs**:
   ```bash
   docker-compose logs api | grep -i "trace\|tempo"
   ```

3. **بررسی Environment Variables**:
   ```bash
   docker-compose exec api printenv | grep OTEL
   ```

4. **بررسی Tempo Logs**:
   ```bash
   docker-compose -f docker-compose.observability.yml logs tempo | tail -20
   ```

