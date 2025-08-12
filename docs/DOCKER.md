# Docker Deployment Guide

This document explains how to deploy and manage the Print Service using Docker Compose.

## Quick Start

### Development Environment

1. **Copy environment file:**
   ```bash
   cp .env.example .env
   ```

2. **Start all services:**
   ```bash
   docker-compose up -d
   ```

3. **View logs:**
   ```bash
   docker-compose logs -f
   ```

4. **Access services:**
   - Print Service API: http://localhost:8080
   - Redis Commander: http://localhost:8081

### Production Environment

1. **Set up environment:**
   ```bash
   cp .env.example .env
   # Edit .env with production values
   export CONFIG_ENV=production
   export REDIS_PASSWORD=your-secure-password
   ```

2. **Deploy with production config:**
   ```bash
   docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
   ```

## Architecture

### Services

- **print-server**: Main API server handling HTTP requests
- **print-worker**: Background worker processing print jobs
- **redis**: Queue and cache storage
- **redis-commander**: Redis management UI (development only)

### Volumes

- **redis_data**: Persistent Redis data
- **print_output**: Generated print files
- **print_temp**: Temporary processing files
- **print_logs**: Application logs

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CONFIG_ENV` | Environment (development/production) | `development` |
| `VERSION` | Application version | `dev` |
| `REDIS_URL` | Redis connection URL | `redis://redis:6379` |
| `SERVER_PORT` | Server port | `8080` |

### Configuration Files

- `configs/development.yaml`: Development settings
- `configs/production.yaml`: Production settings

## Management Commands

### Development

```bash
# Start services
make docker-up

# Stop services
make docker-down

# View logs
make docker-logs

# Rebuild and restart
make docker-rebuild

# Clean up
make docker-clean
```

### Production

```bash
# Deploy production
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Scale workers
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d --scale print-worker=5

# Update services
docker-compose -f docker-compose.yml -f docker-compose.prod.yml pull
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

## Monitoring

### Health Checks

All services include health checks:
- **Server**: HTTP health endpoint
- **Worker**: Process monitoring
- **Redis**: Redis ping command

### Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f print-server

# Follow with timestamps
docker-compose logs -f -t
```

### Metrics (Production)

Enable monitoring profile for Prometheus:
```bash
docker-compose -f docker-compose.yml -f docker-compose.prod.yml --profile monitoring up -d
```

## Troubleshooting

### Common Issues

1. **Port conflicts:**
   ```bash
   # Check port usage
   netstat -tulpn | grep :8080
   
   # Change ports in .env
   SERVER_PORT=8081
   ```

2. **Redis connection issues:**
   ```bash
   # Check Redis health
   docker-compose exec redis redis-cli ping
   
   # View Redis logs
   docker-compose logs redis
   ```

3. **Build failures:**
   ```bash
   # Clean build
   docker-compose build --no-cache
   
   # Remove old images
   docker system prune -a
   ```

### Performance Tuning

1. **Worker scaling:**
   ```bash
   docker-compose up -d --scale print-worker=3
   ```

2. **Resource limits:**
   Edit `docker-compose.prod.yml` deploy section

3. **Redis optimization:**
   Customize `deployments/redis/redis.conf`

## Security

### Production Checklist

- [ ] Set strong Redis password
- [ ] Enable TLS certificates
- [ ] Use non-root user (already configured)
- [ ] Limit resource usage
- [ ] Enable log rotation
- [ ] Regular security updates

### Network Security

- Services communicate via internal network
- Only necessary ports exposed
- Redis not exposed externally in production

## Backup & Recovery

### Data Backup

```bash
# Backup Redis data
docker-compose exec redis redis-cli BGSAVE
docker cp print-service-redis:/data/dump.rdb ./backup/

# Backup print outputs
docker cp print-service-server:/var/lib/print-service/output ./backup/
```

### Restore

```bash
# Restore Redis data
docker cp ./backup/dump.rdb print-service-redis:/data/
docker-compose restart redis

# Restore print outputs
docker cp ./backup/output print-service-server:/var/lib/print-service/
```
