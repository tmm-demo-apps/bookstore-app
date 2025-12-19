#!/bin/bash

# Smoke Test Script for E-commerce Application
# Run this after every code change to verify core functionality

set -e  # Exit on any error

BASE_URL="http://localhost:8080"
COOKIE_JAR=$(mktemp)
TEST_EMAIL="smoke_test_$(date +%s)@example.com"
TEST_PASSWORD="TestPassword123!"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

function log_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
    TESTS_RUN=$((TESTS_RUN + 1))
}

function log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

function log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

function cleanup() {
    rm -f "$COOKIE_JAR"
}

trap cleanup EXIT

echo "========================================="
echo "E-commerce Application Smoke Tests"
echo "========================================="
echo ""

# Test 0: Go code formatting check
log_test "Checking Go code formatting..."
UNFORMATTED=$(gofmt -l . 2>/dev/null)
if [ -z "$UNFORMATTED" ]; then
    log_pass "Go code is properly formatted"
else
    log_fail "Go code is not formatted. Run: go fmt ./..."
    echo "Unformatted files:"
    echo "$UNFORMATTED"
    echo ""
    echo "To fix, run: go fmt ./..."
    exit 1
fi

# Test 1: Server is running
log_test "Checking if server is running..."
if curl -s "$BASE_URL" > /dev/null; then
    log_pass "Server is running"
else
    log_fail "Server is not responding"
    exit 1
fi

# Test 2: Products page loads
log_test "Loading products page..."
RESPONSE=$(curl -s "$BASE_URL")
if echo "$RESPONSE" | grep -q "Products"; then
    log_pass "Products page loaded"
else
    log_fail "Products page did not load correctly"
fi

# Test 3: Cart page is accessible
log_test "Loading cart page..."
RESPONSE=$(curl -s "$BASE_URL/cart")
if echo "$RESPONSE" | grep -q "Shopping Cart"; then
    log_pass "Cart page loaded"
else
    log_fail "Cart page did not load correctly"
fi

# Test 4: Add item to cart (anonymous)
log_test "Adding item to cart as anonymous user..."
RESPONSE=$(curl -s -X POST \
    -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
    -d "product_id=1&quantity=2" \
    -w "%{http_code}" \
    "$BASE_URL/cart/add")
if echo "$RESPONSE" | grep -q "204"; then
    log_pass "Item added to cart"
else
    log_fail "Failed to add item to cart (HTTP $RESPONSE)"
fi

# Test 5: Cart count shows items
log_test "Checking cart count..."
RESPONSE=$(curl -s -b "$COOKIE_JAR" "$BASE_URL/partials/cart-count")
if echo "$RESPONSE" | grep -qE "\([1-9][0-9]*\)"; then
    log_pass "Cart count shows items: $RESPONSE"
else
    log_fail "Cart count incorrect: $RESPONSE"
fi

# Test 6: View cart shows items
log_test "Viewing cart..."
RESPONSE=$(curl -s -b "$COOKIE_JAR" "$BASE_URL/cart")
if echo "$RESPONSE" | grep -q "Shopping Cart"; then
    log_pass "Cart page shows items"
else
    log_fail "Cart page empty or error"
fi

# Test 7: User registration
log_test "Registering new user..."
RESPONSE=$(curl -s -X POST \
    -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
    -d "email=$TEST_EMAIL&password=$TEST_PASSWORD" \
    -w "%{http_code}" \
    -L \
    "$BASE_URL/signup/process")
if echo "$RESPONSE" | grep -q "200"; then
    log_pass "User registered successfully"
else
    log_fail "User registration failed"
fi

# Test 8: User login
log_test "Logging in..."
RESPONSE=$(curl -s -X POST \
    -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
    -d "email=$TEST_EMAIL&password=$TEST_PASSWORD" \
    -w "%{http_code}" \
    -L \
    "$BASE_URL/login/process")
if echo "$RESPONSE" | grep -q "200"; then
    log_pass "User logged in successfully"
else
    log_fail "User login failed"
fi

# Test 9: Add item to cart (authenticated)
log_test "Adding item to cart as authenticated user..."
RESPONSE=$(curl -s -X POST \
    -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
    -d "product_id=2&quantity=3" \
    -w "%{http_code}" \
    "$BASE_URL/cart/add")
if echo "$RESPONSE" | grep -q "204"; then
    log_pass "Item added to cart (authenticated)"
else
    log_fail "Failed to add item to cart (authenticated)"
fi

# Test 10: Checkout page accessible
log_test "Loading checkout page..."
RESPONSE=$(curl -s -b "$COOKIE_JAR" -w "%{http_code}" "$BASE_URL/checkout")
if echo "$RESPONSE" | grep -q "Order Summary"; then
    log_pass "Checkout page loaded"
else
    log_fail "Checkout page did not load"
fi

# Test 11: Database connectivity
log_test "Checking database connectivity..."
DB_TEST=$(docker compose exec -T db psql -U user -d bookstore -c "SELECT 1;" 2>&1)
if echo "$DB_TEST" | grep -q "1 row"; then
    log_pass "Database is accessible"
else
    log_fail "Database connection failed"
fi

