# Multi-stage build for Claude Relay Go Backend

# Stage 1: Builder
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o server \
    ./cmd/server

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Copy binary from builder
COPY --from=builder /app/server /app/server

# Copy configuration and migrations
COPY --from=builder /app/config /app/config
COPY --from=builder /app/migrations /app/migrations

# Change ownership
RUN chown -R appuser:appuser /app

USER appuser

# Expose application port
EXPOSE 8080

# Health check (assumes /health endpoint exists or will be added)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["/app/server", "-config", "/app/config/config.yaml"]
