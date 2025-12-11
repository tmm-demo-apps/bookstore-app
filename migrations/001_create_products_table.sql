-- Create the initial products table with basic schema
-- Additional fields (sku, stock_quantity, image_url, category_id, status) are added in 007_phase1_schema_expansion.sql
-- Seed data is in 008_seed_data.sql
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    price DECIMAL(10, 2) NOT NULL
);
