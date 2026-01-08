
# Prometheus Metrics Setup Prompt

## Overview
You are a senior Go backend engineer. I want you to implement **Prometheus** metrics for the Go project to monitor the system’s performance and request metrics.

> **Important:** Some Docker/Docker Compose work may already be done in previous stages.  
> **First, audit the repository** and **only implement what is missing or incomplete**. If a requirement is already satisfied, do not redo it—just improve it if needed.

### Task:

1. **Prometheus Metrics Setup**:
   - Set up **Prometheus metrics** collection for the Go project.
   - Create a **metrics endpoint** (e.g., `/metrics`) where Prometheus can scrape the application’s metrics.

2. **Metrics to Include**:
   - **Request Latency**: Measure the response time for incoming HTTP requests. This will help monitor the latency of the API endpoints.
   - **Error Rates**: Track the rate of errors (e.g., 4xx and 5xx status codes).
   - **Request Count**: Track the total number of requests received by the server.
   - **Traffic Volume**: Track the amount of data (in bytes) sent and received by the server.

3. **Prometheus Integration**:
   - Expose the metrics in the standard **Prometheus format** so that Prometheus can scrape the data.
   - Integrate the Prometheus client library for Go (e.g., `github.com/prometheus/client_golang/prometheus`).

4. **Ensure No Conflicts**:
   - Ensure that **OpenTelemetry** for tracing does not interfere with the metrics collection from Prometheus.
   - Both **OpenTelemetry** tracing and **Prometheus metrics** should work independently but be used in parallel for comprehensive monitoring.

5. **Expected Output**:
   - A `/metrics` endpoint that exposes the Prometheus metrics in the correct format.
   - Metrics such as **latency**, **error rates**, **request count**, and **traffic volume** should be tracked.
   - The integration with Prometheus should allow Prometheus to scrape these metrics for monitoring.
