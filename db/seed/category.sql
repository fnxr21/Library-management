-- Example data for category-service (database: category_mst).
--
-- Start category-service once first so the table exists. Safe to re-run.
-- IDs are fixed so book.sql can refer to them (book_category.category_id).

BEGIN;

INSERT INTO category (id, created_at, updated_at, name, description) VALUES
  (1, NOW(), NOW(), 'Fiction',          'Stories that are imagined.'),
  (2, NOW(), NOW(), 'Historical',       'Set in or about a real past period.'),
  (3, NOW(), NOW(), 'Fantasy',          'Magic, myths and invented worlds.'),
  (4, NOW(), NOW(), 'Dystopia',         'Societies gone wrong.'),
  (5, NOW(), NOW(), 'History',          'Non-fiction about the past.'),
  (6, NOW(), NOW(), 'Technology',       'Software, computing and engineering.'),
  (7, NOW(), NOW(), 'Self-Development', 'Habits, productivity and personal growth.'),
  (8, NOW(), NOW(), 'Young Adult',      'Written for teenage readers.')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('category', 'id'), (SELECT MAX(id) FROM category));

COMMIT;
