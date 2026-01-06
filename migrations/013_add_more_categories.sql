-- Add additional categories for complete book collection
INSERT INTO categories (name, description) VALUES
('Poetry', 'Poems and poetic works'),
('History', 'Historical accounts and biographies'),
('Political Science', 'Political theory and governance')
ON CONFLICT DO NOTHING;

