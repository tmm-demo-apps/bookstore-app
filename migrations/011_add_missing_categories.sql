-- Add missing categories for Project Gutenberg books
INSERT INTO categories (name, description) VALUES
('Philosophy', 'Philosophical works and treatises'),
('Science Fiction', 'Science fiction and speculative fiction'),
('Drama', 'Plays and dramatic works')
ON CONFLICT DO NOTHING;

