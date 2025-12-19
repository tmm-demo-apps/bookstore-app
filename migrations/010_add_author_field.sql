-- Migration: Add author field to products table
-- Date: 2025-12-19
-- Description: Add author field for book products, making it searchable

-- Add author column to products table
ALTER TABLE products 
ADD COLUMN author VARCHAR(255);

-- Create index for author search performance
CREATE INDEX idx_products_author ON products(author);

-- Update existing book products with sample authors
UPDATE products SET author = 'George Orwell' WHERE name = '1984';
UPDATE products SET author = 'Aldous Huxley' WHERE name = 'Brave New World';
UPDATE products SET author = 'Ray Bradbury' WHERE name = 'Fahrenheit 451';
UPDATE products SET author = 'Richard Dawkins' WHERE name = 'The Selfish Gene';
UPDATE products SET author = 'James D. Watson' WHERE name = 'The Gene: An Intimate History';
UPDATE products SET author = 'Carl Sagan' WHERE name = 'Cosmos';
UPDATE products SET author = 'Neil deGrasse Tyson' WHERE name = 'Astrophysics for People in a Hurry';
UPDATE products SET author = 'Stephen Hawking' WHERE name = 'A Brief History of Time';
UPDATE products SET author = 'Richard Feynman' WHERE name = 'Surely You''re Joking, Mr. Feynman!';
UPDATE products SET author = 'Daniel Kahneman' WHERE name = 'Thinking, Fast and Slow';
UPDATE products SET author = 'Yuval Noah Harari' WHERE name = 'Sapiens: A Brief History of Humankind';
UPDATE products SET author = 'Jared Diamond' WHERE name = 'Guns, Germs, and Steel';
UPDATE products SET author = 'Malcolm Gladwell' WHERE name = 'The Tipping Point';
UPDATE products SET author = 'Steven Pinker' WHERE name = 'The Better Angels of Our Nature';
UPDATE products SET author = 'Sam Harris' WHERE name = 'The Moral Landscape';
UPDATE products SET author = 'Christopher Hitchens' WHERE name = 'God Is Not Great';
UPDATE products SET author = 'Richard Dawkins' WHERE name = 'The God Delusion';
UPDATE products SET author = 'Daniel Dennett' WHERE name = 'Breaking the Spell';
UPDATE products SET author = 'Sam Harris' WHERE name = 'Letter to a Christian Nation';
UPDATE products SET author = 'Victor Frankl' WHERE name = 'Man''s Search for Meaning';
UPDATE products SET author = 'Jordan Peterson' WHERE name = '12 Rules for Life';
UPDATE products SET author = 'Dale Carnegie' WHERE name = 'How to Win Friends and Influence People';
UPDATE products SET author = 'Stephen Covey' WHERE name = 'The 7 Habits of Highly Effective People';

-- Additional books that were missing authors
UPDATE products SET author = 'F. Scott Fitzgerald' WHERE name = 'The Great Gatsby';
UPDATE products SET author = 'Harper Lee' WHERE name = 'To Kill a Mockingbird';
UPDATE products SET author = 'Jane Austen' WHERE name = 'Pride and Prejudice';
UPDATE products SET author = 'J.D. Salinger' WHERE name = 'The Catcher in the Rye';
UPDATE products SET author = 'Tara Westover' WHERE name = 'Educated';
UPDATE products SET author = 'Michelle Obama' WHERE name = 'Becoming';
UPDATE products SET author = 'Rachel Carson' WHERE name = 'Silent Spring';
UPDATE products SET author = 'Andrew Hunt & David Thomas' WHERE name = 'The Pragmatic Programmer';
UPDATE products SET author = 'Robert C. Martin' WHERE name = 'Clean Code';
UPDATE products SET author = 'Thomas H. Cormen et al.' WHERE name = 'Introduction to Algorithms';
UPDATE products SET author = 'Gang of Four' WHERE name = 'Design Patterns';
UPDATE products SET author = 'Gene Kim et al.' WHERE name = 'The Phoenix Project';

-- Add comment for documentation
COMMENT ON COLUMN products.author IS 'Author name for book products, searchable field';

