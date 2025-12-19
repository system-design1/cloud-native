# 🐛 راهنمای کامل Debug با VS Code

این راهنما به صورت گام‌به‌گام نحوه debug کردن پروژه را با VS Code توضیح می‌دهد.

---

## 📋 مرحله 1: نصب پیش‌نیازها

### 1.1 نصب Go Extension

1. VS Code را باز کنید
2. به Extensions بروید (`Ctrl+Shift+X` یا `Cmd+Shift+X` در Mac)
3. "Go" را جستجو کنید
4. Extension "Go" توسط **Go Team at Google** را نصب کنید
5. VS Code را reload کنید (یا `Ctrl+R`)

### 1.2 نصب Go Tools

بعد از نصب extension، VS Code به صورت خودکار پیشنهاد نصب tools را می‌دهد:

1. وقتی notification ظاهر شد، روی **"Install All"** کلیک کنید
2. یا دستی:
   ```bash
   # نصب Delve (debugger)
   go install github.com/go-delve/delve/cmd/dlv@latest
   
   # بررسی نصب
   dlv version
   ```

**نکته:** اگر notification ظاهر نشد، می‌توانید از Command Palette استفاده کنید:
- `Ctrl+Shift+P` (یا `Cmd+Shift+P`)
- "Go: Install/Update Tools" را تایپ کنید
- تمام tools را انتخاب کنید و Enter بزنید

---

## 📋 مرحله 2: آماده‌سازی پروژه

### 2.1 ایجاد فایل .env

```bash
# در terminal VS Code یا terminal خارجی
make dev-setup
```

این دستور:
- فایل `.env` را از `env.example` ایجاد می‌کند
- `DB_HOST` را به `localhost` تنظیم می‌کند (برای local development)

### 2.2 راه‌اندازی Database

```bash
# راه‌اندازی PostgreSQL برای local development
make dev-db-up
```

**بررسی:**
```bash
# بررسی وضعیت database
docker ps | grep postgres
```

**خروجی مورد انتظار:**
```
go-backend-postgres-dev   Up   ...   5432/tcp
```

---

## 📋 مرحله 3: شروع Debug

### روش 1: از Run and Debug Panel (توصیه می‌شود)

1. **باز کردن Run and Debug Panel:**
   - روی آیکون Debug در sidebar کلیک کنید (یا `Ctrl+Shift+D`)
   - یا از منو: `View` → `Run and Debug`

2. **انتخاب Configuration:**
   - از dropdown بالا، **"Debug Go Server (Local)"** را انتخاب کنید

3. **شروع Debug:**
   - روی دکمه سبز **"Start Debugging"** کلیک کنید
   - یا `F5` را فشار دهید

4. **بررسی:**
   - در terminal VS Code باید لاگ‌های برنامه را ببینید
   - باید پیام `"HTTP server is running and ready to accept connections"` را ببینید

### روش 2: از Command Palette

1. `Ctrl+Shift+P` (یا `Cmd+Shift+P`) را فشار دهید
2. "Debug: Start Debugging" را تایپ کنید
3. **"Debug Go Server (Local)"** را انتخاب کنید
4. Enter بزنید

### روش 3: از نوار بالای VS Code

1. در نوار بالای VS Code، dropdown "Debug Go Server (Local)" را پیدا کنید
2. روی دکمه سبز Play کلیک کنید

---

## 📋 مرحله 4: استفاده از Breakpoints

### 4.1 اضافه کردن Breakpoint

1. فایلی که می‌خواهید debug کنید را باز کنید (مثلاً `cmd/server/main.go`)
2. روی شماره خط مورد نظر کلیک کنید (سمت چپ شماره خط)
3. یک نقطه قرمز ظاهر می‌شود - این breakpoint است

**یا:**
- مکان‌نما را روی خط مورد نظر بگذارید
- `F9` را فشار دهید

### 4.2 مثال عملی: Debug یک Handler

1. فایل `internal/api/handlers.go` را باز کنید
2. در تابع `HelloHandler` یک breakpoint قرار دهید (مثلاً خط 74)
3. Debug را شروع کنید (`F5`)
4. در terminal یا Postman، یک request بفرستید:
   ```bash
   curl http://localhost:8080/hello
   ```
5. برنامه در breakpoint متوقف می‌شود
6. می‌توانید variables را inspect کنید

### 4.3 Conditional Breakpoints

برای breakpoint شرطی:

1. روی breakpoint کلیک راست کنید
2. "Edit Breakpoint" را انتخاب کنید
3. شرط را وارد کنید (مثلاً `method == "POST"`)

---

## 📋 مرحله 5: استفاده از Debug Features

### 5.1 Variables Panel

- در سمت چپ، پنل **"Variables"** را ببینید
- تمام variables محلی و global را نشان می‌دهد
- می‌توانید مقادیر را تغییر دهید (در حالت debug)

### 5.2 Watch Expressions

برای monitor کردن یک expression:

1. در پنل **"Watch"** کلیک کنید
2. روی `+` کلیک کنید
3. expression را وارد کنید (مثلاً `cfg.Server.Port`)

