CREATE TABLE books (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    author VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL
);

INSERT INTO books (title, author, price) VALUES
('The Hobbit', 'J.R.R. Tolkien', 10.99),
('The Lord of the Rings', 'J.R.R. Tolkien', 25.99),
('The Silmarillion', 'J.R.R. Tolkien', 15.99);
