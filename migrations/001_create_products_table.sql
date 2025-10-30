CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL
);

INSERT INTO products (name, description, price) VALUES
('Sample Product 1', 'This is a description for the first sample product.', 10.99),
('Sample Product 2', 'This is a description for the second sample product.', 25.99),
('Sample Product 3', 'This is a description for the third sample product.', 15.99);
