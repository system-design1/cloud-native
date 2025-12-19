# 🔧 حل مشکل Docker Build

## مشکل

`make docker-up-rebuild` با خطای Alpine package manager مواجه می‌شود:

```
ERROR: unable to select packages:
  ca-certificates (no such package)
  wget (no such package)
```

## دلیل

مشکل از Alpine mirror است که در دسترس نیست یا permission denied می‌دهد.

## راه‌حل‌های موقت

### راه‌حل 1: استفاده از Image موجود

اگر image قبلاً build شده است:

```bash
# فقط restart کنید (بدون rebuild)
make docker-down
make docker-up
```

### راه‌حل 2: Build بدون cache

```bash
# Build بدون cache
docker-compose build --no-cache api

# یا
make docker-build-rebuild
```

### راه‌حل 3: استفاده از Alpine latest

اگر مشکل ادامه دارد، می‌توانید در Dockerfile از `alpine:latest` استفاده کنید:

```dockerfile
FROM alpine:latest
```

### راه‌حل 4: استفاده از distroless

برای production، می‌توانید از distroless استفاده کنید (نیازی به apk ندارد):

```dockerfile
FROM gcr.io/distroless/static-debian12:nonroot
```

## راه‌حل دائمی

اگر مشکل network است:

1. **بررسی network:**
   ```bash
   ping dl-cdn.alpinelinux.org
   ```

2. **استفاده از proxy:**
   ```bash
   export HTTP_PROXY=http://your-proxy:port
   export HTTPS_PROXY=http://your-proxy:port
   docker-compose build api
   ```

3. **استفاده از VPN یا تغییر DNS**

## تست

بعد از اعمال تغییرات:

```bash
make docker-up-rebuild
```

یا:

```bash
docker-compose build --no-cache api
docker-compose up -d
```

---

**نکته:** اگر مشکل network است، ممکن است نیاز به تغییر DNS یا استفاده از VPN داشته باشید.

