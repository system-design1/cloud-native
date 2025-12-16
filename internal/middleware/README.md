# Middleware Package

این پکیج شامل middleware های مورد نیاز برای logging و error handling است.

## Middleware ها

### 1. CorrelationIDMiddleware

می‌تواند correlation ID را از هدر `X-Correlation-ID` دریافت کند یا یک ID جدید تولید کند. این ID برای tracing درخواست‌ها در سراسر سیستم استفاده می‌شود.

**استفاده:**
```go
router.Use(middleware.CorrelationIDMiddleware())
```

### 2. RequestResponseLoggingMiddleware

تمام درخواست‌ها و پاسخ‌های HTTP را به صورت structured JSON لاگ می‌کند.

**اطلاعات لاگ شده:**
- Correlation ID
- HTTP Method
- Request Path
- Query Parameters
- Client IP
- User Agent
- Response Status Code
- Response Size
- Latency (زمان پاسخ)

**استفاده:**
```go
router.Use(middleware.RequestResponseLoggingMiddleware())
```

**فرمت لاگ:**
```json
{
  "level": "info",
  "correlation_id": "abc-123-def",
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

### 3. ErrorHandlerMiddleware

یک global error handler که تمام خطاها را catch کرده و پاسخ خطای استاندارد را برمی‌گرداند.

**ویژگی‌ها:**
- Catch کردن تمام خطاهای Application
- لاگ کردن خطاها با correlation ID
- بازگرداندن پاسخ خطای استاندارد
- عدم افشای اطلاعات حساس به client

**استفاده:**
```go
router.Use(middleware.ErrorHandlerMiddleware())
```

**استفاده در Handler:**
```go
// در handler خود:
if err != nil {
    middleware.ErrorHandler(c, err)
    return
}
```

**فرمت پاسخ خطا:**
```json
{
  "error": "Bad Request",
  "message": "Validation failed",
  "code": 400,
  "details": "Email is required",
  "request_id": "abc-123-def"
}
```

## ترتیب Middleware

ترتیب middleware ها مهم است:

1. **Recovery** (Gin built-in) - برای catch کردن panic
2. **CorrelationIDMiddleware** - باید قبل از logging باشد
3. **RequestResponseLoggingMiddleware** - باید بعد از CorrelationID باشد
4. **ErrorHandlerMiddleware** - باید آخرین middleware باشد

## مثال استفاده کامل

```go
router := gin.New()

// 1. Recovery
router.Use(gin.Recovery())

// 2. Correlation ID
router.Use(middleware.CorrelationIDMiddleware())

// 3. Logging
router.Use(middleware.RequestResponseLoggingMiddleware())

// 4. Error Handler
router.Use(middleware.ErrorHandlerMiddleware())
```

