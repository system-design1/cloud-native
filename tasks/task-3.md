
# API and Endpoints Creation Prompt

## Overview
You are a senior Go backend engineer. I want you to create the following API and endpoints for the Go project:

### Task:

1. **Create Two Endpoints**:
   - **`/hello`**:
     - Create a simple `GET` endpoint that responds with a static message, such as "Hello, World!".

   - **`/delayed-hello`**:
     - Create a `GET` endpoint that introduces a random delay (between 1 and 3 seconds) before responding with the message "Hello after delay".
     - The response message should include the time delay (e.g., "Hello after delay: 2.34 seconds").

2. **Random Delay**:
   - For the `/delayed-hello` endpoint, use Go's `math/rand` package to generate a random delay time between 1 and 3 seconds.
   - Ensure that the response includes the delay time.

3. **Logging**:
   - Make sure that both endpoints are logged properly using the **Zerolog** middleware (which we set up in the previous step).
   - Log relevant information, such as the HTTP method, the response status, the URL, and the response time.

4. **Expected Output**:
   - Two fully functional endpoints:
     - `/hello` returning a static "Hello, World!" message.
     - `/delayed-hello` returning a message with a random delay and the exact time delay in the response.
