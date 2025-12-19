#!/bin/bash

echo "========================================="
echo "Redis Performance Test"
echo "========================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Clearing Redis cache to start fresh...${NC}"
docker compose exec -T redis redis-cli FLUSHDB > /dev/null
echo -e "${GREEN}✓ Cache cleared${NC}"
echo ""

echo -e "${YELLOW}Test: Measuring Response Times${NC}"
echo ""

# First request (cache miss)
echo "Request 1 (Cache MISS - loads from database):"
TIME1=$(curl -s -w "%{time_total}\n" -o /dev/null http://localhost:8080/)
echo "  Response time: ${TIME1}s"
sleep 0.5

# Second request (cache hit)
echo "Request 2 (Cache HIT - loads from Redis):"
TIME2=$(curl -s -w "%{time_total}\n" -o /dev/null http://localhost:8080/)
echo "  Response time: ${TIME2}s"
sleep 0.5

# Third request (cache hit)
echo "Request 3 (Cache HIT - loads from Redis):"
TIME3=$(curl -s -w "%{time_total}\n" -o /dev/null http://localhost:8080/)
echo "  Response time: ${TIME3}s"
sleep 0.5

# Fourth request (cache hit)
echo "Request 4 (Cache HIT - loads from Redis):"
TIME4=$(curl -s -w "%{time_total}\n" -o /dev/null http://localhost:8080/)
echo "  Response time: ${TIME4}s"
sleep 0.5

# Fifth request (cache hit)
echo "Request 5 (Cache HIT - loads from Redis):"
TIME5=$(curl -s -w "%{time_total}\n" -o /dev/null http://localhost:8080/)
echo "  Response time: ${TIME5}s"

echo ""
echo -e "${BLUE}Performance Analysis:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Calculate average of cached requests
AVG_CACHED=$(echo "scale=4; ($TIME2 + $TIME3 + $TIME4 + $TIME5) / 4" | bc)

echo "First request (DB):    ${TIME1}s"
echo "Cached requests avg:   ${AVG_CACHED}s"
echo ""

# Calculate speedup
if command -v bc &> /dev/null; then
    SPEEDUP=$(echo "scale=2; $TIME1 / $AVG_CACHED" | bc)
    echo -e "${GREEN}Speedup with Redis cache: ${SPEEDUP}x faster${NC}"
fi

echo ""
echo -e "${YELLOW}Cache Statistics:${NC}"
docker compose exec -T redis redis-cli INFO stats | grep -E "keyspace_hits|keyspace_misses"
echo ""

echo -e "${YELLOW}Cached Data in Redis:${NC}"
echo "Keys stored:"
docker compose exec -T redis redis-cli KEYS "*" | grep -v "session:" | head -10
echo ""

echo -e "${BLUE}=========================================${NC}"
echo -e "${GREEN}Performance Test Complete!${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""
echo "Key Findings:"
echo "- First request slower (database query)"
echo "- Subsequent requests faster (Redis cache)"
echo "- Redis provides significant performance improvement"
echo "- Cache reduces database load"
echo ""

