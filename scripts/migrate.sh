#!/bin/bash

# Migration script for KirpichMessenger
# Runs database migrations and verifies success

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "=========================================="
echo "KirpichMessenger Migration Script"
echo "=========================================="
echo ""

cd "$(dirname "$0")/.."

# Check if deploy/.env exists
if [ ! -f "deploy/.env" ]; then
    echo -e "${RED}Error: deploy/.env file not found${NC}"
    echo "Please copy deploy/.env.example to deploy/.env and configure it"
    exit 1
fi

# Check if services are running
echo "Checking if services are running..."
if ! docker compose -f deploy/docker-compose.yml ps postgres | grep -q "Up"; then
    echo -e "${YELLOW}PostgreSQL is not running. Starting...${NC}"
    docker compose -f deploy/docker-compose.yml up -d postgres
    echo "Waiting for PostgreSQL to be ready..."
    sleep 15
fi

echo ""

# Check database connection
echo "Testing database connection..."
if ! docker exec messenger-postgres pg_isready -U messenger -d messenger > /dev/null 2>&1; then
    echo -e "${RED}Error: Cannot connect to database${NC}"
    echo "Please check your configuration in deploy/.env"
    exit 1
fi
echo -e "${GREEN}✓ Database connected${NC}"
echo ""

# Check current version
echo "Checking current database version..."
CURRENT_VERSION=$(docker exec messenger-postgres psql -U messenger -d messenger -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'users';" 2>/dev/null | xargs)

if [ "$CURRENT_VERSION" = "0" ]; then
    echo "Database is empty. Initializing with full schema..."
    docker exec -i messenger-postgres psql -U messenger -d messenger < database/schema.sql
    echo -e "${GREEN}✓ Database initialized (v2)${NC}"
else
    # Check if v2 tables exist
    V2_TABLES=$(docker exec messenger-postgres psql -U messenger -d messenger -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('wiki_pages', 'code_snippets', 'temp_roles', 'rss_feeds');" 2>/dev/null | xargs)

    if [ "$V2_TABLES" = "4" ]; then
        echo -e "${GREEN}✓ Database is already at v2${NC}"
        echo "No migration needed."
    else
        echo "Database needs migration to v2..."
        echo "Running migration..."

        # Backup before migration
        BACKUP_FILE="backup_$(date +%Y%m%d_%H%M%S).sql"
        echo "Creating backup: $BACKUP_FILE"
        docker exec messenger-postgres pg_dump -U messenger messenger > "backup/$BACKUP_FILE"

        # Run migration
        docker exec -i messenger-postgres psql -U messenger -d messenger < database/migration_v2.sql

        echo -e "${GREEN}✓ Migration to v2 completed successfully${NC}"
        echo "Backup saved to: backup/$BACKUP_FILE"
    fi
fi
echo ""

# Verify tables
echo "Verifying database tables..."
TABLE_COUNT=$(docker exec messenger-postgres psql -U messenger -d messenger -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null | xargs)
echo -e "${GREEN}✓ $TABLE_COUNT tables found${NC}"

# Check for v2 specific tables
echo ""
echo "Checking v2 features:"
V2_FEATURES=("wiki_pages" "code_snippets" "temp_roles" "rss_feeds")
ALL_PRESENT=true

for feature in "${V2_FEATURES[@]}"; do
    EXISTS=$(docker exec messenger-postgres psql -U messenger -d messenger -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_name = '$feature';" 2>/dev/null | xargs)
    if [ "$EXISTS" = "1" ]; then
        echo -e "  ${GREEN}✓ $feature${NC}"
    else
        echo -e "  ${RED}✗ $feature${NC}"
        ALL_PRESENT=false
    fi
done

if [ "$ALL_PRESENT" = true ]; then
    echo ""
    echo -e "${GREEN}=========================================="
    echo "All v2 features are installed!"
    echo "==========================================${NC}"
else
    echo ""
    echo -e "${RED}=========================================="
    echo "Some v2 features are missing!"
    echo "==========================================${NC}"
    exit 1
fi

echo ""
echo "Next steps:"
echo "  1. Start all services: docker compose -f deploy/docker-compose.yml up -d"
echo "  2. Run health check: ./scripts/health-check.sh"
echo "  3. Test the API: curl http://localhost:8080/health"
echo ""
