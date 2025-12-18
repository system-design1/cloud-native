
# Stage 9 Prompt — Project Documentation (README, Swagger, ADR)

## Overview
You are a senior Go backend engineer. I want you to **review, complete, and standardize the project documentation**.

> **Important:** Some documentation (README, Swagger setup, or ADRs) may already exist.  
> **First audit the repository**, and **only add or improve what is missing or incomplete**.  
> If something is already done correctly, do not redo it.

---

## Task

### 1) README.md — Comprehensive Project Documentation
Create or improve a **production-quality README.md** that includes the following sections:

- **Project Overview**
  - What the service does
  - High-level architecture (monolithic modular, cloud-native, 12-factor compliant)

- **Tech Stack**
  - Go, Gin
  - PostgreSQL
  - Zerolog
  - OpenTelemetry + Tempo
  - Prometheus
  - Grafana + Loki
  - Docker / Docker Compose

- **Prerequisites**
  - Go version
  - Docker / Docker Compose

- **Configuration**
  - Explain environment-based configuration (12-factor)
  - Reference `.env.example`
  - Describe important environment variables (server port, log level, OTEL, DB, etc.)

- **Running Locally**
  - With Docker Compose
  - Without Docker (go run / make run)

- **Health & Observability**
  - Liveness & readiness endpoints
  - Metrics endpoint (`/metrics`)
  - Tracing (OpenTelemetry)
  - Logging behavior (structured JSON to stdout)

- **Graceful Shutdown**
  - Signal handling and shutdown behavior

---

### 2) Swagger / OpenAPI Documentation
If Swagger/OpenAPI is not already implemented, add it. If it exists, review and improve it.

Requirements:

- Use Swagger/OpenAPI (e.g. swaggo or equivalent)
- Document all public endpoints:
  - `/hello`
  - `/delayed-hello`
  - health endpoints
  - metrics endpoint (if applicable)
- Include:
  - Request/response schemas
  - Status codes
  - Example responses

Documentation updates:
- Add a **Swagger section to README.md** explaining:
  - How to generate/update Swagger docs
  - How to access Swagger UI locally
  - Any required build steps or commands

Do not forget to update documentation whenever Swagger is added or changed.

---

### 3) Architecture Decision Records (ADR)
Add an `adr/` or `docs/adr/` directory with at least **one example ADR**.

ADR requirements:
- Use a simple ADR template including:
  - Title
  - Status
  - Context
  - Decision
  - Consequences
- Example topics:
  - Why Gin was chosen
  - Why OpenTelemetry + Tempo
  - Why Grafana Loki instead of ELK
  - Why Prometheus for metrics

Ensure ADRs are referenced briefly in README.md.

---

## Expected Output
- A clear, professional **README.md**
- Swagger/OpenAPI fully documented and usable
- Documentation updated wherever Swagger is introduced
- At least one well-written **ADR** documenting a real architectural decision