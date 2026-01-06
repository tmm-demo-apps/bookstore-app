#!/bin/bash
set -e

echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          Clean Up Duplicate Categories                                     ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""

echo "📊 Current categories:"
kubectl exec -it -n bookstore postgres-0 -- psql -U bookstore_user -d bookstore -c "
SELECT id, name FROM categories ORDER BY id;
"

echo ""
echo "🔍 Looking for duplicates..."
kubectl exec -it -n bookstore postgres-0 -- psql -U bookstore_user -d bookstore -c "
SELECT name, COUNT(*) as count
FROM categories
GROUP BY name
HAVING COUNT(*) > 1;
"

echo ""
echo "🗑️  Removing duplicate categories (keeping lowest ID)..."
kubectl exec -it -n bookstore postgres-0 -- psql -U bookstore_user -d bookstore -c "
-- For each duplicate category, keep the one with the lowest ID
DELETE FROM categories
WHERE id NOT IN (
    SELECT MIN(id)
    FROM categories
    GROUP BY name
);
"

echo ""
echo "📊 Final categories:"
kubectl exec -it -n bookstore postgres-0 -- psql -U bookstore_user -d bookstore -c "
SELECT id, name, COUNT(p.id) as product_count
FROM categories c
LEFT JOIN products p ON c.id = p.category_id
GROUP BY c.id, c.name
ORDER BY c.name;
"

echo ""
echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          ✅ CLEANUP COMPLETE                                               ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"

