# Task 13: Document k6 Concurrency Load Test for `/hello` Endpoint

## Objective
Create an Persian, production-style documentation describing how to run and analyze a **k6 concurrency load test (without sleep)** for the `/hello` endpoint.
This document is intended for **team reporting** and long-term reference.

---

## Deliverables

### 1. Documentation
Create the following file:

docs/LOAD_TESTING_K6_HELLO_CONCURRENCY.md

The document must be written in **Persian** and follow the same style as existing project tasks.

### 2. Load Test Script
Create the following file:

k6/03-hello-concurrency-no-sleep.js

---

## Documentation Structure (Required)

### 1. Introduction
- Purpose of concurrency load testing
- Why the `/hello` endpoint is used
- Difference between concurrency testing and RPS-based testing

### 2. Prerequisites
- k6 installed locally
- Service running locally
- Use the actual run commands from the project README

### 3. System Under Test (SUT)
Explain why system specs matter for result interpretation.

Include the following commands:
- lsb_release -a
- lscpu
- free -h

Provide a table for reporting system specifications.

### 4. Threshold Strategy
Explain:
- What k6 thresholds are
- Why thresholds are required
- Metrics:
  - http_req_failed (rate)
  - http_req_duration p95
  - http_req_duration p99
- Baseline → target → adjustment approach

### 5. Concurrency Load Test (No Sleep)
Explain:
- Virtual User behavior without sleep
- Why this test exposes contention and saturation
- What this test measures and what it does not

### 6. Test Script Description
Document the concurrency test script and its configuration.

### 7. How to Run the Test
Provide the exact k6 command.
Explain what to monitor during execution (CPU, RAM).

### 8. Capacity Limit Identification
Explain how to identify the system capacity limit using:
- Latency growth
- Error rate
- RPS plateau
- CPU saturation

### 9. Results – Maximum Observed Capacity
Provide a results table to be filled after execution.

### 10. Notes and Limitations
Explain limitations of local testing and client-side bottlenecks.

---

## Required k6 Script

Create k6/03-hello-concurrency-no-sleep.js with the following content:



---

## Acceptance Criteria
- Documentation is written in Persian
- Matches existing task style
- Uses actual project run instructions
- Ready for team sharing
- I want to put the result in this document

