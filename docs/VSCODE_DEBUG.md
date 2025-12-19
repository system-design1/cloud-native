# راهنمای Debug با VS Code

این راهنما نحوه استفاده از VS Code برای اجرا و debug کردن برنامه را توضیح می‌دهد.

## پیش‌نیازها

1. **نصب Go Extension**: 
   - در VS Code، به Extensions بروید (Ctrl+Shift+X)
   - "Go" را جستجو کنید و نصب کنید (توسط Go Team at Google)

2. **نصب Go Tools**:
   - بعد از نصب extension، VS Code به صورت خودکار پیشنهاد نصب tools را می‌دهد
   - یا می‌توانید دستی نصب کنید:
   ```bash
   go install github.com/go-delve/delve/cmd/dlv@latest
   ```

3. **ایجاد فایل .env**:
   ```bash
   make dev-setup
   ```

## استفاده از Debug

### روش 1: Debug با Launch Configuration

1. **راه‌اندازی دیتابیس** (در ترمینال):
   ```bash
   make dev-db-up
   ```

2. **شروع Debug**:
   - به تب "Run and Debug" بروید (Ctrl+Shift+D)
   - از dropdown بالا، "Debug Go Server (Local)" را انتخاب کنید
   - روی دکمه سبز "Start Debugging" کلیک کنید (F5)

3. **تنظیم Breakpoints**:
   - روی خطوط کد کلیک کنید تا breakpoint اضافه شود
   - یا از F9 استفاده کنید

### روش 2: Debug از Command Palette

1. **Command Palette** را باز کنید (Ctrl+Shift+P)
2. "Debug: Start Debugging" را تایپ کنید
3. "Debug Go Server (Local)" را انتخاب کنید

## Configuration های موجود

### 1. Debug Go Server (Local)
- **استفاده**: اجرا و debug برنامه اصلی
- **Environment**: از فایل `.env` استفاده می‌کند
- **Port**: 8080 (قابل تغییر در `.env`)

### 2. Debug Go Server (Attach)
- **استفاده**: اتصال به یک process در حال اجرا
- **نکته**: ابتدا باید برنامه را به صورت دستی اجرا کنید

### 3. Debug Go Tests
- **استفاده**: اجرای تمام تست‌ها با debug
- **نکته**: می‌توانید breakpoint در تست‌ها قرار دهید

### 4. Debug Current Test
- **استفاده**: debug تست در فایل فعلی
- **نکته**: باید در یک فایل تست باشید

## استفاده از Tasks

VS Code tasks برای اجرای دستورات مختلف:

### دسترسی به Tasks:
1. Command Palette (Ctrl+Shift+P)
2. "Tasks: Run Task" را تایپ کنید
3. task مورد نظر را انتخاب کنید

### Tasks موجود:

- **go: build**: Build برنامه
- **go: run**: اجرای برنامه
- **go: test**: اجرای تمام تست‌ها
- **go: test (current package)**: تست package فعلی
- **make: dev-setup**: ایجاد فایل .env
- **make: dev-db-up**: راه‌اندازی دیتابیس
- **make: dev-db-down**: توقف دیتابیس
- **make: dev**: راه‌اندازی کامل

## نکات مهم

### 1. Environment Variables
- VS Code به صورت خودکار فایل `.env` را می‌خواند
- اگر `.env` وجود ندارد، از `make dev-setup` استفاده کنید
- برای تغییر تنظیمات، فایل `.env` را ویرایش کنید

### 2. Database Connection
- قبل از debug، مطمئن شوید دیتابیس در حال اجرا است:
  ```bash
  make dev-db-up
  ```
- `DB_HOST` در `.env` باید `localhost` باشد

### 3. Breakpoints
- می‌توانید breakpoint در هر خط کد قرار دهید
- Breakpoint های conditional هم پشتیبانی می‌شود
- برای disable کردن، روی breakpoint کلیک راست کنید

### 4. Debug Console
- در حین debug، می‌توانید از Debug Console استفاده کنید
- می‌توانید expressions را evaluate کنید
- می‌توانید variables را inspect کنید

## مثال عملی

### مثال 1: Debug یک API Handler

1. فایل handler را باز کنید (مثلاً `internal/api/handlers.go`)
2. یک breakpoint در تابع handler قرار دهید
3. Debug را شروع کنید (F5)
4. یک request به API بفرستید (با Postman یا curl)
5. برنامه در breakpoint متوقف می‌شود
6. می‌توانید variables را inspect کنید

### مثال 2: Debug Configuration Loading

1. فایل `internal/config/config.go` را باز کنید
2. یک breakpoint در تابع `Load()` قرار دهید
3. Debug را شروع کنید
4. می‌توانید ببینید configuration چگونه load می‌شود

### مثال 3: Debug یک Test

1. یک فایل تست را باز کنید
2. یک breakpoint در تست قرار دهید
3. از "Debug Current Test" استفاده کنید
4. تست در breakpoint متوقف می‌شود

## Troubleshooting

### مشکل: "Error: .env file not found"
```bash
make dev-setup
```

### مشکل: "connection refused" برای دیتابیس
```bash
# بررسی کنید که دیتابیس در حال اجرا است
make dev-db-up

# یا بررسی کنید
docker ps | grep postgres
```

### مشکل: "dlv: command not found"
```bash
# نصب delve
go install github.com/go-delve/delve/cmd/dlv@latest

# یا از Go extension استفاده کنید که به صورت خودکار نصب می‌کند
```

### مشکل: Breakpoint کار نمی‌کند
- مطمئن شوید که `GIN_MODE=debug` است
- مطمئن شوید که فایل را save کرده‌اید
- دوباره build کنید

### مشکل: Environment variables load نمی‌شوند
- مطمئن شوید که فایل `.env` در root پروژه است
- مطمئن شوید که format درست است (بدون فاصله قبل و بعد =)
- VS Code را restart کنید

## Shortcuts مفید

- **F5**: Start Debugging
- **F9**: Toggle Breakpoint
- **F10**: Step Over
- **F11**: Step Into
- **Shift+F11**: Step Out
- **Shift+F5**: Stop Debugging
- **Ctrl+Shift+F5**: Restart Debugging

## تنظیمات پیشرفته

اگر می‌خواهید تنظیمات debug را تغییر دهید، فایل `.vscode/launch.json` را ویرایش کنید.

مثلاً برای تغییر port:
```json
{
  "name": "Debug Go Server (Local)",
  "env": {
    "SERVER_PORT": "3000"
  }
}
```

یا برای اضافه کردن arguments:
```json
{
  "args": ["--flag", "value"]
}
```

## نکات نهایی

1. **همیشه از `make dev-db-up` استفاده کنید** قبل از debug
2. **از `GIN_MODE=debug` استفاده کنید** برای لاگ‌های بهتر
3. **از Debug Console استفاده کنید** برای evaluate کردن expressions
4. **Watch expressions اضافه کنید** برای monitor کردن variables
5. **از Call Stack استفاده کنید** برای دیدن flow اجرا