# Test 12: Check for duplicate cart items
log_test "Checking for duplicate cart items in database..."
DUPLICATES=$(docker compose exec -T db psql -U user -d bookstore -t -c "
    SELECT COUNT(*) FROM (
        SELECT user_id, product_id, COUNT(*) 
        FROM cart_items 
        WHERE user_id IS NOT NULL 
        GROUP BY user_id, product_id 
        HAVING COUNT(*) > 1
    ) dups;" | tr -d ' \n')
if [ "$DUPLICATES" = "0" ]; then
    log_pass "No duplicate cart items found"
else
    log_fail "Found $DUPLICATES duplicate cart items!"
fi

# Test 13: Verify unique constraints exist
log_test "Checking database constraints..."
CONSTRAINTS=$(docker compose exec -T db psql -U user -d bookstore -t -c "
    SELECT COUNT(*) FROM pg_indexes 
    WHERE tablename = 'cart_items' 
    AND indexname LIKE 'idx_cart_items_%';" | tr -d ' \n')
if [ "$CONSTRAINTS" -ge "2" ]; then
    log_pass "Cart item unique constraints exist"
else
    log_fail "Missing cart item unique constraints"
fi

# Test 14: Cart merging on login (anonymous → authenticated)
log_test "Testing cart merge on login..."
# Create new cookies for fresh session
MERGE_COOKIE=$(mktemp)
MERGE_EMAIL="merge_test_$(date +%s)@example.com"

# Add items as anonymous user
curl -s -X POST -b "$MERGE_COOKIE" -c "$MERGE_COOKIE" \
    -d "product_id=1&quantity=3" "$BASE_URL/cart/add" > /dev/null

# Register and auto-login (this should merge the cart)
curl -s -X POST -b "$MERGE_COOKIE" -c "$MERGE_COOKIE" \
    -d "email=$MERGE_EMAIL&password=$TEST_PASSWORD" \
    "$BASE_URL/signup/process" > /dev/null

# Check cart count (should still have items)
CART_COUNT=$(curl -s -b "$MERGE_COOKIE" "$BASE_URL/partials/cart-count")
if echo "$CART_COUNT" | grep -qE "\([1-9][0-9]*\)"; then
    log_pass "Cart merged on signup: $CART_COUNT"
else
    log_fail "Cart not merged on signup"
fi

# Cleanup
rm -f "$MERGE_COOKIE"

# Test 15: Cart merging with existing items
log_test "Testing cart merge with existing user cart..."
MERGE2_COOKIE=$(mktemp)
MERGE2_EMAIL="merge2_test_$(date +%s)@example.com"

# Register user
curl -s -X POST -b "$MERGE2_COOKIE" -c "$MERGE2_COOKIE" \
    -d "email=$MERGE2_EMAIL&password=$TEST_PASSWORD" \
    "$BASE_URL/signup/process" > /dev/null

# Add item to authenticated cart
curl -s -X POST -b "$MERGE2_COOKIE" -c "$MERGE2_COOKIE" \
    -d "product_id=1&quantity=2" "$BASE_URL/cart/add" > /dev/null

# Logout
curl -s -b "$MERGE2_COOKIE" -c "$MERGE2_COOKIE" "$BASE_URL/logout" > /dev/null

# Add different quantity as anonymous
curl -s -X POST -b "$MERGE2_COOKIE" -c "$MERGE2_COOKIE" \
    -d "product_id=1&quantity=3" "$BASE_URL/cart/add" > /dev/null

# Login (should merge: 2 + 3 = 5)
curl -s -X POST -b "$MERGE2_COOKIE" -c "$MERGE2_COOKIE" \
    -d "email=$MERGE2_EMAIL&password=$TEST_PASSWORD" \
    "$BASE_URL/login/process" > /dev/null

# Check cart count (should be 5)
MERGED_COUNT=$(curl -s -b "$MERGE2_COOKIE" "$BASE_URL/partials/cart-count")
if echo "$MERGED_COUNT" | grep -q "(5)"; then
    log_pass "Cart quantities merged correctly: $MERGED_COUNT"
else
    log_fail "Cart quantities not merged correctly: $MERGED_COUNT (expected (5))"
fi

# Cleanup
rm -f "$MERGE2_COOKIE"

# Test 17: Redis connectivity
log_test "Checking Redis connectivity..."
if docker compose exec -T redis redis-cli ping 2>/dev/null | grep -q "PONG"; then
    log_pass "Redis is accessible"
else
    log_fail "Redis is not accessible"
fi

# Test 18: Redis session storage
log_test "Checking Redis session storage..."
SESSION_COUNT=$(docker compose exec -T redis redis-cli KEYS "session:*" 2>/dev/null | wc -l)
if [ "$SESSION_COUNT" -gt 0 ]; then
    log_pass "Redis contains $SESSION_COUNT session(s)"
else
    log_fail "No sessions found in Redis"
fi

# Test 19: Redis caching
log_test "Checking Redis cache keys..."
CACHE_KEYS=$(docker compose exec -T redis redis-cli KEYS "*" 2>/dev/null | grep -v "session:" | wc -l)
if [ "$CACHE_KEYS" -gt 0 ]; then
    log_pass "Redis cache contains $CACHE_KEYS key(s)"
else
    log_fail "No cache keys found in Redis"
fi

# Test 20: Elasticsearch connectivity
log_test "Checking Elasticsearch connectivity..."
if curl -s http://localhost:9200/_cluster/health 2>/dev/null | grep -q "status"; then
    log_pass "Elasticsearch is accessible"
else
    log_fail "Elasticsearch is not accessible"
fi

# Test 21: Elasticsearch product index
log_test "Checking Elasticsearch product index..."
INDEX_COUNT=$(curl -s "http://localhost:9200/products/_count" 2>/dev/null | grep -o '"count":[0-9]*' | cut -d':' -f2)
if [ ! -z "$INDEX_COUNT" ] && [ "$INDEX_COUNT" -gt 0 ]; then
    log_pass "Elasticsearch has $INDEX_COUNT products indexed"
else
    log_fail "No products found in Elasticsearch index"
fi

# Summary
echo ""
echo "========================================="
echo "Test Summary"
echo "========================================="
echo "Tests Run:    $TESTS_RUN"
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
echo "========================================="

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi

