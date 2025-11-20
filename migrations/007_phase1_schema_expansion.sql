-- Categories
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT
);

-- Products Expansion
ALTER TABLE products ADD COLUMN sku VARCHAR(50) UNIQUE;
ALTER TABLE products ADD COLUMN stock_quantity INTEGER DEFAULT 0;
ALTER TABLE products ADD COLUMN image_url VARCHAR(255);
ALTER TABLE products ADD COLUMN category_id INTEGER REFERENCES categories(id);
ALTER TABLE products ADD COLUMN status VARCHAR(20) DEFAULT 'active';

-- Users Expansion
ALTER TABLE users ADD COLUMN full_name VARCHAR(255);
ALTER TABLE users ADD COLUMN role VARCHAR(20) DEFAULT 'customer';
ALTER TABLE users ADD COLUMN created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- Orders Expansion
ALTER TABLE orders ADD COLUMN user_id INTEGER REFERENCES users(id);
ALTER TABLE orders ADD COLUMN total_amount DECIMAL(10, 2);
ALTER TABLE orders ADD COLUMN status VARCHAR(20) DEFAULT 'pending';
ALTER TABLE orders ADD COLUMN shipping_info JSONB;

-- Seed Categories
INSERT INTO categories (name, description) VALUES
('Fiction', 'Novels and stories'),
('Non-Fiction', 'Factual books and biographies'),
('Science', 'Physics, Chemistry, Biology'),
('Technology', 'Computers and Programming');

