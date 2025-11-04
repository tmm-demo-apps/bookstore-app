# Testing Strategy for E-commerce Application

## Overview
This document outlines our testing approach to prevent regressions and ensure all features work correctly after changes.

## Test Categories

### 1. Smoke Tests (Run After Every Change)
Quick tests to verify core functionality isn't broken.

#### Product Listing
- [ ] Products page loads at http://localhost:8080
- [ ] All products are displayed
- [ ] Prices are formatted correctly
- [ ] Quantity controls (+/-) are visible

#### Cart Operations (Anonymous User)
- [ ] Add product to cart (quantity 1)
- [ ] Cart badge updates with correct count
- [ ] Add product with quantity > 1
- [ ] Cart badge shows total quantity
- [ ] View cart page shows correct items and quantities
- [ ] Adjust quantity with +/- buttons
- [ ] Cart count updates after adjustment
- [ ] Remove item from cart
- [ ] Cart badge updates to 0

#### Cart Operations (Authenticated User)
- [ ] Register new account
- [ ] Login successfully
- [ ] Add products to cart
- [ ] Cart persists after page refresh
- [ ] Adjust quantities in cart
- [ ] No duplicate cart items created
- [ ] Remove items from cart
- [ ] Logout and back in - cart persists

#### Checkout Flow
- [ ] Click "Proceed to Checkout" from cart
- [ ] Redirected to login if not authenticated
- [ ] After login, redirected back to checkout
- [ ] Checkout page shows all cart items
- [ ] Quantities displayed correctly
- [ ] Subtotals calculated correctly (price × quantity)
- [ ] Total calculated correctly
- [ ] Click "Confirm Order"
- [ ] Redirected to confirmation page
- [ ] Cart is now empty
- [ ] Order was created in database

### 2. Regression Tests (Run Before Commits)
Tests that verify previously fixed bugs don't resurface.

#### Bug: Duplicate Cart Items (Fixed Nov 4, 2025)
- [ ] Add same product multiple times
- [ ] Verify cart shows single line with correct total quantity
- [ ] Adjust quantity with +/- buttons
- [ ] Verify quantity changes by exactly 1
- [ ] Check database: `SELECT * FROM cart_items WHERE user_id = X`
- [ ] Should be only ONE row per product per user

#### Bug: Checkout User/Session Mismatch (Fixed Nov 4, 2025)
- [ ] Login as user
- [ ] Add items to cart
- [ ] Click "Proceed to Checkout"
- [ ] Verify checkout page shows items (not empty)
- [ ] Complete order
- [ ] Verify cart is cleared

### 3. Integration Tests
Test interactions between components.

#### Anonymous to Authenticated Flow
- [ ] Add items to cart as anonymous user
- [ ] Note cart count
- [ ] Register/Login
- [ ] Verify cart is still populated (ideally - not currently implemented)
- [ ] Complete checkout

#### Cross-Browser Tests
- [ ] Test in Chrome
- [ ] Test in Firefox
- [ ] Test in Safari (if on Mac)
- [ ] Verify quantity input works in all browsers
- [ ] Verify hover cart dropdown works in all browsers

### 4. Database Integrity Tests

#### Cart Items Constraints
```sql
-- Should return 0 (no duplicates for same user+product)
SELECT user_id, product_id, COUNT(*) as count 
FROM cart_items 
WHERE user_id IS NOT NULL 
GROUP BY user_id, product_id 
HAVING COUNT(*) > 1;

-- Should return 0 (no duplicates for same session+product)
SELECT session_id, product_id, COUNT(*) as count 
FROM cart_items 
WHERE session_id IS NOT NULL 
GROUP BY session_id, product_id 
HAVING COUNT(*) > 1;

-- Verify unique constraints exist
SELECT indexname, indexdef 
FROM pg_indexes 
WHERE tablename = 'cart_items' 
AND indexname LIKE 'idx_cart_items_%';
```

## Manual Test Script

### Full End-to-End Test (10 minutes)
Run this complete flow before every commit:

```
1. SETUP
   - Ensure app is running: docker compose up -d
   - Open http://localhost:8080
   - Open browser console (F12) to check for errors

2. ANONYMOUS CART TEST
   - Add "Product A" qty 3
   - Add "Product B" qty 5
   - Verify cart badge shows (8)
   - Open cart page
   - Verify Product A shows qty 3, Product B shows qty 5
   - Increase Product A to 4 using + button
   - Verify cart badge shows (9)
   - Decrease Product B to 4 using - button
   - Verify cart badge shows (8)
   - Remove Product A
   - Verify cart badge shows (4)
   - Verify only Product B remains

3. AUTHENTICATION TEST
   - Click "Sign Up"
   - Create account: test_user_[timestamp]@example.com
   - Verify logged in (nav shows username)
   - Verify cart still shows Product B qty 4

4. AUTHENTICATED CART TEST
   - Add "Product C" qty 2
   - Verify cart badge shows (6)
   - Refresh page (F5)
   - Verify cart badge still shows (6)
   - View cart
   - Verify Product B qty 4, Product C qty 2
   - Add Product C again qty 3 from products page
   - Verify cart shows Product C qty 5 (not duplicate row)

5. CHECKOUT TEST
   - Click "Proceed to Checkout"
   - Verify checkout page shows:
     * Product B: qty 4, subtotal correct
     * Product C: qty 5, subtotal correct
     * Total = sum of subtotals
   - Click "Confirm Order"
   - Verify confirmation page appears
   - Click "Continue Shopping"
   - Verify cart badge shows (0)

6. DATABASE VERIFICATION
   docker compose exec db psql -U user -d bookstore -c "
     SELECT user_id, product_id, COUNT(*) 
     FROM cart_items 
     GROUP BY user_id, product_id 
     HAVING COUNT(*) > 1;"
   - Should return 0 rows (no duplicates)

7. LOGOUT/LOGIN TEST
   - Logout
   - Login with same credentials
   - Verify cart is empty (post-checkout state persists)

PASS CRITERIA: All steps complete without errors
```

## Automated Test Script

See `test-smoke.sh` for an automated smoke test using curl.

## Test Failure Protocol

When a test fails:

1. **Document the failure**
   - Screenshot or error message
   - Steps to reproduce
   - Expected vs actual behavior

2. **Check recent changes**
   - Run `git log --oneline -5`
   - Review files changed in last commit
   - Identify which change likely caused the issue

3. **Create a fix**
   - Make minimal change to fix the bug
   - Run ALL smoke tests before committing
   - Update this test document if needed

4. **Add regression test**
   - Add new test case to prevent future recurrence
   - Document the bug in TESTING.md

## CI/CD Integration (Future)

When implementing automated testing:
- Run smoke tests on every PR
- Run full regression suite before merging to main
- Run database integrity checks after migrations
- Fail builds if any tests fail

## Test Data Management

### Test Products
Ensure these products exist in the database:
- Product A: ID 1
- Product B: ID 2  
- Product C: ID 3

### Test Users
Create dedicated test users:
- test_user@example.com
- test_admin@example.com

## Performance Benchmarks

Monitor these after changes:
- Page load time: < 500ms
- Cart add operation: < 200ms
- Checkout page load: < 500ms
- Database queries: < 50ms per query

## Known Limitations

Current testing gaps (to be addressed):
- No automated tests yet (all manual)
- No load testing
- No security testing (SQL injection, XSS)
- No session persistence testing across server restarts
- No mobile responsive testing

