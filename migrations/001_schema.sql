-- DemoApp Database Schema
-- Consolidated migration with complete table definitions
-- All tables created with final schema from day 1

-- Categories
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT
);

-- Products (all fields from day 1)
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    sku VARCHAR(50) UNIQUE,
    stock_quantity INTEGER DEFAULT 0,
    image_url VARCHAR(255),
    category_id INTEGER REFERENCES categories(id),
    status VARCHAR(20) DEFAULT 'active',
    author VARCHAR(255),
    popularity_score INTEGER DEFAULT 0  -- Gutenberg download count, used for sorting
);

-- Index for popularity sorting
CREATE INDEX idx_products_popularity ON products(popularity_score DESC);
CREATE INDEX idx_products_category_popularity ON products(category_id, popularity_score DESC);

-- Users (complete schema)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    role VARCHAR(20) DEFAULT 'customer',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Cart Items (correct constraint from start - session_id nullable when user_id present)
CREATE TABLE cart_items (
    id SERIAL PRIMARY KEY,
    session_id VARCHAR(255),
    user_id INTEGER REFERENCES users(id),
    product_id INTEGER REFERENCES products(id),
    quantity INTEGER NOT NULL DEFAULT 1,
    CONSTRAINT session_or_user CHECK (
        session_id IS NOT NULL OR user_id IS NOT NULL
    )
);

-- Orders (complete schema)
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    session_id VARCHAR(255),
    user_id INTEGER REFERENCES users(id),
    total_amount DECIMAL(10, 2),
    status VARCHAR(20) DEFAULT 'pending',
    shipping_info JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id),
    product_id INTEGER REFERENCES products(id),
    quantity INTEGER NOT NULL,
    price DECIMAL(10, 2) NOT NULL
);

-- Reviews (complete schema with indexes)
CREATE TABLE reviews (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    title VARCHAR(255),
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(product_id, user_id)
);

-- Indexes for efficient queries
CREATE INDEX idx_reviews_product ON reviews(product_id);
CREATE INDEX idx_reviews_user ON reviews(user_id);
CREATE INDEX idx_reviews_rating ON reviews(rating);
CREATE INDEX idx_reviews_created_at ON reviews(created_at DESC);

-- Comments for documentation
COMMENT ON TABLE categories IS 'Product categories for organizing books';
COMMENT ON TABLE products IS 'Book products with metadata from Project Gutenberg';
COMMENT ON COLUMN products.popularity_score IS 'Gutenberg 30-day download count for sorting';
COMMENT ON TABLE users IS 'User accounts for authentication and orders';
COMMENT ON TABLE cart_items IS 'Shopping cart items - supports both anonymous (session) and authenticated users';
COMMENT ON TABLE orders IS 'Customer orders';
COMMENT ON TABLE order_items IS 'Individual items within an order';
COMMENT ON TABLE reviews IS 'Product reviews and ratings from users';

