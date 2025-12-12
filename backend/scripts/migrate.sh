#!/bin/sh

# migrate.sh - Database migration script using golang-migrate

set -e

# Default values
MIGRATIONS_DIR="migrations"
DB_HOST="${RELAY_DATABASE_HOST:-localhost}"
DB_PORT="${RELAY_DATABASE_PORT:-5432}"
DB_USER="${RELAY_DATABASE_USER:-relay}"
DB_PASSWORD="${RELAY_DATABASE_PASSWORD:-relay123}"
DB_NAME="${RELAY_DATABASE_DATABASE:-claude_relay}"
DB_SSLMODE="${RELAY_DATABASE_SSL_MODE:-disable}"

# Build database URL
DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"

# Command help
usage() {
    echo "Usage: $0 {up|down|version|force|drop} [steps]"
    echo ""
    echo "Commands:"
    echo "  up [N]       Apply all or N up migrations"
    echo "  down [N]     Apply all or N down migrations"
    echo "  version      Print current migration version"
    echo "  force V      Set version V but don't run migration (ignores dirty state)"
    echo "  drop         Drop everything inside database"
    echo ""
    echo "Environment Variables:"
    echo "  RELAY_DATABASE_HOST      Database host (default: localhost)"
    echo "  RELAY_DATABASE_PORT      Database port (default: 5432)"
    echo "  RELAY_DATABASE_USER      Database user (default: postgres)"
    echo "  RELAY_DATABASE_PASSWORD  Database password (default: postgres)"
    echo "  RELAY_DATABASE_DATABASE  Database name (default: claude_relay)"
    echo "  RELAY_DATABASE_SSL_MODE  SSL mode (default: disable)"
    exit 1
}

# Check if golang-migrate is installed
if ! command -v migrate >/dev/null 2>&1; then
    echo "Error: golang-migrate is not installed"
    echo "Please install it with:"
    echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    echo "Or on macOS:"
    echo "  brew install golang-migrate"
    exit 1
fi

# Parse command
COMMAND="${1:-}"
STEPS="${2:-}"

case "$COMMAND" in
    up)
        echo "Applying migrations up..."
        if [ -n "$STEPS" ]; then
            migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" up "$STEPS"
        else
            migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" up
        fi
        echo "Migrations applied successfully"
        ;;
    down)
        echo "Applying migrations down..."
        if [ -n "$STEPS" ]; then
            migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" down "$STEPS"
        else
            echo "Warning: This will rollback ALL migrations!"
            printf "Are you sure? (yes/no) "
            IFS= read -r REPLY
            echo
            case "$REPLY" in
                [Yy]es)
                    migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" down
                    ;;
                *)
                    echo "Cancelled"
                    exit 0
                    ;;
            esac
        fi
        echo "Migrations rolled back successfully"
        ;;
    version)
        echo "Current migration version:"
        migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" version
        ;;
    force)
        if [ -z "$STEPS" ]; then
            echo "Error: version number required for force command"
            usage
        fi
        echo "Forcing version to $STEPS..."
        migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" force "$STEPS"
        echo "Version forced successfully"
        ;;
    drop)
        echo "Warning: This will DROP ALL database objects!"
        printf "Are you sure? Type 'DROP' to confirm: "
        IFS= read -r REPLY
        echo
        if [ "$REPLY" = "DROP" ]; then
            migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" drop
            echo "Database dropped successfully"
        else
            echo "Cancelled"
            exit 0
        fi
        ;;
    *)
        echo "Error: Invalid command '$COMMAND'"
        usage
        ;;
esac
