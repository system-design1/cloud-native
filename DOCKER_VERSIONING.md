# 📌 Docker Image Versioning: Latest vs Pinned Versions

## مقایسه Latest vs Pinned Version

### ❌ استفاده از `latest` (مشکلات)

**مشکلات:**
1. **عدم Reproducibility**: هر بار که pull می‌کنید، ممکن است version متفاوت باشد
2. **Breaking Changes**: ممکن است version جدید breaking changes داشته باشد
3. **عدم کنترل**: نمی‌دانید چه version ای در حال اجرا است
4. **مشکلات Production**: در production، باید دقیقاً بدانید چه چیزی در حال اجرا است
5. **Debugging سخت‌تر**: اگر مشکلی پیش بیاید، نمی‌دانید کدام version مشکل دارد

**مثال:**
```yaml
image: grafana/grafana:latest  # ❌ بد
```

### ✅ استفاده از Pinned Version (توصیه می‌شود)

**مزایا:**
1. **Reproducibility**: همیشه همان version اجرا می‌شود
2. **کنترل**: می‌دانید دقیقاً چه version ای در حال اجرا است
3. **امنیت**: می‌توانید version های خاص را برای security patches انتخاب کنید
4. **Stability**: version های stable را می‌توانید انتخاب کنید
5. **Debugging آسان‌تر**: می‌دانید کدام version مشکل دارد

**مثال:**
```yaml
image: grafana/grafana:10.4.0  # ✅ خوب
```

---

## وضعیت فعلی پروژه

### ✅ درست (Pinned):
- `postgres:17.2-alpine` در `docker-compose.yml` ✅
- `golang:1.25-alpine` در `Dockerfile` ✅
- `alpine:3.20` در `Dockerfile` ✅
- `grafana/tempo:2.5.0` در `docker-compose.observability.yml` ✅
- `jaegertracing/all-in-one:1.57` در `docker-compose.observability.yml` ✅
- `prom/prometheus:v2.53.0` در `docker-compose.observability.yml` ✅
- `grafana/grafana:10.4.0` در `docker-compose.observability.yml` ✅

---

## Version های انتخاب شده

### Grafana: `10.4.0`
- **دلیل**: LTS version، stable و پشتیبانی طولانی‌مدت
- **تاریخ**: December 2024
- **منبع**: https://hub.docker.com/r/grafana/grafana/tags

### Tempo: `2.5.0`
- **دلیل**: Stable version، سازگار با Grafana 10.x
- **تاریخ**: December 2024
- **منبع**: https://hub.docker.com/r/grafana/tempo/tags

### Prometheus: `v2.53.0`
- **دلیل**: Latest stable version
- **تاریخ**: December 2024
- **منبع**: https://hub.docker.com/r/prom/prometheus/tags

### Jaeger: `1.57`
- **دلیل**: Latest stable version
- **تاریخ**: December 2024
- **منبع**: https://hub.docker.com/r/jaegertracing/all-in-one/tags

---

## نحوه به‌روزرسانی Versions

### 1. بررسی Latest Stable Version:

```bash
# برای هر image
docker pull grafana/grafana:latest
docker inspect grafana/grafana:latest | grep -i version

# یا از Docker Hub:
# https://hub.docker.com/r/grafana/grafana/tags
```

### 2. تست Version جدید:

```bash
# Pull version جدید
docker pull grafana/grafana:10.5.0

# تست در development
# تغییر docker-compose.observability.yml
# اجرا و تست
make observability-up-rebuild
```

### 3. به‌روزرسانی:

```bash
# بعد از تست موفق
# تغییر version در docker-compose.observability.yml
# commit تغییرات
git add docker-compose.observability.yml
git commit -m "chore: update Grafana to 10.5.0"
```

---

## Best Practices

### 1. برای Production:
- ✅ همیشه از pinned version استفاده کنید
- ✅ از LTS versions استفاده کنید (برای Grafana)
- ✅ Version ها را در changelog document کنید

### 2. برای Development:
- ✅ بهتر است از pinned version استفاده کنید
- ⚠️ استفاده از latest قابل قبول است اما توصیه نمی‌شود

### 3. Security Updates:
- ✅ برای security patches، version را به‌روزرسانی کنید
- ✅ همیشه changelog را بررسی کنید
- ✅ در development تست کنید قبل از production

---

## خلاصه

✅ **همه image ها حالا pinned شده‌اند**
✅ **Reproducibility تضمین شده است**
✅ **کنترل کامل بر versions**
✅ **Production-ready**

**نکته**: برای به‌روزرسانی، از `make observability-up-rebuild` استفاده کنید.
