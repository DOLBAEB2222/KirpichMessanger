#!/bin/bash

# Media cleanup script
# Removes unused media files older than 30 days

set -e

UPLOADS_DIR="${UPLOADS_DIR:-./uploads}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
DRY_RUN="${DRY_RUN:-false}"

if [ ! -d "$UPLOADS_DIR" ]; then
    echo "Uploads directory not found: $UPLOADS_DIR"
    exit 0
fi

echo "Starting media cleanup..."
echo "Uploads directory: $UPLOADS_DIR"
echo "Retention period: $RETENTION_DAYS days"
echo "Dry run: $DRY_RUN"
echo ""

# Find files older than retention period
find_cmd="find \"$UPLOADS_DIR\" -type f -mtime +$RETENTION_DAYS"

if [ "$DRY_RUN" = "true" ]; then
    echo "DRY RUN - Files that would be deleted:"
    eval "$find_cmd" 2>/dev/null | while read -r file; do
        size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo "0")
        echo "  $file (${size} bytes)"
    done
else
    # Count files before deletion
    file_count=$(eval "$find_cmd" 2>/dev/null | wc -l)
    
    if [ "$file_count" -eq 0 ]; then
        echo "No files to clean up."
    else
        # Calculate total size
        total_size=$(eval "$find_cmd -exec stat -f%z {} + 2>/dev/null | awk '{sum+=$1} END {print sum}' || eval "$find_cmd -exec stat -c%s {} + 2>/dev/null | awk '{sum+=$1} END {print sum}'" || echo "0")
        
        # Delete files
        eval "$find_cmd -delete" 2>/dev/null || true
        
        echo "Deleted $file_count files"
        echo "Freed approximately $total_size bytes"
    fi
fi

# Remove empty directories
echo ""
echo "Removing empty directories..."
if [ "$DRY_RUN" = "true" ]; then
    find "$UPLOADS_DIR" -type d -empty 2>/dev/null | while read -r dir; do
        echo "  Would remove: $dir"
    done
else
    find "$UPLOADS_DIR" -type d -empty -delete 2>/dev/null || true
fi

echo ""
echo "Cleanup complete!"

# Show disk usage
echo ""
echo "Current disk usage:"
du -sh "$UPLOADS_DIR" 2>/dev/null || echo "Unable to calculate disk usage"
