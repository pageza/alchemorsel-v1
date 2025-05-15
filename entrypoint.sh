#!/bin/sh
set -e

# Load JWT_SECRET from Docker secrets if not set
if [ -z "$JWT_SECRET" ] && [ -f "/run/secrets/jwt_secret" ]; then
  export JWT_SECRET=$(cat /run/secrets/jwt_secret)
fi

# Load PGUSER from Docker secrets if not set
if [ -z "$PGUSER" ] && [ -f "/run/secrets/postgres_user" ]; then
  export PGUSER=$(cat /run/secrets/postgres_user)
fi

# Load PGPASSWORD from Docker secrets if not set
if [ -z "$PGPASSWORD" ] && [ -f "/run/secrets/postgres_password" ]; then
  export PGPASSWORD=$(cat /run/secrets/postgres_password)
fi

# Load POSTGRES_USER from Docker secrets if not set
if [ -f /run/secrets/postgres_user ]; then
    export POSTGRES_USER=$(cat /run/secrets/postgres_user)
fi

# Load POSTGRES_PASSWORD from Docker secrets if not set
if [ -f /run/secrets/postgres_password ]; then
    export POSTGRES_PASSWORD=$(cat /run/secrets/postgres_password)
fi

# Log the environment variables
echo "Starting app with PGHOST=$PGHOST, PGPORT=$PGPORT, PGDATABASE=$PGDATABASE, PGUSER=$PGUSER, JWT_SECRET is set: ${JWT_SECRET:+yes}"

# Build DSN from environment variables if DB_SOURCE is not set
if [ -z "$DB_SOURCE" ]; then
  [ -z "$DB_HOST" ] && DB_HOST="$PGHOST"
  [ -z "$DB_PORT" ] && DB_PORT="$PGPORT"
  [ -z "$DB_USER" ] && DB_USER="$PGUSER"
  [ -z "$DB_PASSWORD" ] && DB_PASSWORD="$PGPASSWORD"
  [ -z "$DB_DATABASE" ] && DB_DATABASE="$PGDATABASE"
  export DB_SOURCE="host=${DB_HOST} port=${DB_PORT} user=${DB_USER} password=${DB_PASSWORD} dbname=${DB_DATABASE} sslmode=disable"
fi

echo "Using DB_SOURCE: $DB_SOURCE"

# Execute the main binary
exec /app/main 