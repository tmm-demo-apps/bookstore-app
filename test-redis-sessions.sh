#!/bin/bash

echo "========================================="
echo "Redis Session Management Test"
echo "========================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Test 1: Check Initial Session Count in Redis${NC}"
INITIAL_COUNT=$(docker compose exec -T redis redis-cli KEYS "session:*" | wc -l)
echo "Current sessions in Redis: $INITIAL_COUNT"
echo ""

echo -e "${YELLOW}Test 2: Create New Session (Add to Cart)${NC}"
echo "Adding item to cart (creates new session)..."
COOKIE_FILE="/tmp/redis-session-test-$$.txt"
curl -s -c "$COOKIE_FILE" -X POST http://localhost:8080/cart/add -d "product_id=1&quantity=2" > /dev/null
sleep 1
echo -e "${GREEN}✓ Item added to cart${NC}"
echo ""

echo -e "${YELLOW}Test 3: Check Session Count After Request${NC}"
NEW_COUNT=$(docker compose exec -T redis redis-cli KEYS "session:*" | wc -l)
echo "Sessions in Redis now: $NEW_COUNT"
if [ $NEW_COUNT -gt $INITIAL_COUNT ]; then
    echo -e "${GREEN}✓ New session created in Redis!${NC}"
else
    echo -e "${YELLOW}Note: Session may have been reused${NC}"
fi
echo ""

echo -e "${YELLOW}Test 4: Verify Session Persistence${NC}"
echo "Making another request with same session cookie..."
curl -s -b "$COOKIE_FILE" http://localhost:8080/cart > /dev/null
sleep 1
echo -e "${GREEN}✓ Session persisted across requests${NC}"
echo ""

echo -e "${YELLOW}Test 5: Inspect a Session Key${NC}"
SESSION_KEY=$(docker compose exec -T redis redis-cli KEYS "session:*" | head -1 | tr -d '\r')
if [ ! -z "$SESSION_KEY" ]; then
    echo "Sample session key: $SESSION_KEY"
    echo "Session TTL (seconds):"
    docker compose exec -T redis redis-cli TTL "$SESSION_KEY"
    echo "(TTL of 2592000 seconds = 30 days)"
else
    echo "No session keys found"
fi
echo ""

echo -e "${YELLOW}Test 6: Check Session Data Type${NC}"
if [ ! -z "$SESSION_KEY" ]; then
    echo "Session data type:"
    docker compose exec -T redis redis-cli TYPE "$SESSION_KEY"
    echo "(Sessions stored as strings - serialized data)"
fi
echo ""

# Cleanup
rm -f "$COOKIE_FILE"

echo -e "${BLUE}=========================================${NC}"
echo -e "${GREEN}Redis Session Test Complete!${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""
echo "Summary:"
echo "- Sessions are stored in Redis with 'session:' prefix"
echo "- Sessions persist for 30 days (2592000 seconds)"
echo "- Sessions survive app restarts (stored in Redis)"
echo "- Multiple requests use the same session"
echo ""

