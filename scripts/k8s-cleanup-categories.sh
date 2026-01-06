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
echo "🔄 Consolidating duplicate categories..."
kubectl exec -it -n bookstore postgres-0 -- psql -U bookstore_user -d bookstore -c "
-- For each duplicate category name, move all products to the category with the lowest ID
-- Then delete the empty duplicate categories

DO \$\$
DECLARE
    cat_name TEXT;
    keep_id INT;
    dup_id INT;
BEGIN
    -- Loop through each duplicate category name
    FOR cat_name IN 
        SELECT name 
        FROM categories 
        GROUP BY name 
        HAVING COUNT(*) > 1
    LOOP
        -- Get the ID to keep (lowest ID)
        SELECT MIN(id) INTO keep_id 
        FROM categories 
        WHERE name = cat_name;
        
        RAISE NOTICE 'Processing category: % (keeping ID: %)', cat_name, keep_id;
        
        -- Move all products from duplicate categories to the one we're keeping
        FOR dup_id IN 
            SELECT id 
            FROM categories 
            WHERE name = cat_name AND id != keep_id
        LOOP
            RAISE NOTICE '  Moving products from ID % to ID %', dup_id, keep_id;
            
            UPDATE products 
            SET category_id = keep_id 
            WHERE category_id = dup_id;
            
            -- Delete the now-empty duplicate category
            DELETE FROM categories WHERE id = dup_id;
            
            RAISE NOTICE '  Deleted duplicate category ID %', dup_id;
        END LOOP;
    END LOOP;
END \$\$;
"

echo ""
echo "📊 Final categories:"
kubectl exec -it -n bookstore postgres-0 -- psql -U bookstore_user -d bookstore -c "
SELECT c.id, c.name, COUNT(p.id) as product_count
FROM categories c
LEFT JOIN products p ON c.id = p.category_id
GROUP BY c.id, c.name
ORDER BY c.name;
"

echo ""
echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          ✅ CLEANUP COMPLETE                                               ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"

