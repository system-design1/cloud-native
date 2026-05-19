# Traefik Baseline Availability Lab

## Purpose

This is Phase 1 of the OTP availability lab:

```text
Client -> Traefik -> OTP Service
```

This phase is not high availability yet. It introduces one Traefik gateway in front of the existing OTP service so later phases can add:

```text
Client -> VIP -> Traefik-1 / Traefik-2 -> OTP Service
```

with Keepalived.

## Prerequisite

The OTP service must already be running on the host on port `8080`.

There are three local modes to keep distinct:

```text
Direct backend mode:     http://localhost:8080
Traefik gateway mode:    http://localhost:8081
Traefik dashboard:       http://localhost:8082/dashboard/
```

Check the direct baseline first:

```bash
curl -i http://localhost:8080/health
```

## Start Traefik

Validate the Traefik lab compose file:

```bash
make traefik-config
```

Start only Traefik:

```bash
make traefik-up
```

Start the normal backend stack first, then Traefik:

```bash
make traefik-stack-up
```

Equivalent direct Docker Compose command from the repository root:

```bash
docker compose -f deploy/availability-lab/traefik-baseline/compose.yml up -d
```

## Test Gateway

```bash
curl -i http://localhost:8081/health
```

Expected:

- same health response as direct OTP
- response includes:

```text
X-Gateway-Node: traefik-baseline
```

## Open Traefik Dashboard

```text
http://localhost:8082/dashboard/
```

This lab uses `api.insecure=true` for local demo convenience. That dashboard configuration is not production-safe.

## Prove Traffic Goes Through Traefik

Use these signals:

- response header `X-Gateway-Node`
- Traefik access logs
- dashboard router/service visibility

Follow Traefik logs:

```bash
make traefik-logs
```

Show Traefik lab containers:

```bash
make traefik-ps
```

## Simple Traffic Loop

```bash
while true; do
  date +"%H:%M:%S"
  curl -s -o /dev/null -w "status=%{http_code} time=%{time_total}\\n" http://localhost:8081/health
  sleep 0.5
done
```

## Failure Demo

1. Start OTP service.
2. Start Traefik.
3. Start the traffic loop.
4. Stop OTP service.
5. Expected:
   - Traefik returns `502` or `503`.
   - Access logs show upstream failure.
6. Restart OTP service.
7. Expected:
   - gateway requests recover.

Traefik stop test:

```bash
make traefik-down
```

Expected:

- `localhost:8081` becomes unavailable.
- direct `localhost:8080/health` still works if OTP service is running.

To stop Traefik first and then the normal backend stack:

```bash
make traefik-stack-down
```

## What This Phase Proves

- Traefik can act as a gateway in front of OTP.
- Traffic can be observed through Traefik.
- A single gateway is still a single point of failure.
- This motivates the next Keepalived HA phase.

## Deferred

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
