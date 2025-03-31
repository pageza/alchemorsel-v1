#!/bin/bash
set -e

# This script is executed by the PostgreSQL Docker entrypoint if the database is new.
# It creates an additional database called 'ephemeral_db'.

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE ephemeral_db;
EOSQL 