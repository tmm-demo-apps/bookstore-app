# Quick Test Checklist

## Before Every Commit

```bash
# 1. Run automated smoke tests
./test-smoke.sh

# 2. If all pass, commit
git add -A
git commit -m "your message"

# 3. If tests fail, fix the issue first!
```

## Manual Spot Check (2 minutes)

When smoke tests pass, quickly verify in browser:

1. **Open** http://localhost:8080
2. **Add item** to cart (with quantity > 1)
3. **Check** cart badge shows correct count
4. **Login** or **Sign Up**
5. **View cart** - items still there?
6. **Adjust quantity** with +/- buttons
7. **Click** "Proceed to Checkout"
8. **Verify** checkout shows items with quantities
9. **Confirm Order**
10. **Check** cart is now empty

✅ If all steps work → Safe to push!  
❌ If anything fails → Check `diary.md` for similar bugs

## Quick Database Check

```bash
# Check for duplicate cart items (should return 0)
docker compose exec db psql -U user -d bookstore -c "
  SELECT COUNT(*) FROM (
    SELECT user_id, product_id, COUNT(*) 
    FROM cart_items 
    GROUP BY user_id, product_id 
    HAVING COUNT(*) > 1
  ) dups;"
```

## Common Issues

### Server not responding
```bash
docker compose down && docker compose up --build -d
```

### Tests failing after migration
```bash
# Check if migrations applied
docker compose exec db psql -U user -d bookstore -c "\d cart_items"
```

### Cart behavior weird
1. Clear browser cookies
2. Check console for JavaScript errors (F12)
3. Verify unique constraints exist in database

## Files to Check After Changes

| Changed File | Run These Tests |
|-------------|----------------|
| `cart.go` | All cart operations + duplicate check |
| `checkout.go` | Checkout flow (steps 6-10 above) |
| `products.go` | Product listing + add to cart |
| `*.html` templates | Full manual flow in browser |
| Database migrations | Database integrity tests |
| Any handler | Full smoke test suite |

## Test Output

**Good:**
```
Tests Run:    13
Tests Passed: 13
Tests Failed: 0
All tests passed!
```

**Bad:**
```
[FAIL] Item added to cart
Tests Failed: 1
```
→ Don't commit until fixed!

## Emergency Rollback

If you committed broken code:
```bash
git log --oneline -5          # Find good commit
git revert HEAD               # Undo last commit
./test-smoke.sh              # Verify tests pass
git push                      # Push the revert
```

---

**Remember:** A few minutes of testing saves hours of debugging! 🧪✅

