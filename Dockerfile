# Build stage
# Pin Go version for reproducibility
FROM golang:1.25-alpine AS builder

# Install build dependencies
# Use retry logic with multiple mirrors to handle network issues
RUN for mirror in http://dl-cdn.alpinelinux.org/alpine http://mirror1.hs-esslingen.de/pub/Mirrors/alpine http://alpine.mirror.far.fi; do \
        echo "Trying mirror: $mirror" && \
        sed -i "s|http://dl-cdn.alpinelinux.org/alpine|$mirror|g" /etc/apk/repositories 2>/dev/null || true && \
        apk update --no-cache && \
        apk add --no-cache git make && \
        break || continue; \
    done || \
    (echo "All mirrors failed, trying default" && \
     apk update --no-cache && \
     apk add --no-cache git make)

# Set working directory
WORKDIR /build

# Set GOPROXY with multiple mirrors for Go modules
# Try Chinese mirror first (usually faster and more reliable), then direct
ENV GOPROXY=https://goproxy.cn,direct,https://proxy.golang.org,https://goproxy.io

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies (this layer will be cached unless go.mod/go.sum change)
# Using BuildKit cache mount for better performance - cache will persist across builds
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download && \
    go mod verify

# Copy source code (this will invalidate cache only when source changes)
COPY . .

# Build the application (using cache for go build cache as well)
# -trimpath: removes file system paths from the compiled binary for reproducibility
# -ldflags="-s -w": strips debug symbols and reduces binary size
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -a -installsuffix cgo -o go-backend-service ./cmd/server

# Final stage
# Pin Alpine version for reproducibility and security
FROM alpine:3.20

# Install ca-certificates and wget for healthcheck
# Use retry logic with multiple mirrors to handle network issues
RUN for mirror in http://dl-cdn.alpinelinux.org/alpine http://mirror1.hs-esslingen.de/pub/Mirrors/alpine http://alpine.mirror.far.fi; do \
        echo "Trying mirror: $mirror" && \
        sed -i "s|http://dl-cdn.alpinelinux.org/alpine|$mirror|g" /etc/apk/repositories 2>/dev/null || true && \
        apk update --no-cache && \
        apk add --no-cache ca-certificates wget && \
        break || continue; \
    done || \
    (echo "All mirrors failed, trying default" && \
     apk update --no-cache && \
     apk add --no-cache ca-certificates wget) && \
    rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/go-backend-service .

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check - checks the /health endpoint
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Use ENTRYPOINT exec-form for proper signal handling (SIGTERM)
# This ensures the application receives signals correctly for graceful shutdown
ENTRYPOINT ["./go-backend-service"]