### 5.3 Call Stack

- در پنل **"Call Stack"** می‌توانید ببینید چگونه به این نقطه رسیده‌اید
- می‌توانید روی هر frame کلیک کنید تا به آن بروید

### 5.4 Debug Console

- در پایین، **"Debug Console"** را باز کنید
- می‌توانید expressions را evaluate کنید
- می‌توانید variables را inspect کنید

**مثال:**
```
> cfg.Server.Port
8080
> len(cfg.Database.Host)
9
```

---

## 📋 مرحله 6: Navigation در Debug

### Shortcuts مهم:

| کلید | عمل |
|------|-----|
| `F5` | Continue (ادامه اجرا) |
| `F9` | Toggle Breakpoint |
| `F10` | Step Over (اجرای خط فعلی) |
| `F11` | Step Into (ورود به function) |
| `Shift+F11` | Step Out (خروج از function) |
| `Shift+F5` | Stop Debugging |
| `Ctrl+Shift+F5` | Restart Debugging |

### توضیح:

- **Step Over (`F10`)**: خط فعلی را اجرا می‌کند و به خط بعد می‌رود
- **Step Into (`F11`)**: اگر خط فعلی یک function call باشد، وارد function می‌شود
- **Step Out (`Shift+F11`)**: از function فعلی خارج می‌شود
- **Continue (`F5`)**: اجرا را ادامه می‌دهد تا breakpoint بعدی

---

## 📋 مرحله 7: Debug Tests

### 7.1 Debug تمام Tests

1. از dropdown، **"Debug Go Tests"** را انتخاب کنید
2. `F5` را فشار دهید
3. تمام tests با debug اجرا می‌شوند

### 7.2 Debug یک Test خاص

1. فایل test را باز کنید (مثلاً `internal/config/config_test.go`)
2. در یک test function یک breakpoint قرار دهید
3. از dropdown، **"Debug Current Test"** را انتخاب کنید
4. `F5` را فشار دهید

---

## 🎯 مثال عملی کامل

### مثال: Debug یک API Request

1. **راه‌اندازی:**
   ```bash
   make dev-setup
   make dev-db-up
   ```

2. **شروع Debug:**
   - `F5` را فشار دهید
   - یا از Run and Debug panel شروع کنید

3. **قرار دادن Breakpoint:**
   - فایل `internal/api/handlers.go` را باز کنید
   - در خط 74 (تابع `HelloHandler`) یک breakpoint قرار دهید

4. **ارسال Request:**
   ```bash
   # در terminal دیگر
   curl http://localhost:8080/hello
   ```

5. **Inspect Variables:**
   - برنامه در breakpoint متوقف می‌شود
   - در Variables panel، `c` را ببینید
   - `c.Request.Method` را expand کنید
   - `c.Request.URL.Path` را ببینید

6. **Step Through:**
   - `F10` را فشار دهید تا خط به خط پیش بروید
   - مقادیر variables را در Variables panel ببینید

7. **Continue:**
   - `F5` را فشار دهید تا اجرا ادامه یابد
   - Response را در terminal ببینید

---

## 🔧 Troubleshooting

### مشکل: "Error: .env file not found"

**راه‌حل:**
```bash
make dev-setup
```

### مشکل: "connection refused" برای database

**راه‌حل:**
```bash
# بررسی وضعیت database
docker ps | grep postgres

# راه‌اندازی database
make dev-db-up

# بررسی .env
grep DB_HOST .env
# باید localhost باشد
```

### مشکل: Breakpoint کار نمی‌کند

**راه‌حل:**
1. مطمئن شوید `GIN_MODE=debug` در `.env` است
2. فایل را save کنید (`Ctrl+S`)
3. Debug را restart کنید (`Ctrl+Shift+F5`)

### مشکل: "dlv: command not found"

**راه‌حل:**
```bash
go install github.com/go-delve/delve/cmd/dlv@latest

# بررسی
dlv version
```

### مشکل: Environment variables load نمی‌شوند

**راه‌حل:**
1. مطمئن شوید `.env` در root پروژه است
2. Format درست است (بدون فاصله قبل و بعد `=`)
3. VS Code را restart کنید

---

## 💡 نکات پیشرفته

### 1. Debug با Arguments

در `.vscode/launch.json` می‌توانید arguments اضافه کنید:

```json
{
  "name": "Debug Go Server (Local)",
  "args": ["--flag", "value"]
}
```

### 2. Debug با Environment Variables اضافی

```json
{
  "name": "Debug Go Server (Local)",
  "env": {
    "CUSTOM_VAR": "value",
    "LOG_LEVEL": "debug"
  }
}
```

### 3. Debug در Remote Server

برای debug در remote server، از "Attach" configuration استفاده کنید.

---

## 📚 منابع بیشتر

- [VS Code Go Extension](https://marketplace.visualstudio.com/items?itemName=golang.Go)
- [Delve Documentation](https://github.com/go-delve/delve)
- [VS Code Debugging](https://code.visualstudio.com/docs/editor/debugging)

---

**موفق باشید! 🎉**

