
# Stage 8 Prompt — Docker Production Hardening & Documentation

## Overview
You are a senior Go backend engineer. I want you to **review and harden** the Docker-related setup for production and improve documentation.

> **Important:** A lot of Docker work may already be implemented in previous stages (Dockerfile, compose, .dockerignore).  
> **First, audit the repository** and **only change/add what is missing or suboptimal**.  
> If something is already correct, leave it as-is.

---

## Task

### 1) Audit What Exists
- Inspect current:
  - `Dockerfile`
  - `docker-compose.yml` (if present)
  - `.dockerignore`
  - `README.md`
- Make a short checklist in the PR/commit message (or in comments) of what was already compliant vs what you changed.

### 2) Production-Ready Dockerfile Improvements
If not already done, or if improvements are needed, update the Dockerfile to follow production best practices:

- **Multi-stage build** with dependency caching:
  - Copy `go.mod` and `go.sum` first, run `go mod download`, then copy the rest of the source.
- **Minimal runtime image**:
  - Prefer distroless (or alpine if distroless is not feasible).
- **Non-root runtime user**:
  - Ensure the container runs as non-root.
- **Correct signal handling**:
  - Ensure the app receives SIGTERM properly (use `ENTRYPOINT` exec-form, avoid shell form).
- **Healthcheck**:
  - Add a container healthcheck against an internal health endpoint (liveness or readiness).
- **Build reproducibility**:
  - Pin base image versions where reasonable (avoid `latest`).
- **Performance**:
  - Set build flags sensibly (e.g., `-trimpath`, stripping symbols if appropriate).
  - Ensure the binary location and working directory are clean and minimal.

### 3) .dockerignore Hardening
If not already correct, ensure `.dockerignore` excludes all unnecessary files to reduce build context and speed up builds:

- `.git/`
- `.idea/`, `.vscode/`
- `**/*.log`
- `tmp/`, `bin/`, `dist/`
- test output artifacts
- OS metadata files (`.DS_Store`)
- local env files (`.env`, `.env.*`) but keep `.env.example`

### 4) Documentation Updates
Update `README.md` with a small, clear section that covers:

- Running locally (Docker Compose)
- Building and running production image (Docker build/run examples)
- Environment variable configuration (link to `.env.example`)
- Health endpoints used for container healthchecks
- Where logs are written (stdout/stderr, structured JSON)

### 5) CI/CD Notes (Do Not Implement Yet)
- Do **not** implement CI/CD in this stage unless it already exists and needs minor fixes.
- Add a short “CI/CD Plan” section in `README.md` describing a future pipeline:
  - Run `go test ./...`
  - Run `golangci-lint`
  - Build Docker image
  - (Optional) push image to registry
  - (Optional) run security scans (Trivy)
- Keep it brief and actionable.

---

## Expected Output
- A Dockerfile that is production-leaning, secure (non-root), and efficient (cached layers).
- A hardened `.dockerignore`.
- Updated documentation in `README.md`.
- A brief CI/CD plan section in docs (planning only).