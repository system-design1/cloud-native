#!/bin/bash
# Health Checker Script
# Automatically calls health endpoints to generate traces in local development

INTERVAL=${HEALTH_CHECK_INTERVAL:-10}  # Default: 10 seconds

echo "Starting health checker (interval: ${INTERVAL}s)"
echo "Endpoints: /health, /ready, /live"
echo "Press Ctrl+C to stop"

while true; do
    # Call health endpoints
    curl -s http://localhost:8080/health > /dev/null 2>&1
    curl -s http://localhost:8080/ready > /dev/null 2>&1
    curl -s http://localhost:8080/live > /dev/null 2>&1
    
    sleep ${INTERVAL}
done

