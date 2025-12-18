# راهنمای استفاده از Tempo API

## نکته مهم
**Tempo یک UI ندارد!** پورت 3200 فقط برای API است. برای مشاهده traces از Grafana استفاده کنید.

## دسترسی به Grafana
- URL: http://localhost:3000
- Username: `admin`
- Password: `admin`
- Tempo datasource از پیش تنظیم شده است

## Tempo API Endpoints

### 1. بررسی وضعیت Tempo
```bash
curl http://localhost:3200/ready
```

### 2. جستجوی Traces
```bash
# جستجوی traces با service name
curl "http://localhost:3200/api/search?tags=service.name=go-backend-service"

# جستجوی traces با limit
curl "http://localhost:3200/api/search?limit=10"
```

### 3. دریافت Trace با ID
```bash
# دریافت trace با trace ID
curl "http://localhost:3200/api/traces/{traceID}"
```

### 4. دریافت Metrics
```bash
# دریافت metrics مربوط به traces
curl "http://localhost:3200/metrics"
```

## مثال: مشاهده Traces در Grafana

1. باز کردن Grafana: http://localhost:3000
2. رفتن به "Explore" (آیکون کمپاس در منوی سمت چپ)
3. انتخاب datasource: "Tempo"
4. در قسمت "TraceQL" یا "Search" می‌توانید:
   - جستجوی traces بر اساس service name
   - جستجوی traces بر اساس tags
   - مشاهده trace details

## ارسال Traces به Tempo

از OpenTelemetry exporter استفاده کنید:
- Endpoint: `localhost:4318` (برای HTTP)
- Endpoint: `localhost:4317` (برای gRPC)

