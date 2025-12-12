# Claude Relay Docker Compose Deployment

Production-ready Docker Compose setup for Claude Relay application.

## Quick Start

1. Copy the environment template:
```bash
cp .env.example .env
```

2. Edit `.env` and configure required variables:
```bash
# REQUIRED: Generate secure secrets
JWT_SECRET=your-secure-jwt-secret-min-32-characters
ENCRYPTION_KEY=your-32-character-encryption-key

# REQUIRED: Claude OAuth credentials
CLAUDE_CLIENT_ID=your-claude-client-id
CLAUDE_CLIENT_SECRET=your-claude-client-secret
CLAUDE_REDIRECT_URI=http://your-domain.com/callback

# REQUIRED: Codex OAuth credentials
CODEX_CLIENT_SECRET=your-codex-client-secret
```

3. Prepare configuration files:
```bash
cd config
cp config.example.yaml config.yaml
cp model_prices_and_context_window.example.json model_prices_and_context_window.json
```

Edit `config/config.yaml` with your settings. The config files will be mounted into the backend container.

4. Start all services:
```bash
docker-compose up -d
```

The init container will automatically run database initialization before the backend starts.

5. Check service health:
```bash
docker-compose ps
```

## Initial Setup

### Database Initialization

The init container automatically handles database initialization when you start the services. It:
- Waits for PostgreSQL to be ready
- Runs all database migrations using golang-migrate
- Creates all required tables, indexes, and constraints

The backend service will only start after the init container completes successfully.

### Create Admin User

Connect to the backend container and create an admin user:

```bash
docker-compose exec backend /bin/sh
# Inside container, use the admin creation tool or API
```

Or use the API directly:
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "secure-password"
  }'
```

## Service URLs

- Frontend: http://localhost:80
- Backend API: http://localhost:8080
- Backend Health: http://localhost:8080/health
- Metrics: http://localhost:9090/metrics
- PostgreSQL: localhost:5432
- Redis: localhost:6379

## Environment Variables

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| JWT_SECRET | JWT signing secret (min 32 chars) | `your-secure-secret-key-here` |
| ENCRYPTION_KEY | Data encryption key (exactly 32 chars) | `12345678901234567890123456789012` |
| CLAUDE_CLIENT_ID | Claude OAuth client ID | `your-client-id` |
| CLAUDE_CLIENT_SECRET | Claude OAuth client secret | `your-client-secret` |
| CODEX_CLIENT_SECRET | Codex OAuth client secret | `your-codex-secret` |

### Optional Variables

| Variable | Default | Description |
|----------|---------|-------------|
| DB_USER | relay | Database username |
| DB_PASSWORD | relay123 | Database password |
| DB_NAME | claude_relay | Database name |
| BACKEND_PORT | 8080 | Backend service port |
| FRONTEND_PORT | 80 | Frontend service port |
| SERVER_MODE | release | Server mode (debug/release) |
| LOG_LEVEL | info | Logging level |

## Configuration Files

The backend service uses configuration files mounted from the `config/` directory:

- `config/config.yaml`: Main application configuration
- `config/model_prices_and_context_window.json`: Model pricing and context window settings

These files are mounted as volumes and can be updated without rebuilding the container. After updating, restart the backend service:

```bash
docker-compose restart backend
```

For local调试（更详细日志），你也可以使用提供的 `config.debug.yaml` 作为模板：

```bash
cd config
cp config.debug.yaml config.yaml
```

`config.example.yaml` 默认将日志写入容器内的 `./logs/claude_relay.log` 和 `./logs/claude_relay_error.log` 文件；
`config.debug.yaml` 默认将日志写入容器内的 `./logs/claude_relay_debug.log` 和 `./logs/claude_relay_debug_error.log` 文件，并开启更详细的 debug 日志和 payload 日志。

## Data Persistence

Data is persisted in Docker volumes:
- `postgres_data`: PostgreSQL database files
- `redis_data`: Redis data files

To backup data:
```bash
docker-compose exec postgres pg_dump -U relay claude_relay > backup.sql
```

To restore data:
```bash
docker-compose exec -T postgres psql -U relay claude_relay < backup.sql
```

## Logs

View logs for all services:
```bash
docker-compose logs -f
```

View logs for specific service:
```bash
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f postgres
docker-compose logs -f redis
docker-compose logs init  # View init container logs (one-time execution)
```

## Maintenance

### Update Services

Pull latest images and restart:
```bash
docker-compose pull
docker-compose up -d
```

### Restart Services

```bash
docker-compose restart
```

### Stop Services

```bash
docker-compose stop
```

### Remove Everything (including volumes)

```bash
docker-compose down -v
```

## Troubleshooting

### Backend won't start

1. Check if PostgreSQL is healthy:
```bash
docker-compose ps postgres
```

2. Check backend logs:
```bash
docker-compose logs backend
```

3. Verify database connection:
```bash
docker-compose exec postgres psql -U relay -d claude_relay -c "SELECT 1;"
```

### Database connection errors

1. Ensure PostgreSQL is running and healthy
2. Verify credentials in `.env` file
3. Check if migrations have been run
4. Review backend logs for specific errors

### Redis connection errors

1. Check Redis health:
```bash
docker-compose exec redis redis-cli ping
```

2. If password is set, test with password:
```bash
docker-compose exec redis redis-cli -a your-password ping
```

### Frontend can't connect to backend

1. Verify `FRONTEND_API_URL` in `.env` matches your backend URL
2. Check if backend is healthy: `curl http://localhost:8080/health`
3. Review frontend logs: `docker-compose logs frontend`

### Port conflicts

If ports are already in use, change them in `.env`:
```bash
BACKEND_PORT=8081
FRONTEND_PORT=8080
DB_PORT=5433
REDIS_PORT=6380
```

### Permission errors

Ensure the user running docker-compose has proper permissions:
```bash
sudo usermod -aG docker $USER
newgrp docker
```

## Security Recommendations

1. Change all default passwords in `.env`
2. Use strong JWT_SECRET (min 32 characters)
3. Use proper ENCRYPTION_KEY (exactly 32 characters)
4. Enable HTTPS in production (use reverse proxy like nginx)
5. Restrict database access to backend only
6. Use Docker secrets for sensitive data in production
7. Regularly update Docker images
8. Enable firewall rules to restrict access

## Production Deployment

For production deployment:

1. Use a reverse proxy (nginx/traefik) with SSL/TLS
2. Set `SERVER_MODE=release`
3. Use strong, unique secrets
4. Configure proper backup strategy
5. Set up monitoring and alerting
6. Use Docker secrets instead of environment variables
7. Implement rate limiting at proxy level
8. Configure log aggregation
9. Set up health check monitoring
10. Use managed database services for better reliability

## Support

For issues and questions:
- Check logs: `docker-compose logs`
- Review backend documentation: `../backend/README.md`
- Check GitHub issues
