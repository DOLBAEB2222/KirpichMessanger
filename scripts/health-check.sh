#!/bin/bash

# Health check script for KirpichMessenger
# Checks all services and reports their status

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "KirpichMessenger Health Check"
echo "=========================================="
echo ""

# Function to check service status
check_service() {
    local service=$1
    local url=$2
    local name=$3

    echo -n "Checking $name... "

    if docker compose -f deploy/docker-compose.yml ps $service | grep -q "Up"; then
        if [ -n "$url" ]; then
            if curl -s -f "$url" > /dev/null 2>&1; then
                echo -e "${GREEN}✓ Healthy${NC}"
                return 0
            else
                echo -e "${YELLOW}⚠ Running but unhealthy${NC}"
                return 1
            fi
        else
            echo -e "${GREEN}✓ Running${NC}"
            return 0
        fi
    else
        echo -e "${RED}✗ Down${NC}"
        return 2
    fi
}

# Check services
cd "$(dirname "$0")/.."

echo "Services:"
echo "---------"
check_service "postgres" "" "PostgreSQL"
check_service "redis" "" "Redis"
check_service "api" "http://localhost:8080/health" "API"
check_service "caddy" "http://localhost:2019/metrics" "Caddy"
check_service "coturn" "" "Coturn"
echo ""

# Check database
echo "Database:"
echo "---------"
if docker compose -f deploy/docker-compose.yml ps postgres | grep -q "Up"; then
    echo -n "Checking database connection... "
    if docker exec messenger-postgres pg_isready -U messenger -d messenger > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Connected${NC}"
    else
        echo -e "${RED}✗ Failed${NC}"
    fi

    echo -n "Checking database size... "
    DB_SIZE=$(docker exec messenger-postgres psql -U messenger -d messenger -t -c "SELECT pg_size_pretty(pg_database_size('messenger'));" 2>/dev/null | xargs)
    if [ -n "$DB_SIZE" ]; then
        echo -e "${GREEN}$DB_SIZE${NC}"
    else
        echo -e "${YELLOW}Unknown${NC}"
    fi

    echo -n "Checking tables... "
    TABLES=$(docker exec messenger-postgres psql -U messenger -d messenger -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null | xargs)
    if [ -n "$TABLES" ]; then
        echo -e "${GREEN}$TABLES tables${NC}"
    else
        echo -e "${YELLOW}Unknown${NC}"
    fi

    # Check for v2 tables
    echo -n "Checking v2 features... "
    V2_TABLES=$(docker exec messenger-postgres psql -U messenger -d messenger -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('wiki_pages', 'code_snippets', 'temp_roles', 'rss_feeds');" 2>/dev/null | xargs)
    if [ "$V2_TABLES" = "4" ]; then
        echo -e "${GREEN}✓ All v2 features installed${NC}"
    elif [ -n "$V2_TABLES" ]; then
        echo -e "${YELLOW}$V2_TABLES/4 features installed${NC}"
    else
        echo -e "${RED}✗ Not installed${NC}"
    fi
else
    echo -e "${RED}PostgreSQL is not running${NC}"
fi
echo ""

# Check Redis
echo "Redis:"
echo "------"
if docker compose -f deploy/docker-compose.yml ps redis | grep -q "Up"; then
    echo -n "Checking Redis connection... "
    if docker exec messenger-redis redis-cli ping > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Connected${NC}"
    else
        echo -e "${RED}✗ Failed${NC}"
    fi

    echo -n "Checking Redis memory... "
    REDIS_MEM=$(docker exec messenger-redis redis-cli INFO memory | grep used_memory_human | cut -d: -f2 | tr -d '\r')
    if [ -n "$REDIS_MEM" ]; then
        echo -e "${GREEN}$REDIS_MEM${NC}"
    else
        echo -e "${YELLOW}Unknown${NC}"
    fi

    echo -n "Checking cache keys... "
    KEYS=$(docker exec messenger-redis redis-cli DBSIZE 2>/dev/null | tr -d '\r')
    if [ -n "$KEYS" ]; then
        echo -e "${GREEN}$KEYS keys${NC}"
    else
        echo -e "${YELLOW}Unknown${NC}"
    fi
else
    echo -e "${RED}Redis is not running${NC}"
fi
echo ""

# Check resources
echo "Resource Usage:"
echo "---------------"
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}" 2>/dev/null || echo -e "${YELLOW}Docker stats not available${NC}"
echo ""

# Check disk space
echo "Disk Space:"
echo "-----------"
df -h /var/lib/docker 2>/dev/null | tail -1 | awk '{print "Used: " $3 " / " $2 " (" $5 ")"}' || echo -e "${YELLOW}Disk info not available${NC}"
echo ""

echo "=========================================="
echo "Health check complete!"
echo "=========================================="
