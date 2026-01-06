-- Fix cart_items session_id constraint
-- The session_id should be nullable when user_id is present

-- First, drop the CHECK constraint if it exists
ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS session_or_user;

-- Make session_id nullable
ALTER TABLE cart_items ALTER COLUMN session_id DROP NOT NULL;

-- Re-add the constraint to ensure either session_id OR user_id is present
ALTER TABLE cart_items ADD CONSTRAINT session_or_user CHECK (
    (user_id IS NOT NULL AND session_id IS NULL) OR
    (user_id IS NULL AND session_id IS NOT NULL) OR
    (user_id IS NOT NULL AND session_id IS NOT NULL)
);

