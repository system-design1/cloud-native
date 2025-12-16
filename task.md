
# Project Setup for Go Backend Service

## Overview
You are a senior Go backend engineer. I want to create the initial setup for a Go project with the following specifications:

- **Language**: Go
- **Framework**: Gin
- **Project type**: REST API
- **Architecture**: Monolithic Modular
- **Database**: PostgreSQL
- **Authentication**: JWT
- **Logging**: Zerolog (Structured JSON Logging with Correlation IDs)
- **Configuration**: Externalized configuration using environment variables with validation (No hardcoded values)

## Task:

### 1. Project Structure:
- Create the initial structure for the project with the following directories:
  - `cmd/`: For the application's entry point.
  - `pkg/`: For shared libraries and packages.
  - `internal/`: For internal application code.
  - `configs/`: For configuration files (environment-based).

### 2. Go Modules:
- Initialize the Go project with Go Modules (run `go mod init`).

### 3. Environment-Based Configuration:
- Implement configuration loading using environment variables (use `os.Getenv` or a config management library like **Koanf**).
- Validate the environment variables (ensure no values are hardcoded).

### 4. Gin Setup:
- Set up a basic **Gin** HTTP server with a simple `GET` route (`/hello`) that returns "Hello, World!" as a response.

### 5. Dockerfile:
- Create a multi-stage **Dockerfile** for building and running the app:
  - Use a Go base image for building the app.
  - Copy the app into the container and use a minimal image for running the app.
- Set up a non-root user to run the application inside the container.

### 6. Docker Compose:
- Create a basic `docker-compose.yml` file for local development:
  - Add a PostgreSQL service.
  - Map ports for the API and the database.
  - Make sure to define environment variables for configuration in the `.env` file.

### 7. .dockerignore:
- Set up a `.dockerignore` file to ignore unnecessary files during the Docker build process (e.g., `node_modules`, `.git`, etc.).

### 8. Create makefile
#### Overview
You are a senior Go backend engineer. I want you to create a **Makefile** for the Go project with the following requirements:

#### Requirements:

1. **Build the Go Project**:
   - A target to build the Go project and output the binary with the project name.

2. **Run the Application**:
   - A target to run the Go project directly without Docker using `go run`.

3. **Docker Build**:
   - A target to build the Docker image for the Go project using the `Dockerfile`.

4. **Docker Compose**:
   - A target to start the Docker containers with `docker-compose up -d`.
   - A target to stop the Docker containers with `docker-compose down`.

5. **Clean**:
   - A target to clean the project by removing the generated binary file.

6. **Testing**:
   - A target to run the Go tests using `go test`.

7. **Swagger Documentation (Optional)**:
   - If using Swagger for API documentation, include a target to generate the Swagger spec file.

#### Additional Information:
- **Docker Variables**: Use `docker` and `docker-compose` for Docker-related targets.
- **Go Variables**: Use `go` for the Go build and test commands.
- Use the default project name `go-backend-service` for the binary output.
- Use `make` commands to perform the above operations.

Please ensure the file is well-commented, easy to understand, and follows best practices.

## Expected Output:
- A Go project with the above structure.
- A working Gin-based server running inside a Docker container.
- Proper environment-based configuration and validation.