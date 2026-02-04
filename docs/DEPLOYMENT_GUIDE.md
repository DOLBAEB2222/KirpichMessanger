# Deployment Guide for KirpichMessenger with New Features

This guide covers deploying KirpichMessenger with the new unique features optimized for low-end hardware (i3-2120, 4GB RAM).

## Prerequisites

- Ubuntu 24.04 LTS (or similar)
- Docker & Docker Compose
- 4GB RAM minimum
- 20GB SSD minimum
- Domain name (for production)

## Quick Deployment

### 1. Clone Repository
```bash
git clone https://github.com/your-repo/messenger.git
cd messenger
```

### 2. Configure Environment
```bash
cp deploy/.env.example deploy/.env
nano deploy/.env
```

**Critical Settings:**
```bash
# Generate secure secrets
JWT_SECRET=$(openssl rand -hex 32)
POSTGRES_PASSWORD=$(openssl rand -hex 32)
TURN_PASSWORD=$(openssl rand -hex 32)

# Set your domain
DOMAIN=your-domain.com
```

### 3. Database Setup

**Option A: Fresh Install (Recommended)**
```bash
cd deploy
docker compose up -d postgres redis

# Wait for PostgreSQL to be ready
sleep 10

# Initialize database
docker exec -i messenger-postgres psql -U messenger -d messenger < ../database/schema.sql
```

**Option B: Migration from v1**
```bash
cd deploy
docker compose up -d postgres redis

# Wait for PostgreSQL to be ready
sleep 10

# Run migration
docker exec -i messenger-postgres psql -U messenger -d messenger < ../database/migration_v2.sql
```

### 4. Start All Services
```bash
docker compose up -d
```

### 5. Verify Deployment
```bash
# Check service status
docker compose ps

# Check logs
docker compose logs -f api

# Test API
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "ok",
  "timestamp": 1705320000,
  "service": "messenger-api"
}
```

## Resource Optimization Details

### Memory Allocation (Total: ~2.15GB)

| Service | Memory | CPU | Purpose |
|---------|--------|-----|---------|
| PostgreSQL | 700MB | 1.0 core | Database |
| Redis | 250MB | 0.5 core | Cache |
| API | 900MB | 1.5 cores | Application |
| Caddy | 150MB | 0.5 core | Proxy |
| coturn | 150MB | 0.5 core | WebRTC |

### PostgreSQL Optimization

Configuration in `config/postgres.conf`:
- Max connections: 80
- Shared buffers: 128MB
- Effective cache size: 512MB
- Work memory: 4MB per operation
- WAL size: 128-512MB
- Worker processes: 2

### Redis Optimization

Configuration in `deploy/docker-compose.yml`:
- Max memory: 200MB
- Eviction policy: allkeys-lru
- Save intervals: Optimized for low I/O
- Persistence: AOF with everysec

### Application Optimization

Environment variables in `deploy/docker-compose.yml`:
- `GOMAXPROCS=2` - Limit to 2 CPU cores
- `DB_MAX_OPEN_CONNS=20` - Connection pool size
- `DB_MAX_IDLE_CONNS=10` - Idle connections
- `DB_CONN_MAX_LIFETIME=300` - Connection reuse (5 min)

## Testing New Features

### Wiki Pages
```bash
# Create a wiki page
curl -X POST http://localhost:8080/api/v1/wiki \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "channel_id": "uuid",
    "slug": "getting-started",
    "title": "Getting Started",
    "content": "# Welcome\n\nThis is our wiki."
  }'

# Get wiki tree
curl http://localhost:8080/api/v1/wiki/uuid/tree \
  -H "Authorization: Bearer $TOKEN"
```

### Code Snippets
```bash
# Create code snippet
curl -X POST http://localhost:8080/api/v1/code \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "message_id": "uuid",
    "chat_id": "uuid",
    "language": "python",
    "code": "print(\"Hello\")"
  }'
```

### Temporary Roles
```bash
# Grant moderator role
curl -X POST http://localhost:8080/api/v1/temp-roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_id": "uuid",
    "target_type": "channel",
    "user_id": "uuid",
    "role_type": "moderator",
    "permissions": ["edit_messages"],
    "duration_hours": 24
  }'
```

### RSS Feeds
```bash
# Add RSS feed
curl -X POST http://localhost:8080/api/v1/rss \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "channel_id": "uuid",
    "url": "https://example.com/feed.xml"
  }'
```

