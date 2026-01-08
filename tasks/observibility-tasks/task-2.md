# Section 2: Middleware and Logging Setup Prompt

## Overview
You are a senior Go backend engineer. I want you to create the following middleware and logging setup for the Go project:

### Task:

1. **Request/Response Logging Middleware**:
   - Create a middleware for logging incoming HTTP requests and outgoing responses using **Zerolog**.
   - The logs should be structured as JSON and should include:
     - **Correlation ID** for each request (to track the request flow across services).
     - HTTP method (e.g., `GET`, `POST`, etc.).
     - Request URL.
     - Response status code.
     - Response time (latency).
   - The logs should be written to `stdout` in **structured JSON format**.

2. **Global Error Handler Middleware**:
   - Create a global error handler middleware that:
     - Catches all errors in the application and sends a consistent error response format.
     - If an error occurs, the middleware should:
       - Log the error using **Zerolog** with relevant details.
       - Return a standardized error response with an appropriate HTTP status code.
       - Ensure no sensitive data is exposed in the error responses.

3. **Zerolog Setup**:
   - Configure **Zerolog** for logging.
   - Use **JSON format** for all logs and ensure logs include a **correlation ID** for tracing.
   - Use **log.Level.Info**, **log.Level.Error**, etc., to log messages with the appropriate severity.
   - Logs should be output to **stdout**.

4. **Middleware Chain**:
   - Apply the middleware to the Gin router.
   - Ensure that the logging middleware is applied to all routes and that errors are handled globally.

### Expected Output:
- A **Zerolog**-based logging middleware for request/response logging.
- A **global error handler** that catches errors and returns consistent error responses.
- Logs are structured in JSON format with correlation IDs, written to `stdout`.