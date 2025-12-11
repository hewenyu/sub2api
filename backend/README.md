# Claude Relay Go - Backend

High-performance AI API relay service supporting Claude Code and Codex.

## Features

- Support for Claude Code and Codex relay
- Intelligent account scheduling and load balancing
- Detailed usage statistics and cost calculation
- Comprehensive rate limiting and quota management
- User-friendly web management interface

## Tech Stack

- Go 1.21+
- Gin Web Framework
- GORM ORM
- PostgreSQL 15+
- Redis 7+
- Zap Logger
- Viper Configuration

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 15 or higher
- Redis 7 or higher
- Docker and Docker Compose (optional)

## Quick Start

### Local Development

1. Clone the repository
```bash
git clone https://github.com/Wei-Shaw/sub2api.git
cd claude-relay-go/backend
```

2. Install dependencies
```bash
go mod download
```

3. Configure environment
```bash
cp config/config.example.yaml config/config.yaml
# Edit config/config.yaml to update database, Redis, and other settings
```

4. Build the application
```bash
make build
```

5. Run the service
```bash
make run
```

The server will start on port 8080 by default. You can verify it's running by visiting:
```bash
curl http://localhost:8080/health
```

### Docker Compose Deployment

For a complete development environment with PostgreSQL and Redis:

```bash
docker-compose up -d
```

This will start:
- PostgreSQL 15 on port 5432
- Redis 7 on port 6379

## Development

### Hot Reload

For development with hot reload:

```bash
# Install air if not already installed
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

### Run Tests

```bash
make test
```

### Code Linting

```bash
# Install golangci-lint if not already installed
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
make lint
```

### Build Docker Image

```bash
make docker-build
```

## Project Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── config/
│   ├── config.go                # Configuration structs and loader
│   ├── config.yaml              # Development configuration
│   └── config.example.yaml      # Example configuration template
├── pkg/
│   └── logger/
│       └── logger.go            # Logging utility
├── internal/                    # Internal packages (for future implementation)
│   ├── api/                     # API handlers
│   ├── middleware/              # HTTP middlewares
│   ├── service/                 # Business logic
│   ├── repository/              # Data access layer
│   ├── model/                   # Data models
│   └── util/                    # Utility functions
├── migrations/                  # Database migrations
├── scripts/                     # Utility scripts
├── Makefile                     # Build and task automation
├── Dockerfile                   # Docker image definition
├── docker-compose.yml           # Docker Compose configuration
└── .air.toml                    # Air hot-reload configuration
```

## Configuration

The application uses YAML configuration files. All configuration values can be overridden using environment variables with the `RELAY_` prefix.

Example:
```bash
export RELAY_SERVER_PORT=9090
export RELAY_DATABASE_HOST=localhost
export RELAY_SECURITY_JWT_SECRET=my-secret-key
```

See `config/config.example.yaml` for all available configuration options.

## Available Commands

- `make help` - Show available commands
- `make build` - Build the application binary
- `make run` - Run the application
- `make test` - Run tests with coverage
- `make lint` - Run code linting
- `make clean` - Clean build artifacts
- `make docker-build` - Build Docker image
- `make docker-up` - Start Docker Compose services
- `make docker-down` - Stop Docker Compose services

## API Endpoints

### Health Check
- `GET /health` - Returns server health status

More endpoints will be added as development progresses.

## License

MIT
