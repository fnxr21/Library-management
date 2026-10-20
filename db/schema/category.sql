-- Schema for category-service (database: category_mst).
--
-- Mirrors category-service/internal/model/category.go. Runs automatically on
-- the first start of the database volume (see db/README.md).

CREATE TABLE IF NOT EXISTS category (
    id           BIGSERIAL PRIMARY KEY,

    -- gorm.Model timestamps; deleted_at enables soft delete
    created_at   TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ,
    deleted_at   TIMESTAMPTZ,

    name         VARCHAR(255),
    description  VARCHAR(255)
);

CREATE INDEX IF NOT EXISTS idx_category_deleted_at ON category (deleted_at);
