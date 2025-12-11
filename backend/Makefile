.PHONY: help build run test lint clean docker-build docker-up docker-down migrate-up migrate-down migrate-version migrate-force migrate-drop

help:
	@echo "Available commands:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make lint           - Run linter"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-up      - Start Docker Compose"
	@echo "  make docker-down    - Stop Docker Compose"
	@echo "  make migrate-up     - Apply all database migrations"
	@echo "  make migrate-down   - Rollback database migrations"
	@echo "  make migrate-version - Show current migration version"
	@echo "  make migrate-force  - Force set migration version (use with care)"
	@echo "  make migrate-drop   - Drop all database objects (dangerous)"

build:
	@echo "Building..."
	@go build -o bin/server cmd/server/main.go

run:
	@echo "Running..."
	@go run cmd/server/main.go -config config/config.yaml

test:
	@echo "Testing..."
	@go test -v -cover ./...

test-coverage:
	@echo "Testing with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint:
	@echo "Linting..."
	@golangci-lint run

clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf tmp/
	@rm -rf logs/
	@rm -f coverage.out

docker-build:
	@echo "Building Docker image..."
	@docker build -t claude-relay-go:latest .

docker-up:
	@echo "Starting Docker Compose..."
	@docker-compose up -d

docker-down:
	@echo "Stopping Docker Compose..."
	@docker-compose down

migrate-up:
	@echo "Applying database migrations..."
	@bash scripts/migrate.sh up

migrate-down:
	@echo "Rolling back database migrations..."
	@bash scripts/migrate.sh down 1

migrate-version:
	@echo "Current migration version:"
	@bash scripts/migrate.sh version

migrate-force:
	@echo "Forcing migration version (use with VERSION=N)..."
	@bash scripts/migrate.sh force $(VERSION)

migrate-drop:
	@echo "Dropping all database objects..."
	@bash scripts/migrate.sh drop
