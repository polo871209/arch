#!/bin/bash

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
timeout=30
while [ $timeout -gt 0 ]; do
    if docker-compose exec -T postgres pg_isready -U rpc_user -d rpc_dev >/dev/null 2>&1; then
        echo "PostgreSQL is ready!"
        exit 0
    fi
    sleep 1
    timeout=$((timeout-1))
done

echo "❌ PostgreSQL failed to start within 30 seconds"
exit 1