Analyze Phase 1 of the availability lab: single Traefik gateway in front of the OTP service.

Goal:
I want to demonstrate availability concepts step by step.
The first implementation phase should create a baseline gateway scenario:

Client -> Traefik -> OTP Service

Current state:
- Direct baseline already exists:
  Client -> OTP Service
- I now need Traefik as the first gateway layer.
- Later phases will use two Traefik instances with Keepalived and a VIP.
- Do not implement Keepalived yet.
- Do not implement Nginx yet.
- Do not add OTP app replicas yet.

Constraints:
- Run locally on my Ubuntu machine.
- Prefer Docker Compose for this first phase.
- Keep the implementation small.
- Do not change OTP application code unless absolutely necessary.
- The goal is a demo/lab foundation for gateway availability.

Please analyze:
1. Recommended local topology.
2. Whether to create a separate compose file under deploy/availability-lab.
3. Required Traefik static config.
4. Required Traefik dynamic config.
5. Which OTP endpoint to use first: /health or /ready.
6. How to expose Traefik locally.
7. How to prove traffic goes through Traefik.
8. Whether to add a response header like X-Gateway-Node.
9. How to run a simple traffic loop.
10. How to test proxy behavior when OTP service stops/restarts.
11. Exact files to create.
12. What should be deferred to the Keepalived phase.

Important:
- Do not implement yet.
- Do not modify files yet.
- Only analyze.
- Keep this phase focused only on single Traefik gateway baseline.

----------

# Codex Implementation Prompt — Phase 1: Traefik Single Gateway Baseline

Implement Phase 1 of the local availability lab: a single Traefik gateway in front of the existing OTP service.

## Goal

Create a minimal local gateway baseline:

```text
Client -> Traefik -> OTP Service
```

This phase is **not HA yet**.

The goal is to create a clean local foundation before later phases:

```text
Client -> VIP -> Traefik-1 / Traefik-2 -> OTP Service
```

with Keepalived.

## Current State

- Direct baseline already exists:

```text
Client -> OTP Service
```

- OTP service is expected to run on the local host on port `8080`.
- This phase should proxy traffic from Traefik to the already-running OTP service.
- Do not run the OTP service inside this availability lab compose file yet.

## Scope

Create only these files:

```text
deploy/availability-lab/traefik-baseline/compose.yml
deploy/availability-lab/traefik-baseline/traefik.yml
deploy/availability-lab/traefik-baseline/dynamic.yml
deploy/availability-lab/traefik-baseline/README.md
```

## Important Constraints

Do not modify OTP application code.

Do not modify:

- `cmd/server/main.go`
- `internal/...`
- existing main Docker Compose files
- config/env files
- migrations
- tests

Do not add:

- Keepalived
- Nginx
- OTP service replicas
- traffic-loop script
- production TLS
- service discovery
- advanced retry/circuit breaker config

Keep this phase focused only on:

```text
Single Traefik gateway baseline
```

## Required Topology

Use Traefik in Docker Compose and proxy to the OTP service running on the host:

```text
Client on Ubuntu
    |
    v
Traefik container exposed on localhost:8081
    |
    v
Existing OTP service on host localhost:8080
```

Inside the Traefik container, use:

```text
host.docker.internal:8080
```

On Linux, Docker needs this explicit mapping in compose:

```yaml
extra_hosts:
  - "host.docker.internal:host-gateway"
```

## Port Mapping

Expose:

```text
localhost:8081 -> Traefik web entrypoint
localhost:8082 -> Traefik dashboard
```

Expected local URLs:

```text
http://localhost:8081/health
http://localhost:8082/dashboard/
```

## Traefik Static Config Requirements

Create `traefik.yml`.

It should include:

- `web` entrypoint on `:80`
- `traefik` dashboard/API entrypoint on `:8080`
- file provider pointing to `/etc/traefik/dynamic.yml`
- access logs enabled
- normal log level, e.g. `INFO`
- dashboard enabled for local demo

Important:

`api.insecure=true` is acceptable only for this local lab.

The README must explicitly say this dashboard config is not production-safe.

## Traefik Dynamic Config Requirements

Create `dynamic.yml`.

It should include:

- one router for OTP
- route all paths using `PathPrefix(`/`)`
- service pointing to:

```text
http://host.docker.internal:8080
```

- backend health check against:

```text
/health
```

- middleware that adds this response header:

```text
X-Gateway-Node: traefik-baseline
```

This header is required for demo purposes so we can prove traffic passed through Traefik.

## README Requirements

Create a clear `README.md` under:

```text
deploy/availability-lab/traefik-baseline/README.md
```

