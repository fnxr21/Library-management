-- Schema for user-service (database: user_mst).
--
-- Mirrors user-service/internal/model/user.go. Runs automatically on the
-- first start of the database volume (see db/README.md).
--
-- "user" is a reserved word in PostgreSQL (SELECT user returns the current
-- login), so the table name is always written in double quotes.

CREATE TABLE IF NOT EXISTS "user" (
    id          BIGSERIAL PRIMARY KEY,

    -- gorm.Model timestamps; deleted_at enables soft delete
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ,

    username    VARCHAR(100) NOT NULL,
    email       VARCHAR(255) NOT NULL,
    password    VARCHAR(255) NOT NULL,          -- bcrypt hash, never plain text
    level       SMALLINT NOT NULL DEFAULT 2     -- 1 = admin, 2 = member
);

CREATE INDEX IF NOT EXISTS idx_user_deleted_at ON "user" (deleted_at);

-- One account per email. A UNIQUE index (not a check in Go) is what stops two
-- simultaneous sign-ups with the same email; Go turns the violation into 409.
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_email ON "user" (email);
