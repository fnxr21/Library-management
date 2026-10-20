-- Schema for author-service (database: author_mst).
--
-- Mirrors author-service/internal/model/author.go. docker-compose runs this
-- file automatically the first time the database volume is created, so the
-- table exists before the service starts. The service's GORM AutoMigrate then
-- finds the table already correct and changes nothing.
--
-- If you change the Go model, change this file too (see db/README.md).

CREATE TABLE IF NOT EXISTS author (
    -- BIGSERIAL = bigint + a sequence (author_id_seq) that hands out 1, 2, 3...
    id          BIGSERIAL PRIMARY KEY,

    -- From gorm.Model. deleted_at is NULL for live rows; Delete() sets it
    -- instead of removing the row (soft delete).
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ,

    name        VARCHAR(255),
    biography   VARCHAR(255)
);

-- Every normal query adds "WHERE deleted_at IS NULL"; this index keeps it fast.
-- The name must match what GORM expects (idx_<table>_<column>).
CREATE INDEX IF NOT EXISTS idx_author_deleted_at ON author (deleted_at);