The README must include:

### 1. Purpose

Explain that this is Phase 1 of the availability lab:

```text
Client -> Traefik -> OTP Service
```

Explain clearly that this is not HA yet.

### 2. Prerequisite

OTP service must already be running on host port `8080`.

Add a check command:

```bash
curl -i http://localhost:8080/health
```

### 3. Start Traefik

Command from the lab directory:

```bash
docker compose up -d
```

or from repo root:

```bash
docker compose -f deploy/availability-lab/traefik-baseline/compose.yml up -d
```

### 4. Test Gateway

Add:

```bash
curl -i http://localhost:8081/health
```

Expected:

- same health response as direct OTP
- response includes:

```text
X-Gateway-Node: traefik-baseline
```

### 5. Open Traefik Dashboard

Add:

```text
http://localhost:8082/dashboard/
```

Mention that insecure dashboard is local-lab only.

### 6. Prove Traffic Goes Through Traefik

Explain these signals:

- response header `X-Gateway-Node`
- Traefik access logs
- dashboard router/service visibility

Command:

```bash
docker compose -f deploy/availability-lab/traefik-baseline/compose.yml logs -f traefik
```

### 7. Simple Traffic Loop

Add this command:

```bash
while true; do
  date +"%H:%M:%S"
  curl -s -o /dev/null -w "status=%{http_code} time=%{time_total}\\n" http://localhost:8081/health
  sleep 0.5
done
```

### 8. Failure Demo

Explain test:

1. Start OTP service.
2. Start Traefik.
3. Start traffic loop.
4. Stop OTP service.
5. Expected:
   - Traefik returns `502` or `503`.
   - Access logs show upstream failure.
6. Restart OTP service.
7. Expected:
   - gateway requests recover.

Also include Traefik stop test:

```bash
docker compose -f deploy/availability-lab/traefik-baseline/compose.yml stop traefik
```

Expected:

- `localhost:8081` becomes unavailable.
- direct `localhost:8080/health` still works if OTP service is running.

### 9. What This Phase Proves

Explain:

- Traefik can act as a gateway in front of OTP.
- Traffic can be observed through Traefik.
- Single gateway is still a single point of failure.
- This motivates the next Keepalived HA phase.

### 10. Deferred

List deferred items:

- Keepalived
- VIP
- second Traefik instance
- Nginx
- OTP replicas
- running OTP inside this lab compose
- Redis/Postgres HA
- TLS
- Traefik Docker provider
- `/ready` based dependency-aware routing
- production dashboard security
- load testing beyond simple curl/hey

## Validation Commands After Implementation

After creating files, do not run application tests unless needed because this phase does not touch Go code.

Run:

```bash
docker compose -f deploy/availability-lab/traefik-baseline/compose.yml config
```

If the OTP service is running on host port `8080`, run:

```bash
docker compose -f deploy/availability-lab/traefik-baseline/compose.yml up -d
curl -i http://localhost:8081/health
curl -I http://localhost:8082/dashboard/
docker compose -f deploy/availability-lab/traefik-baseline/compose.yml logs --tail=50 traefik
```

If OTP service is not running, do not fake success. Clearly say that runtime validation requires OTP service on `localhost:8080`.

## Output Required

After implementation, summarize:

1. Created files
2. Traefik topology
3. How to start it
4. How to test it
5. What was not implemented yet
6. Validation commands executed and results

---------

Analyze and then implement a small correction for the Traefik baseline availability lab.

Current implementation works, but we need to align it with the project workflow.

Requirements:
1. Keep Traefik as a separate availability lab compose file.
2. Do not move Traefik into the main docker-compose yet.
3. Update Traefik image from traefik:v3.2 to traefik:v3.7.1.
4. Add Makefile targets for the Traefik lab:
   - traefik-config
   - traefik-up
   - traefik-down
   - traefik-logs
   - traefik-ps
   - traefik-stack-up
   - traefik-stack-down
5. `traefik-up` should start only Traefik.
6. `traefik-stack-up` should start the normal backend stack first, then start Traefik.
7. `traefik-stack-down` should stop Traefik first, then stop the backend stack.
8. README should use Makefile commands as primary commands.
9. README should clearly document three modes:
   - direct backend mode: localhost:8080
   - Traefik gateway mode: localhost:8081
   - Traefik dashboard: localhost:8082
10. Keep PathPrefix(`/`) so all current and future backend routes pass through Traefik.
11. Do not add Nginx yet.
12. Do not add Keepalived yet.
13. Do not modify OTP application code.

