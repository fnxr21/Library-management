-- Example data for user-service (database: user_mst).
--
-- Start user-service once first so the table exists. Safe to re-run:
-- an email that is already registered is skipped.
--
-- DEMO ACCOUNTS ONLY. Every account below has the password: password123
-- The password column holds its bcrypt hash, never the plain text (see
-- docs/learning-guide.md section 7). Do not use these accounts or this
-- password anywhere outside a local lab.
--
-- level: 1 = admin, 2 = member
-- "user" is a reserved word in PostgreSQL, so the table name is quoted.
-- IDs are left to the sequence, so these rows never clash with the admin
-- that user-service creates from ADMIN_EMAIL at startup.

INSERT INTO "user" (created_at, updated_at, username, email, password, level) VALUES
  (NOW(), NOW(), 'librarian', 'librarian@example.com', '$2a$10$ZUSYbW7htEj/Wkr0KJGO9Okl2JjZ7.vVxVPSs92w8cqwXNXkjU4ua', 1),
  (NOW(), NOW(), 'budi',      'budi@example.com',      '$2a$10$ZUSYbW7htEj/Wkr0KJGO9Okl2JjZ7.vVxVPSs92w8cqwXNXkjU4ua', 2),
  (NOW(), NOW(), 'siti',      'siti@example.com',      '$2a$10$ZUSYbW7htEj/Wkr0KJGO9Okl2JjZ7.vVxVPSs92w8cqwXNXkjU4ua', 2),
  (NOW(), NOW(), 'andi',      'andi@example.com',      '$2a$10$ZUSYbW7htEj/Wkr0KJGO9Okl2JjZ7.vVxVPSs92w8cqwXNXkjU4ua', 2)
ON CONFLICT (email) DO NOTHING;
