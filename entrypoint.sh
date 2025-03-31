#!/bin/sh
set -e

echo "Waiting for Postgres at ${POSTGRES_HOST}:${POSTGRES_PORT}..."
until pg_isready -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER"; do
  echo "Postgres is unavailable - sleeping 2s..."
  sleep 2
done

echo "Postgres is available. Running migrations..."

# Run migrations using the migrate CLI; assumes migration files are at /app/migrations
migrate -path /app/migrations -database "$DB_SOURCE" up || echo "Migrations already up-to-date or no changes needed."

echo "Starting application..."
exec "$@" 