## Monitoring and Maintenance

### Check Resource Usage
```bash
# Docker stats
docker stats

# Service-specific stats
docker stats messenger-postgres
docker stats messenger-redis
docker stats messenger-api
```

### Database Maintenance
```bash
# Connect to database
docker exec -it messenger-postgres psql -U messenger -d messenger

# Run vacuum (monthly)
docker exec messenger-postgres psql -U messenger -d messenger -c "VACUUM ANALYZE;"

# Check database size
docker exec messenger-postgres psql -U messenger -d messenger -c "SELECT pg_size_pretty(pg_database_size('messenger'));"
```

### Redis Maintenance
```bash
# Check memory usage
docker exec messenger-redis redis-cli INFO memory

# Check cache hit rate
docker exec messenger-redis redis-cli INFO stats | grep keyspace

# Flush cache (use with caution)
docker exec messenger-redis redis-cli FLUSHDB
```

### Logs
```bash
# All logs
docker compose logs -f

# API logs
docker compose logs -f api

# PostgreSQL logs
docker compose logs -f postgres

# Last 100 lines
docker compose logs --tail=100
```

## Troubleshooting

### High Memory Usage
```bash
# Check which service uses most memory
docker stats --no-stream

# Restart specific service
docker compose restart api

# Clear Redis cache
docker exec messenger-redis redis-cli FLUSHDB
```

### Database Connection Errors
```bash
# Check PostgreSQL status
docker compose ps postgres

# Check PostgreSQL logs
docker compose logs postgres

# Restart PostgreSQL
docker compose restart postgres
```

### Slow API Response
```bash
# Check API logs
docker compose logs --tail=50 api

# Restart API
docker compose restart api

# Clear Redis cache
docker exec messenger-redis redis-cli FLUSHDB
```

### RSS Feed Not Updating
```bash
# Manually refresh feed
curl -X POST http://localhost:8080/api/v1/rss/:id/refresh \
  -H "Authorization: Bearer $TOKEN"

# Check feed status
curl http://localhost:8080/api/v1/rss/:id \
  -H "Authorization: Bearer $TOKEN"
```

## Scaling for Higher Load

If you need to support more than 500 users, consider:

1. **Increase Resources**:
   - RAM: Upgrade to 8GB+
   - CPU: 4+ cores

2. **Horizontal Scaling**:
   - Run multiple API instances behind load balancer
   - Use PostgreSQL read replicas
   - Redis cluster mode

3. **Optimization**:
   - Enable CDN for media files
   - Use connection pooling
   - Implement message queue for RSS feeds

## Security Checklist

- [ ] Change default passwords in `.env`
- [ ] Set strong `JWT_SECRET` (32+ characters)
- [ ] Configure firewall rules
- [ ] Enable SSL/TLS (Caddy handles automatically)
- [ ] Set up database backups
- [ ] Monitor logs for suspicious activity
- [ ] Regular security updates

## Backup and Restore

### Backup Database
```bash
# Create backup
docker exec messenger-postgres pg_dump -U messenger messenger > backup_$(date +%Y%m%d).sql

# Backup volume
docker run --rm -v messenger_postgres_data:/data -v $(pwd):/backup \
  alpine tar czf /backup/postgres_backup_$(date +%Y%m%d).tar.gz /data
```

### Restore Database
```bash
# Restore from SQL dump
docker exec -i messenger-postgres psql -U messenger messenger < backup_20240115.sql

# Restore from volume backup
docker run --rm -v messenger_postgres_data:/data -v $(pwd):/backup \
  alpine tar xzf /backup/postgres_backup_20240115.tar.gz -C /
```

## Production Deployment Tips

1. **Use Production Environment**:
   ```bash
   export APP_ENV=production
   ```

2. **Set Up Monitoring**:
   - Prometheus + Grafana for metrics
   - Alertmanager for alerts
   - Log aggregation (ELK stack)

3. **Load Balancing**:
   - Use Caddy for HTTP/HTTPS
   - Consider HAProxy for TCP (WebRTC)

4. **Database Backups**:
   - Automated daily backups
   - Off-site storage
   - Test restore procedure

5. **SSL Certificates**:
   - Caddy handles Let's Encrypt automatically
   - Ensure domain DNS is configured correctly

## Support and Resources

- API Documentation: See `README.md`
- Feature Documentation: See `docs/FEATURES.md`
- Technical Details: See `docs/TDD.md`
- Issues: Report on GitHub Issues
