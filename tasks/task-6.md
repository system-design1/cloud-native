
# Stage 6 Prompt — Docker & Docker Compose

## Overview
You are a senior Go backend engineer. I want you to **implement or review** the Docker and Docker Compose setup for this Go (Gin) REST API project.

> **Important:** Some Docker/Docker Compose work may already be done in previous stages.  
> **First, audit the repository** and **only implement what is missing or incomplete**. If a requirement is already satisfied, do not redo it—just improve it if needed.

---

## Task

### 1) Audit Existing Containerization
- Check if the repo already contains:
  - `Dockerfile`
  - `docker-compose.yml`
  - `.dockerignore`
  - `.env.example` (or similar)
- If any of these exist:
  - Verify they match the requirements below.
  - Fix gaps, improve best practices, and keep changes minimal and clean.

### 2) Multi-stage Dockerfile (Production-Ready)
Create or update a **multi-stage Dockerfile** with **layer optimization**:

- Use a builder stage:
  - Cache dependencies efficiently (copy `go.mod`/`go.sum` first, download modules, then copy source).
  - Build a static or mostly-static binary where appropriate.
- Use a minimal runtime stage:
  - Prefer a minimal base (e.g., distroless or alpine) compatible with the built binary.
  - Run as a **non-root user**.
- Ensure:
  - App listens on a configurable port (from env).
  - Logs go to **stdout/stderr** only (no file logging).
  - Add a **HEALTHCHECK** in the Dockerfile (or in compose) that calls the app’s health endpoint.

### 3) Docker Compose for Local Development
Create or update `docker-compose.yml` to support local dev:

- Services:
  - `postgres` (PostgreSQL)
- Behavior:
  - Load configuration from an `.env` file (or environment section).
  - Expose ports (API + Postgres) for local access.
  - Use a named volume for Postgres data.
  - Add `depends_on` for basic ordering, plus a healthcheck for Postgres if feasible.
- Developer ergonomics:
  - Provide an easy way to run in dev mode (either rebuild on changes, or mount source and run via `go run` inside a dev container — pick a simple approach and document it).

### 4) .dockerignore
Create or update `.dockerignore` to keep Docker context small, excluding:
- `.git/`
- `bin/` or build outputs
- `.idea/`, `.vscode/`
- `*.log`
- local env files (except `.env.example` if needed)
- OS junk files (e.g., `.DS_Store`)

### 5) Documentation
Update `README.md` (or add a section) explaining:
- How to run locally with Docker Compose
- Required environment variables (refer to `.env.example`)
- Common commands (`docker compose up`, rebuild, logs, down)
- Health endpoints used for readiness/liveness

---

## Expected Output
- A clean, production-leaning **multi-stage `Dockerfile`**
- A working `docker-compose.yml` for local development with Postgres
- A correct `.dockerignore`
- Clear documentation updates