After implementation:
- run docker compose config for Traefik lab
- run make availability-traefik-config
- summarize changed files
----------


Clean up the Traefik availability lab Makefile targets.

Current state:
- The Traefik lab works.
- Both long targets `availability-traefik-*` and short aliases `traefik-*` exist.
- I want to keep only the short `traefik-*` targets.
- Remove the `availability-traefik-*` targets entirely.
- Do not change lab behavior.
- Do not modify Traefik config files.
- Do not modify OTP application code.
- Do not modify Docker Compose behavior unless needed for target cleanup.

Keep these targets as the primary Makefile commands:

- traefik-config
- traefik-up
- traefik-down
- traefik-logs
- traefik-ps
- traefik-stack-up
- traefik-stack-down

Expected behavior:

- `make traefik-config`
  runs:
  `$(DOCKER_COMPOSE) -f $(TRAEFIK_LAB_COMPOSE) config`

- `make traefik-up`
  starts only the Traefik lab gateway.

- `make traefik-down`
  stops only the Traefik lab gateway.

- `make traefik-logs`
  follows Traefik lab logs.

- `make traefik-ps`
  shows Traefik lab containers.

- `make traefik-stack-up`
  starts the normal backend stack first using `docker-up`,
  then starts Traefik.

- `make traefik-stack-down`
  stops Traefik first,
  then stops the backend stack using `docker-down`.

Also update any README commands under:
`deploy/availability-lab/traefik-baseline/README.md`
so they use only the short `traefik-*` targets.

Do not keep aliases.
Do not keep duplicate target names.
Do not add new functionality.

After implementation:
- run `make traefik-config`
- run `make help | grep -i traefik` or equivalent
- summarize changed files and confirm only short targets remain.

-------
Update the root README.md to reflect the current project state.

Scope:
- Modify only README.md.
- Do not modify code.
- Do not modify Makefile.
- Do not modify compose files.
- Do not modify env.example.
- Do not modify docs files.

Current project state:
- The backend still supports direct access on localhost:8080.
- A Traefik single-gateway availability lab exists under:
  deploy/availability-lab/traefik-baseline/
- Traefik gateway access is:
  http://localhost:8081
- Traefik dashboard is:
  http://localhost:8082/dashboard/
- Traefik routes all paths to the backend using PathPrefix(`/`).
- Traefik adds:
  X-Gateway-Node: traefik-baseline
- The lab uses make targets:
  traefik-config
  traefik-up
  traefik-down
  traefik-logs
  traefik-ps
  traefik-stack-up
  traefik-stack-down

Required README updates:

1. Update project structure section:
   - Add deploy/availability-lab/traefik-baseline/
   - Mention it contains the local Traefik gateway availability lab.

2. Add a short section under Development or Makefile usage:
   "Availability Lab: Traefik Gateway"
   Include:
   - Direct backend mode: http://localhost:8080
   - Traefik gateway mode: http://localhost:8081
   - Traefik dashboard: http://localhost:8082/dashboard/

3. Add Makefile commands:
   - make traefik-config
   - make traefik-up
   - make traefik-down
   - make traefik-logs
   - make traefik-ps
   - make traefik-stack-up
   - make traefik-stack-down

4. Update API Endpoints section:
   - Keep /v1/otp/code if it still exists.
   - Add current OTP endpoints:
     - POST /v1/otp/send
     - POST /v1/otp/verify
   - Mention that OTP flow currently includes Redis state, fake SMS provider, request logging, verification logging, resend protection, and send rate limiting.
   - Do not over-document every internal detail; keep README concise and link to docs/current-state.md if it exists.

5. Update Environment Variables section:
   Add OTP-related env names at a high level:
   - OTP_CODE_LENGTH
   - OTP_TTL
   - OTP_MAX_ATTEMPTS
   - OTP_TENANT_CACHE_TTL
   - OTP_PROVIDER_TIMEOUT
   - OTP_FAKE_SMS_MIN_DELAY
   - OTP_FAKE_SMS_MAX_DELAY
   - OTP_FAKE_SMS_DEBUG_CODE_REDIS
   - OTP_FAKE_SMS_DEBUG_CODE_TTL
   - OTP_SEND_RATE_LIMIT_ENABLED
   - OTP_SEND_RATE_LIMIT_MAX
   - OTP_SEND_RATE_LIMIT_WINDOW

6. Keep README concise.
   Do not turn README into a full OTP architecture document.
   Detailed architecture belongs in docs/current-state.md / docs/architecture.md.

7. Preserve the current Persian style of the README.

After editing:
- Do not run Go tests because only README changed.
- Run no code changes.
- Summarize the updated sections.