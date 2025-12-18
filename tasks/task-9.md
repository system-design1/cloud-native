
# Graceful Shutdown & Lifecycle Management Prompt

## Overview
You are a senior Go backend engineer. Implement **graceful shutdown and lifecycle management** for this Go (Gin) REST API.

> **Important:** Some parts may already be implemented.  
> First **audit the existing code**, and **only add or improve what is missing**.  
> Do not duplicate correct implementations.

---

## Goals
Ensure the application can start, run, and shut down safely in **production and containerized environments**, following cloud‑native best practices.

---

## Task

### 1) Signal Handling
Implement proper OS signal handling:

- Listen for:
  - `SIGINT`
  - `SIGTERM`
- Use `context.Context` to propagate shutdown signals across the application.
- Ensure the application exits with a clean shutdown sequence.

---

### 2) HTTP Server Graceful Shutdown
Update the HTTP server lifecycle to:

- Use `http.Server` (not `gin.Run`)
- Configure:
  - Read timeout
  - Write timeout
  - Idle timeout
- On shutdown signal:
  - Stop accepting new requests
  - Allow **in‑flight requests** to finish within a configurable timeout
  - Log shutdown progress and completion

---

### 3) Dependency Cleanup
Ensure all long‑living resources are properly closed on shutdown:

- Database connections (PostgreSQL)
- OpenTelemetry exporters / tracer provider
- Any background goroutines or workers
- HTTP server itself

Each cleanup step should:
- Respect context deadlines
- Log success or failure with context

---

### 4) Health & Lifecycle States
Integrate lifecycle state awareness:

- Readiness endpoint should:
  - Return **unready** when shutdown starts
- Liveness endpoint should:
  - Stay alive until shutdown is complete
- This behavior must be compatible with:
  - Docker healthchecks
  - Kubernetes readiness/liveness probes

---

### 5) Logging & Observability
During shutdown:

- Log:
  - Received signal
  - Start of shutdown
  - Each cleanup step
  - Forced termination (timeout exceeded)
- Include:
  - Correlation / trace IDs if available
  - Structured JSON logs (Zerolog)

---

### 6) Configuration
All shutdown behavior must be configurable via **environment variables**, such as:

- Shutdown timeout
- Server timeouts
- Grace period duration

Validate configuration at startup.

---

### 7) Code Quality
- Keep shutdown logic centralized (e.g., lifecycle or server package)
- Avoid `os.Exit` except as a last resort
- Use clear function names and comments
- Follow idiomatic Go patterns

---

## Expected Output
- Fully graceful HTTP server shutdown
- Proper signal handling
- Clean resource teardown
- Lifecycle‑aware health endpoints
- Production‑ready, observable shutdown behavior