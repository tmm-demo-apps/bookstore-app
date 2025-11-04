-- Migration: Consolidate duplicate cart items
-- This fixes a bug where multiple cart item rows could exist for the same product+user/session
-- After this migration, each user/session will have at most one row per product

-- Step 1: Create a temporary table with consolidated data
CREATE TEMP TABLE consolidated_cart_items AS
SELECT 
    MIN(id) as id,
    user_id,
    session_id,
    product_id,
    SUM(quantity) as quantity
FROM cart_items
GROUP BY 
    COALESCE(user_id, 0),
    COALESCE(session_id, ''),
    product_id
HAVING COUNT(*) > 0;

-- Step 2: Clear the cart_items table
DELETE FROM cart_items;

-- Step 3: Insert consolidated data back
INSERT INTO cart_items (id, user_id, session_id, product_id, quantity)
SELECT id, user_id, session_id, product_id, 
       CASE WHEN quantity > 99 THEN 99 ELSE quantity END as quantity
FROM consolidated_cart_items;

-- Step 4: Reset the sequence for the id column to avoid conflicts
SELECT setval('cart_items_id_seq', COALESCE((SELECT MAX(id) FROM cart_items), 1));

-- Step 5: Add unique constraints to prevent future duplicates
-- Add unique constraint for user_id and product_id
CREATE UNIQUE INDEX IF NOT EXISTS idx_cart_items_user_product 
    ON cart_items (user_id, product_id) 
    WHERE user_id IS NOT NULL;

-- Add unique constraint for session_id and product_id
CREATE UNIQUE INDEX IF NOT EXISTS idx_cart_items_session_product 
    ON cart_items (session_id, product_id) 
    WHERE session_id IS NOT NULL;

