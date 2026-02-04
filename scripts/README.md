# Scripts

This directory contains utility scripts for managing KirpichMessenger.

## Available Scripts

### migrate.sh
Database migration script for upgrading from v1 to v2 or initializing a fresh installation.

**Usage:**
```bash
./scripts/migrate.sh
```

**What it does:**
- Checks if PostgreSQL is running
- Creates backup before migration
- Runs database migration or full initialization
- Verifies all v2 features are installed
- Provides detailed output at each step

**Requirements:**
- Docker and Docker Compose installed
- PostgreSQL container running
- `.env` file configured in `deploy/`

### health-check.sh
System health monitoring script that checks all services and reports their status.

**Usage:**
```bash
./scripts/health-check.sh
```

**What it checks:**
- Service status (PostgreSQL, Redis, API, Caddy, Coturn)
- Database connection and size
- Number of database tables
- v2 feature installation status
- Redis connection and memory usage
- Cache key count
- Resource usage (CPU and memory)
- Disk space

**Requirements:**
- Docker and Docker Compose installed
- All services running

### cleanup.sh
Cleanup script for removing old data and temporary files.

**Usage:**
```bash
./scripts/cleanup.sh
```

**What it does:**
- Removes old log files
- Cleans temporary files
- Clears expired cache entries
- Removes old backups (configurable)

**Requirements:**
- Docker and Docker Compose installed

## Making Scripts Executable

On Linux/Mac:
```bash
chmod +x scripts/*.sh
```

On Windows (Git Bash):
```bash
# Scripts should be executable automatically
# If not, use Git Bash:
chmod +x scripts/*.sh
```

## Running Scripts

From project root:
```bash
# Migration
./scripts/migrate.sh

# Health check
./scripts/health-check.sh

# Cleanup
./scripts/cleanup.sh
```

From scripts directory:
```bash
cd scripts

# Migration
./migrate.sh

# Health check
./health-check.sh

# Cleanup
./cleanup.sh
```

## Output Interpretation

### migrate.sh
- ✓ Green checkmark: Success
- ⚠ Yellow warning: Non-critical issue
- ✗ Red cross: Error

### health-check.sh
- ✓ Healthy: Service running normally
- ⚠ Warning: Service running but with issues
- ✗ Down: Service not running

## Troubleshooting

### Script Permission Denied
```bash
# Fix permission
chmod +x scripts/*.sh

# Or run with bash
bash scripts/migrate.sh
```

### PostgreSQL Not Running
```bash
# Start PostgreSQL
cd deploy
docker compose up -d postgres

# Wait 15 seconds, then run script again
```

### Docker Not Available
```bash
# Install Docker
# Ubuntu/Debian:
curl -fsSL https://get.docker.com | sh

# Start Docker
sudo systemctl start docker
sudo systemctl enable docker
```

## Automated Execution

### Cron Jobs

Setup automated health checks:
```bash
# Edit crontab
crontab -e

# Add health check every 15 minutes
*/15 * * * * /path/to/messenger/scripts/health-check.sh >> /var/log/messenger-health.log 2>&1
```

### Systemd Service

Create a systemd service for health monitoring:
```bash
# Create service file
sudo nano /etc/systemd/system/messenger-health.service
```

Content:
```ini
[Unit]
Description=KirpichMessenger Health Check
After=docker.service

[Service]
Type=oneshot
User=your-user
WorkingDirectory=/path/to/messenger
ExecStart=/path/to/messenger/scripts/health-check.sh

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable messenger-health.timer
sudo systemctl start messenger-health.timer
```

## Contributing

When adding new scripts:
1. Make script executable (`chmod +x`)
2. Add usage documentation
3. Include error handling
4. Provide clear output
5. Add requirements section
6. Update this README
7. Test on both fresh and existing installations

## Support

For script issues:
- Check logs in `/var/log/messenger/`
- Run with `-x` flag for debugging: `bash -x scripts/migrate.sh`
- Open an issue on GitHub
