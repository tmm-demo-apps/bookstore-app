#!/bin/bash

echo "========================================="
echo "Redis Caching Test"
echo "========================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Test 1: First Request (Should be Cache MISS)${NC}"
echo "Loading homepage..."
curl -s http://localhost:8080/ > /dev/null
sleep 1
echo -e "${GREEN}✓ First request completed${NC}"
echo ""

echo -e "${YELLOW}Test 2: Second Request (Should be Cache HIT)${NC}"
echo "Loading homepage again..."
curl -s http://localhost:8080/ > /dev/null
sleep 1
echo -e "${GREEN}✓ Second request completed${NC}"
echo ""

echo -e "${YELLOW}Test 3: Product Detail Page${NC}"
echo "Loading product detail (ID: 1)..."
curl -s http://localhost:8080/products/1 > /dev/null
sleep 1
echo "Loading same product again..."
curl -s http://localhost:8080/products/1 > /dev/null
sleep 1
echo -e "${GREEN}✓ Product detail requests completed${NC}"
echo ""

echo -e "${YELLOW}Test 4: Check Redis Keys${NC}"
echo "Keys stored in Redis:"
docker compose exec -T redis redis-cli KEYS "*" | head -10
echo ""

echo -e "${YELLOW}Test 5: Check Cache Statistics${NC}"
echo "Recent cache activity from logs:"
docker compose logs app 2>&1 | grep -i "cache" | tail -15
echo ""

echo -e "${BLUE}=========================================${NC}"
echo -e "${GREEN}Redis Cache Test Complete!${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""
echo "Summary:"
echo "- First requests should show 'Cache MISS'"
echo "- Subsequent requests should show 'Cache HIT'"
echo "- Redis should contain keys like 'products:all', 'product:1', 'categories:all'"
echo ""

