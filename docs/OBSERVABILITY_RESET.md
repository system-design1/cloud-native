# Reset Observability Stack and Recover Jaeger During Load Tests

## Background / Problem Statement

When running high-throughput k6 load tests (e.g., 10k–13k RPS) against high-traffic endpoints such as `/hello`, the Jaeger UI can become unresponsive (endless loading spinner) and traces may not load.

This typically happens because trace volume is too high:
- Excessive tracing for high-volume routes overwhelms the tracing pipeline (OTel exporter/collector + trace storage).
- Jaeger UI queries can time out or become very slow under heavy ingestion and large datasets.

**Data retention is not important** for this scenario. The primary goal is to quickly return the tracing stack (Jaeger/Tempo/etc.) to a clean, working state.

---

## Solution: Reset Observability Stack

The project provides a simple Makefile target to reset the observability stack:

```bash
make observability-reset
```

This command:
1. **Stops** the observability stack (Tempo, Jaeger, Prometheus, Grafana, Loki, Promtail)
2. **Deletes all trace storage data** (volumes containing traces, metrics, dashboards)
3. **Starts** the observability stack again from scratch

This allows Jaeger to load normally after heavy load tests.

---

## What Gets Deleted

The reset operation removes **all data** stored in observability volumes:
- **Traces**: All traces stored in Tempo and Jaeger
- **Metrics**: All Prometheus metrics data
- **Dashboards**: All Grafana dashboards and configurations (if stored in volumes)
- **Logs**: All Loki log data

**⚠️ WARNING**: This operation **permanently deletes all observability data**. Use this only when you need to recover from an overloaded state during load testing or development.

---

## How to Use

### Step 1: Run the Reset Command

```bash
make observability-reset
```

**Output:**
```
WARNING: This will delete all observability data (traces, metrics, dashboards)
Resetting observability stack...
Stopping observability stack and removing volumes...
Starting observability stack from scratch...
...
Observability stack reset complete!
```

### Step 2: Verify Recovery

After the reset completes, verify that everything is working:

1. **Check containers are running:**
   ```bash
   docker ps | grep -E "tempo|jaeger|prometheus|grafana|loki"
   ```

2. **Open Jaeger UI:**
   - Navigate to: http://localhost:16686
   - The UI should load normally (no endless spinner)

3. **Run a small request and verify traces appear:**
   ```bash
   # For routes that are sampled (e.g., /test-error if sampled at 100%)
   curl http://localhost:8080/test-error
   ```
   - Check Jaeger UI: You should see traces for `/test-error` appear within a few seconds.

4. **Verify Prometheus is collecting metrics:**
   - Navigate to: http://localhost:9090
   - Run a query like: `up{job="go-backend-service"}`
   - You should see metrics being collected

---

## When to Use This

Use `make observability-reset` when:

- ✅ **Jaeger UI is unresponsive** after high-load k6 tests
- ✅ **Traces are not loading** in Jaeger UI (endless spinner)
- ✅ **You need a clean slate** for testing or development
- ✅ **Observability stack is slow** due to excessive data volume

**Do NOT use this in production** unless you understand the consequences of losing all observability data.

---

## Alternative: Use Sampling Policy Instead

Instead of resetting after load tests, you can **prevent** the problem by using route-based sampling policy during load tests. This reduces trace volume for high-traffic endpoints while keeping full tracing for important routes.

See [LOAD_TESTING_TRACING_SAMPLING.md](./LOAD_TESTING_TRACING_SAMPLING.md) for details on how to configure sampling policies for load testing.

**Recommended approach:**
1. **Before load test**: Apply load-test sampling policy (e.g., DROP for `/hello`, ALWAYS for `/test-error`)
2. **During load test**: Run your k6 test
3. **After load test**: Reset observability stack if needed, or revert to normal sampling policy

---

## Impact on Prometheus Metrics

**Important**: The reset operation deletes Prometheus metrics data stored in volumes. However, this does **not** affect the application's ability to expose metrics.

- ✅ **Metrics collection continues**: The application continues to expose metrics at `/metrics`
- ✅ **Prometheus resumes scraping**: After reset, Prometheus immediately starts collecting new metrics
- ✅ **No application impact**: The reset only affects the observability stack, not your application

**Note**: OTEL trace sampling changes (e.g., DROP for `/hello`) do **not** affect Prometheus metrics. Metrics collection is independent from trace sampling:
- Prometheus scrapes `/metrics` endpoint regardless of trace sampling policy
- Dropping/low-sampling `/hello` traces is safe for Prometheus dashboards and counters
- Metrics are always collected, even if traces are dropped

---

## Manual Reset (Alternative)

If you prefer to run the commands manually:

```bash
# Stop and remove volumes
docker compose -f docker-compose.observability.yml down -v

# Start again
docker compose -f docker-compose.observability.yml up -d
```

---

## Related Documentation

- [LOAD_TESTING_TRACING_SAMPLING.md](./LOAD_TESTING_TRACING_SAMPLING.md): How to prevent Jaeger overload by adjusting sampling during load tests
- [OBSERVABILITY.md](./OBSERVABILITY.md): Complete observability stack documentation
- [TEMPO_GUIDE.md](./TEMPO_GUIDE.md): Tempo tracing backend guide

---

## Troubleshooting

### Reset command fails

**Problem**: `make observability-reset` fails with permission errors.

**Solution**: Ensure Docker has proper permissions:
```bash
# Check Docker is running
docker ps

# If permission denied, add user to docker group (Linux)
sudo usermod -aG docker $USER
# Then log out and log back in
```

### Containers don't start after reset

**Problem**: After reset, containers fail to start.

**Solution**: Check logs and ensure ports are not in use:
```bash
# Check logs
make observability-logs

# Check if ports are in use
lsof -i :16686  # Jaeger
lsof -i :9090   # Prometheus
lsof -i :3000   # Grafana
```

### Jaeger still slow after reset

**Problem**: Jaeger UI is still slow even after reset.

**Solution**: 
1. Ensure the reset completed successfully (check `docker ps`)
2. Wait a few seconds for containers to fully start
3. Check if your application is still generating excessive traces (apply sampling policy)
4. Verify you're not running multiple load tests simultaneously

