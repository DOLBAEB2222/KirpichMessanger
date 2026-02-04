# Upgrade Guide v1 → v2

This guide explains how to upgrade an existing KirpichMessenger installation to version 2 with new features.

## New Features in v2

1. **Wiki in Channels** - Documentation pages for channels
2. **Code Snippets** - Syntax-highlighted code sharing
3. **Temporary Roles** - Time-limited role assignments
4. **RSS Aggregator** - Subscribe to RSS feeds in channels

## Prerequisites

- Backup your current installation
- Ensure you have at least 500MB free space
- Stop the application before upgrading

## Upgrade Steps

### 1. Backup Current Data

```bash
# Backup database
cd deploy
docker exec messenger-postgres pg_dump -U messenger messenger > backup_before_upgrade.sql

# Backup volumes
docker run --rm -v messenger_postgres_data:/data -v $(pwd):/backup \
  alpine tar czf /backup/postgres_data_before_upgrade.tar.gz /data

docker run --rm -v messenger_redis_data:/data -v $(pwd):/backup \
  alpine tar czf /backup/redis_data_before_upgrade.tar.gz /data
```

### 2. Pull Latest Code

```bash
cd /path/to/messenger
git fetch origin
git checkout main
git pull origin main
```

### 3. Update Environment Variables

Open `.env` and check if new variables need to be set:

```bash
nano deploy/.env
```

Optional additions for better resource control:
```bash
GOMAXPROCS=2
DB_MAX_OPEN_CONNS=20
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=300
REDIS_MAX_POOL_SIZE=10
```

### 4. Stop Services

```bash
cd deploy
docker compose down
```

### 5. Run Database Migration

```bash
# Start only PostgreSQL
docker compose up -d postgres

# Wait for PostgreSQL to be ready
sleep 15

# Run migration
docker exec -i messenger-postgres psql -U messenger -d messenger < ../database/migration_v2.sql

# Verify migration
docker exec messenger-postgres psql -U messenger -d messenger -c "\dt"
```

Expected new tables:
- `wiki_pages`
- `code_snippets`
- `temp_roles`
- `rss_feeds`
- `rss_items`

### 6. Rebuild and Start Services

```bash
# Build new images
docker compose build

# Start all services
docker compose up -d
```

### 7. Verify Installation

```bash
# Check all services are running
docker compose ps

# Check API health
curl http://localhost:8080/health

# Check logs for errors
docker compose logs -f api
```

### 8. Test New Features

```bash
# Get authentication token
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"phone":"+1234567890","password":"yourpassword"}' | jq -r '.token')

# Test wiki endpoint
curl http://localhost:8080/api/v1/wiki/123e4567-e89b-12d3-a456-426614174000/tree \
  -H "Authorization: Bearer $TOKEN"

# Test code endpoint
curl http://localhost:8080/api/v1/code/chat/123e4567-e89b-12d3-a456-426614174001 \
  -H "Authorization: Bearer $TOKEN"

# Test temp roles endpoint
curl http://localhost:8080/api/v1/temp-roles/user/123e4567-e89b-12d3-a456-426614174002 \
  -H "Authorization: Bearer $TOKEN"

# Test RSS endpoint
curl http://localhost:8080/api/v1/rss \
  -H "Authorization: Bearer $TOKEN"
```

## Rollback Procedure

If something goes wrong, you can rollback:

```bash
cd deploy

# Stop services
docker compose down

# Restore database
docker exec -i messenger-postgres psql -U messenger -d messenger < backup_before_upgrade.sql

# Restore volumes
docker run --rm -v messenger_postgres_data:/data -v $(pwd):/backup \
  alpine tar xzf /backup/postgres_data_before_upgrade.tar.gz -C /

docker run --rm -v messenger_redis_data:/data -v $(pwd):/backup \
  alpine tar xzf /backup/redis_data_before_upgrade.tar.gz -C /

# Revert code
git checkout previous-version-tag

# Restart services
docker compose up -d
```

## Known Issues and Solutions

### Issue: Migration Fails with "type code_language does not exist"

**Solution**: Run the full migration script instead of individual statements:
```bash
docker exec -i messenger-postgres psql -U messenger -d messenger < ../database/migration_v2.sql
```

### Issue: "memory limit reached" after upgrade

**Solution**: The new resource limits are tighter. Adjust in `docker-compose.yml`:
```yaml
deploy:
  resources:
    limits:
      memory: 1G  # Increase from 900M
```

### Issue: Some endpoints return 404

**Solution**: Verify routes are registered in `main.go`. Check that new handlers are imported.

### Issue: RSS feed returns parsing error

**Solution**: Some RSS feeds use non-standard formats. Check the feed URL manually and verify it's valid RSS 2.0 or Atom.

## Performance After Upgrade

After the upgrade, monitor performance:

```bash
# Check resource usage
docker stats

# Check database performance
docker exec messenger-postgres psql -U messenger -d messenger -c "
  SELECT schemaname, tablename, n_tup_ins, n_tup_upd, n_tup_del
  FROM pg_stat_user_tables
  ORDER BY n_tup_ins + n_tup_upd + n_tup_upd DESC;
"

# Check Redis cache hit rate
docker exec messenger-redis redis-cli INFO stats | grep keyspace
```

## Next Steps

1. **Explore New Features**:
   - Read `docs/FEATURES.md` for detailed documentation
   - Test each feature in your development environment
   - Create wiki pages for your channels
   - Add RSS feeds to relevant channels

2. **Configure RSS Feeds**:
   - Identify useful RSS feeds for your channels
   - Set up periodic refresh (if needed)
   - Monitor feed fetching errors

3. **Set Up Temporary Roles**:
   - Define role templates for your use cases
   - Train moderators on using the system
   - Monitor role expirations

4. **Enable Code Snippets**:
   - Update client to support syntax highlighting
   - Add code snippet button to message composer
   - Test with different languages

## Support

If you encounter issues:
1. Check `docs/DEPLOYMENT_GUIDE.md` for troubleshooting
2. Review logs: `docker compose logs -f api`
3. Open an issue on GitHub with:
   - Error messages
   - Logs output
   - Steps to reproduce
   - System configuration

## Version Compatibility

| Component | v1 | v2 | Notes |
|-----------|----|----|------|
| PostgreSQL | 15+ | 16 | Recommend upgrade to 16 |
| Redis | 6+ | 7 | Recommend upgrade to 7 |
| Go | 1.21+ | 1.21+ | No change |
| Docker Compose | v2 | v2.20+ | Minimum version |

## Security Notes

After upgrade:
1. Review `.env` file for any new required secrets
2. Update firewall rules if needed (no new ports required)
3. Check that SSL certificates are still valid
4. Review new API endpoints in authentication middleware

## Frequently Asked Questions

**Q: Will I lose my existing data?**
A: No, the migration only adds new tables. Existing data is preserved.

**Q: Can I skip the migration?**
A: No, the new features require new database tables. Migration is required.

**Q: How long does the upgrade take?**
A: Typically 5-10 minutes, depending on data size.

**Q: Can I upgrade without downtime?**
A: For minimal downtime, you can:
   1. Start PostgreSQL in a new container
   2. Run migration
   3. Stop old API
   4. Start new API
   This requires database replication setup.

**Q: What if the migration fails?**
A: Restore from backup and try again. Check error logs for specific issues.
