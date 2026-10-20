-- Example data for author-service (database: author_mst).
--
-- The table is created by the service (GORM AutoMigrate), so start
-- author-service once before running this. Safe to run more than once:
-- rows that already exist are skipped (ON CONFLICT DO NOTHING).
--
-- IDs are fixed so book.sql can refer to them (book.author_id).

BEGIN;

INSERT INTO author (id, created_at, updated_at, name, biography) VALUES
  (1, NOW(), NOW(), 'Pramoedya Ananta Toer', 'Indonesian novelist, author of the Buru Quartet.'),
  (2, NOW(), NOW(), 'Andrea Hirata',         'Indonesian writer known for Laskar Pelangi.'),
  (3, NOW(), NOW(), 'Tere Liye',             'Indonesian author of popular fiction for young readers.'),
  (4, NOW(), NOW(), 'George Orwell',         'English novelist and essayist, author of 1984.'),
  (5, NOW(), NOW(), 'J.R.R. Tolkien',        'English writer, author of The Hobbit and The Lord of the Rings.'),
  (6, NOW(), NOW(), 'Yuval Noah Harari',     'Israeli historian, author of Sapiens.'),
  (7, NOW(), NOW(), 'Robert C. Martin',      'Software engineer, author of Clean Code.'),
  (8, NOW(), NOW(), 'James Clear',           'Writer on habits and decision making.')
ON CONFLICT (id) DO NOTHING;

-- id is a serial column. Inserting explicit IDs does not move its sequence,
-- so without this the next API insert would try id=1 and fail.
SELECT setval(pg_get_serial_sequence('author', 'id'), (SELECT MAX(id) FROM author));

COMMIT;
