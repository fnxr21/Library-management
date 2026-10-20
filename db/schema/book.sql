-- Schema for book-service (database: book_mst).
--
-- Mirrors book-service/internal/model/book.go. Runs automatically on the
-- first start of the database volume (see db/README.md).
--
-- author_id and category_id point to rows in OTHER databases (author_mst,
-- category_mst). PostgreSQL cannot enforce foreign keys across databases, so
-- book-service checks them over gRPC instead and copies the names here.

CREATE TABLE IF NOT EXISTS book (
    id           BIGSERIAL PRIMARY KEY,

    -- gorm.Model timestamps; deleted_at enables soft delete
    created_at   TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ,
    deleted_at   TIMESTAMPTZ,

    name         VARCHAR(255),
    description  VARCHAR(255),
    author_id    BIGINT,          -- author.id in author_mst (checked by book-service)
    author       VARCHAR(255),    -- copy of the author's name (Go field: AuthorName)

    -- Stock only changes through one conditional UPDATE (race-safe, see
    -- repository/book.go). The CHECK is a second safety net: the database
    -- itself refuses a negative stock.
    stock        INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT stock_non_negative CHECK (stock >= 0)
);

CREATE INDEX IF NOT EXISTS idx_book_deleted_at ON book (deleted_at);

-- Link table: which categories a book has (many-to-many with another service).
CREATE TABLE IF NOT EXISTS book_category (
    book_id       BIGINT NOT NULL,
    category_id   BIGINT NOT NULL,   -- category.id in category_mst
    categoryname  VARCHAR(255),      -- copy of the category's name

    -- The pair is the key: a book cannot have the same category twice.
    PRIMARY KEY (book_id, category_id),

    -- book_id is in this same database, so here a real foreign key works.
    -- Hard-deleting a book also deletes its links.
    CONSTRAINT fk_book_categories FOREIGN KEY (book_id)
        REFERENCES book (id) ON DELETE CASCADE
);
