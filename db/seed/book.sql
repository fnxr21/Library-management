-- Example data for book-service (database: book_mst).
--
-- Start book-service once first so the tables exist. Safe to re-run.
--
-- Each service has its own database, so there are no foreign keys to
-- author_mst or category_mst. Instead, this file copies the names from
-- author.sql and category.sql, exactly as book-service does when a book is
-- created through the API (see docs/architecture.md, "Data ownership").
-- Keep the IDs and names in sync with those files.

BEGIN;

-- author_id -> author.sql, author = copy of that author's name
INSERT INTO book (id, created_at, updated_at, name, description, author_id, author, stock) VALUES
  ( 1, NOW(), NOW(), 'Bumi Manusia',            'First novel of the Buru Quartet.',               1, 'Pramoedya Ananta Toer', 5),
  ( 2, NOW(), NOW(), 'Anak Semua Bangsa',       'Second novel of the Buru Quartet.',              1, 'Pramoedya Ananta Toer', 3),
  ( 3, NOW(), NOW(), 'Laskar Pelangi',          'Ten children and their school on Belitung.',     2, 'Andrea Hirata',         7),
  ( 4, NOW(), NOW(), 'Sang Pemimpi',            'Sequel to Laskar Pelangi.',                      2, 'Andrea Hirata',         4),
  ( 5, NOW(), NOW(), 'Bumi',                    'First book of the Bumi series.',                 3, 'Tere Liye',             6),
  ( 6, NOW(), NOW(), 'Hujan',                   'A love story after a global disaster.',          3, 'Tere Liye',             2),
  ( 7, NOW(), NOW(), '1984',                    'Life under Big Brother.',                        4, 'George Orwell',         4),
  ( 8, NOW(), NOW(), 'Animal Farm',             'A farm revolution that goes wrong.',             4, 'George Orwell',         3),
  ( 9, NOW(), NOW(), 'The Hobbit',              'Bilbo Baggins and the dragon Smaug.',            5, 'J.R.R. Tolkien',        5),
  (10, NOW(), NOW(), 'Sapiens',                 'A brief history of humankind.',                  6, 'Yuval Noah Harari',     4),
  (11, NOW(), NOW(), 'Clean Code',              'A handbook of agile software craftsmanship.',    7, 'Robert C. Martin',      3),
  (12, NOW(), NOW(), 'Atomic Habits',           'Tiny changes, remarkable results.',              8, 'James Clear',           6),
  -- stock 0: handy for testing that borrowing fails with "insufficient stock"
  (13, NOW(), NOW(), 'The Lord of the Rings',   'The journey to destroy the One Ring.',           5, 'J.R.R. Tolkien',        0),
  -- stock 1: handy for testing concurrent borrows of the last copy
  (14, NOW(), NOW(), 'Rumah Kaca',              'Fourth novel of the Buru Quartet.',              1, 'Pramoedya Ananta Toer', 1)
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('book', 'id'), (SELECT MAX(id) FROM book));

-- category_id -> category.sql, categoryname = copy of that category's name.
-- (book_id, category_id) is the primary key, so duplicates are skipped.
INSERT INTO book_category (book_id, category_id, categoryname) VALUES
  ( 1, 1, 'Fiction'), ( 1, 2, 'Historical'),
  ( 2, 1, 'Fiction'), ( 2, 2, 'Historical'),
  ( 3, 1, 'Fiction'), ( 3, 8, 'Young Adult'),
  ( 4, 1, 'Fiction'), ( 4, 8, 'Young Adult'),
  ( 5, 1, 'Fiction'), ( 5, 3, 'Fantasy'),     ( 5, 8, 'Young Adult'),
  ( 6, 1, 'Fiction'), ( 6, 4, 'Dystopia'),
  ( 7, 1, 'Fiction'), ( 7, 4, 'Dystopia'),
  ( 8, 1, 'Fiction'), ( 8, 4, 'Dystopia'),
  ( 9, 1, 'Fiction'), ( 9, 3, 'Fantasy'),
  (10, 5, 'History'),
  (11, 6, 'Technology'),
  (12, 7, 'Self-Development'),
  (13, 1, 'Fiction'), (13, 3, 'Fantasy'),
  (14, 1, 'Fiction'), (14, 2, 'Historical')
ON CONFLICT DO NOTHING;

COMMIT;
