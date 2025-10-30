CREATE TABLE cart_items (
    id SERIAL PRIMARY KEY,
    session_id VARCHAR(255) NOT NULL,
    book_id INTEGER REFERENCES books(id),
    quantity INTEGER NOT NULL
);
