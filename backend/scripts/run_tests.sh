#!/bin/bash
set -e

# Start test database
echo "Starting test database..."
docker compose -f docker-compose.test.yml up -d

# Wait for database to be ready
echo "Waiting for database to be ready..."
until docker compose -f docker-compose.test.yml exec -T test-db pg_isready -U postgres; do
    echo "Database is not ready... waiting"
    sleep 2
done

# Run tests
echo "Running tests..."
go test ./... -v

# Cleanup
echo "Cleaning up..."
docker compose -f docker-compose.test.yml down

echo "Done!"
