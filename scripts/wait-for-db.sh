#!/bin/bash
# Wait for PostgreSQL database to be ready

set -e

DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-5432}"
MAX_ATTEMPTS=30
ATTEMPT=0

echo "Waiting for database at ${DB_HOST}:${DB_PORT} to be ready..."

while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
    # Try to connect using nc (netcat) or check if port is accessible
    if command -v nc >/dev/null 2>&1; then
        if nc -z "${DB_HOST}" "${DB_PORT}" 2>/dev/null; then
            echo "Database is ready!"
            exit 0
        fi
    elif command -v timeout >/dev/null 2>&1 && command -v bash >/dev/null 2>&1; then
        # Fallback: try to connect using bash's /dev/tcp
        if timeout 1 bash -c "echo > /dev/tcp/${DB_HOST}/${DB_PORT}" 2>/dev/null; then
            echo "Database is ready!"
            exit 0
        fi
    else
        # Last resort: just assume it's ready after a few seconds
        if [ $ATTEMPT -ge 3 ]; then
            echo "Database check tools not available, assuming database is ready..."
            exit 0
        fi
    fi
    
    ATTEMPT=$((ATTEMPT + 1))
    echo "Attempt $ATTEMPT/$MAX_ATTEMPTS: Database not ready, waiting 1 second..."
    sleep 1
done

echo "ERROR: Database at ${DB_HOST}:${DB_PORT} is not ready after $MAX_ATTEMPTS attempts"
echo "Make sure the database is running:"
echo "  - For docker-compose: docker-compose up -d postgres"
echo "  - For dev: make dev-db-up"
exit 1

