#!/bin/sh
set -e

echo "Waiting for PostgreSQL to be ready..."

until pg_isready -h "${RELAY_DATABASE_HOST:-postgres}" -U "${RELAY_DATABASE_USER:-relay}" -d "${RELAY_DATABASE_DATABASE:-claude_relay}" 2>/dev/null; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 2
done

echo "PostgreSQL is ready - running migrations"

cd /init
./migrate.sh up

echo "Database migrations completed successfully"
