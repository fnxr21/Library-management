# db/

SQL for the four databases. Each service owns one database (see
[docs/architecture.md](../docs/architecture.md)).

```
db/
├── schema/        CREATE TABLE statements, one file per database
│   ├── author.sql      → author_mst
│   ├── category.sql    → category_mst
│   ├── book.sql        → book_mst
│   └── user.sql        → user_mst
├── seed/          example rows (INSERT), same file names
└── seed.sh        loads seed/ into the running databases
```

## Order of things

```
1. docker compose starts a database with an EMPTY volume
      └─> Postgres runs /docker-entrypoint-initdb.d/01-schema.sql   (= db/schema/<x>.sql)
2. the service starts
      └─> GORM AutoMigrate sees the tables are already correct, changes nothing
3. make up / sh db/seed.sh
      └─> inserts the example rows from db/seed/
```

`make up` does all three steps.

## schema/: creating the tables

The files are mounted in `docker-compose.yml`:

```yaml
- ./db/schema/author.sql:/docker-entrypoint-initdb.d/01-schema.sql:ro
```

Postgres runs files in `/docker-entrypoint-initdb.d/` **only when the data volume is empty**,
the very first time. They do **not** run again on later starts, and they do not run on a volume
that already has data. To apply them by hand to a running database (safe: every statement uses
`IF NOT EXISTS`):

```sh
docker compose exec -T authordb sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB"' < db/schema/author.sql
```

To start from scratch and let the schema files run again: `make reset` (**deletes all data**).

### Two definitions of the same tables

The tables are described twice: in the SQL here and in the Go models (`internal/model/*.go`),
which GORM AutoMigrate uses. They were checked to produce an **identical** schema, down to the
index and constraint names. Matching names matter: GORM recognizes its own indexes by name
(`idx_<table>_<column>`), and a different name would make it create a duplicate.

When you change a model, change the SQL file too. To check that they still match:

```sh
# 1. tables made by the SQL files only
make reset && docker compose stop author-service category-service book-service user-service
docker compose exec -T bookdb sh -c 'pg_dump -s -U "$POSTGRES_USER" "$POSTGRES_DB"' > /tmp/files.sql
# 2. let GORM run over them
docker compose start author-service category-service book-service user-service && sleep 5
docker compose exec -T bookdb sh -c 'pg_dump -s -U "$POSTGRES_USER" "$POSTGRES_DB"' > /tmp/after.sql
# 3. no output = GORM changed nothing = the SQL file matches the model
diff /tmp/files.sql /tmp/after.sql
```

### Reading the schema

The files are commented. Things worth noticing:

| Where | What to learn |
|---|---|
| every table | `BIGSERIAL` = bigint + auto-increment sequence; `deleted_at` + index = soft delete |
| `book.sql` | `CHECK (stock >= 0)`: the database refuses negative stock even if the code is wrong |
| `book.sql` | `author_id`, `category_id` have **no** foreign key, because they point to other databases |
| `book.sql` | `book_category` has a composite primary key and a real FK with `ON DELETE CASCADE` |
| `user.sql` | `UNIQUE` index on email is what really prevents duplicate accounts under concurrency |
| `user.sql` | `"user"` must be quoted: `user` alone is a PostgreSQL keyword |

## seed/: example data

See the "Example data" section of the [main README](../README.md#example-data). The rows use
fixed IDs so `book.sql` can refer to authors and categories, and each file resets its ID
sequence afterwards so new rows created through the API continue from there.
