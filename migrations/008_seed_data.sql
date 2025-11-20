-- Seed Products
INSERT INTO products (name, description, price, sku, stock_quantity, category_id, image_url, status) VALUES
-- Fiction
('The Great Gatsby', 'A novel by F. Scott Fitzgerald.', 15.99, 'B-FIC-001', 50, 1, 'https://via.placeholder.com/150?text=Gatsby', 'active'),
('1984', 'Dystopian social science fiction novel and cautionary tale by George Orwell.', 12.99, 'B-FIC-002', 100, 1, 'https://via.placeholder.com/150?text=1984', 'active'),
('To Kill a Mockingbird', 'Novel by Harper Lee.', 14.50, 'B-FIC-003', 30, 1, 'https://via.placeholder.com/150?text=Mockingbird', 'active'),
('Pride and Prejudice', 'Romantic novel of manners written by Jane Austen.', 9.99, 'B-FIC-004', 20, 1, 'https://via.placeholder.com/150?text=Pride', 'active'),
('The Catcher in the Rye', 'Novel by J. D. Salinger.', 11.00, 'B-FIC-005', 45, 1, 'https://via.placeholder.com/150?text=Catcher', 'active'),

-- Non-Fiction
('Sapiens: A Brief History of Humankind', 'Book by Yuval Noah Harari.', 22.00, 'B-NF-001', 60, 2, 'https://via.placeholder.com/150?text=Sapiens', 'active'),
('Educated', 'Memoir by Tara Westover.', 18.00, 'B-NF-002', 40, 2, 'https://via.placeholder.com/150?text=Educated', 'active'),
('Becoming', 'Memoir by Michelle Obama.', 25.00, 'B-NF-003', 80, 2, 'https://via.placeholder.com/150?text=Becoming', 'active'),
('Thinking, Fast and Slow', 'Book by Daniel Kahneman.', 19.50, 'B-NF-004', 55, 2, 'https://via.placeholder.com/150?text=Thinking', 'active'),
('Silent Spring', 'Environmental science book by Rachel Carson.', 14.00, 'B-NF-005', 15, 2, 'https://via.placeholder.com/150?text=Silent', 'active'),

-- Science
('A Brief History of Time', 'Popular-science book on cosmology by Stephen Hawking.', 16.50, 'B-SCI-001', 25, 3, 'https://via.placeholder.com/150?text=Hawking', 'active'),
('The Selfish Gene', 'Book on evolution by Richard Dawkins.', 17.00, 'B-SCI-002', 35, 3, 'https://via.placeholder.com/150?text=Gene', 'active'),
('Cosmos', 'Popular science book by Carl Sagan.', 20.00, 'B-SCI-003', 50, 3, 'https://via.placeholder.com/150?text=Cosmos', 'active'),
('The Gene: An Intimate History', 'Book by Siddhartha Mukherjee.', 21.00, 'B-SCI-004', 10, 3, 'https://via.placeholder.com/150?text=TheGene', 'active'),
('Astrophysics for People in a Hurry', 'Book by Neil deGrasse Tyson.', 13.00, 'B-SCI-005', 100, 3, 'https://via.placeholder.com/150?text=Tyson', 'active'),

-- Technology
('The Pragmatic Programmer', 'Book about computer programming and software engineering.', 35.00, 'B-TECH-001', 70, 4, 'https://via.placeholder.com/150?text=Pragmatic', 'active'),
('Clean Code', 'Handbook of Agile Software Craftsmanship.', 32.00, 'B-TECH-002', 65, 4, 'https://via.placeholder.com/150?text=CleanCode', 'active'),
('Introduction to Algorithms', 'Standard textbook on algorithms.', 80.00, 'B-TECH-003', 10, 4, 'https://via.placeholder.com/150?text=Algorithms', 'active'),
('Design Patterns', 'Elements of Reusable Object-Oriented Software.', 45.00, 'B-TECH-004', 30, 4, 'https://via.placeholder.com/150?text=Patterns', 'active'),
('The Phoenix Project', 'Novel about IT, DevOps, and helping your business win.', 24.00, 'B-TECH-005', 90, 4, 'https://via.placeholder.com/150?text=Phoenix', 'active');

