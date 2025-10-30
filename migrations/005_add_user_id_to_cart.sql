ALTER TABLE cart_items
ADD COLUMN user_id INTEGER REFERENCES users(id),
ADD CONSTRAINT session_or_user CHECK (
    (user_id IS NOT NULL AND session_id IS NULL) OR
    (user_id IS NULL AND session_id IS NOT NULL)
);
