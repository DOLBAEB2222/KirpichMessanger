#!/bin/bash

# Media Cleanup Script
# Removes orphaned media files and old deleted message media

set -e

MEDIA_DIR="${MEDIA_PATH:-/app/media}"
RETENTION_DAYS="${MEDIA_RETENTION_DAYS:-30}"
LOG_FILE="/var/log/messenger/cleanup.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log "Starting media cleanup job"
log "Media directory: $MEDIA_DIR"
log "Retention days: $RETENTION_DAYS"

# Check if database connection is available
if ! docker exec messenger-postgres pg_isready -U messenger > /dev/null 2>&1; then
    log "ERROR: Database is not available"
    exit 1
fi

# Get list of media files referenced in database
log "Fetching active media files from database..."
ACTIVE_FILES=$(docker exec messenger-postgres psql -U messenger -d messenger -t -c \
    "SELECT file_path FROM media_files WHERE created_at > NOW() - INTERVAL '${RETENTION_DAYS} days';" | \
    tr -d ' ' | grep -v '^$')

# Count active files
ACTIVE_COUNT=$(echo "$ACTIVE_FILES" | wc -l)
log "Active media files: $ACTIVE_COUNT"

# Find orphaned files
log "Scanning for orphaned files..."
DELETED_COUNT=0
FREED_SPACE=0

if [ -d "$MEDIA_DIR" ]; then
    for file in $(find "$MEDIA_DIR" -type f -mtime +${RETENTION_DAYS}); do
        # Check if file is in active list
        if ! echo "$ACTIVE_FILES" | grep -q "$(basename $file)"; then
            # Get file size before deletion
            FILE_SIZE=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo 0)
            
            # Delete file
            rm -f "$file"
            
            DELETED_COUNT=$((DELETED_COUNT + 1))
            FREED_SPACE=$((FREED_SPACE + FILE_SIZE))
            
            log "Deleted orphaned file: $(basename $file) ($(numfmt --to=iec-i --suffix=B $FILE_SIZE 2>/dev/null || echo ${FILE_SIZE}B))"
        fi
    done
else
    log "WARNING: Media directory does not exist: $MEDIA_DIR"
fi

# Convert bytes to human readable
if command -v numfmt > /dev/null 2>&1; then
    FREED_SPACE_HR=$(numfmt --to=iec-i --suffix=B $FREED_SPACE)
else
    FREED_SPACE_HR="${FREED_SPACE} bytes"
fi

log "Cleanup completed"
log "Files deleted: $DELETED_COUNT"
log "Space freed: $FREED_SPACE_HR"

# Clean up empty directories
log "Removing empty directories..."
find "$MEDIA_DIR" -type d -empty -delete 2>/dev/null || true

# Optional: Clean up old database records
log "Cleaning up old deleted message records..."
DELETED_MESSAGES=$(docker exec messenger-postgres psql -U messenger -d messenger -t -c \
    "DELETE FROM messages WHERE is_deleted = true AND updated_at < NOW() - INTERVAL '${RETENTION_DAYS} days'; SELECT ROW_COUNT();")

log "Deleted message records removed: $(echo $DELETED_MESSAGES | tr -d ' ')"

# Vacuum database
log "Running database vacuum..."
docker exec messenger-postgres psql -U messenger -d messenger -c "VACUUM ANALYZE messages, media_files;" > /dev/null 2>&1

log "Database vacuum completed"
log "Cleanup job finished successfully"

# Send notification if cleanup freed significant space (> 1GB)
if [ $FREED_SPACE -gt 1073741824 ]; then
    log "NOTICE: Large cleanup performed - freed $FREED_SPACE_HR"
    # Add email notification here if configured
fi

exit 0
