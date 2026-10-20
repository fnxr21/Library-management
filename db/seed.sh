#!/usr/bin/env sh
# Load the example data in db/seed/ into every database.
#
# Usage (from anywhere, after the stack is running):
#   docker compose up -d --build
#   sh db/seed.sh
#
# The tables are created by the services on startup (GORM AutoMigrate); the
# script waits up to 60s for them, so it can run right after `docker compose up`.
# Usually you run it through `make up` from the repository root.
#
# psql runs inside each database container and uses that container's own
# POSTGRES_USER / POSTGRES_DB, so this script never needs your credentials.
# Re-running is safe: existing rows are skipped.
set -e
cd "$(dirname "$0")/.."

psql_in() { # <compose service>: run psql inside that database container, SQL on stdin
  docker compose exec -T "$1" sh -c 'psql -v ON_ERROR_STOP=1 -q -tA -U "$POSTGRES_USER" -d "$POSTGRES_DB"'
}

# Each service creates its tables a few seconds after `docker compose up`
# returns, so wait for them instead of failing.
wait_for_table() { # <compose service> <table>
  tries=0
  until [ "$(echo "SELECT to_regclass('public.\"$2\"') IS NOT NULL;" | psql_in "$1" 2>/dev/null)" = "t" ]; do
    tries=$((tries + 1))
    if [ "$tries" -ge 60 ]; then
      echo "table $2 in $1 did not appear after 60s. Check: docker compose logs" >&2
      exit 1
    fi
    [ "$tries" -eq 1 ] && echo "waiting for $1 to create table $2..."
    sleep 1
  done
}

seed() { # <compose service> <sql file>
  echo "seeding $1 <- $2"
  if ! docker compose exec -T "$1" sh -c 'psql -v ON_ERROR_STOP=1 -q -o /dev/null -U "$POSTGRES_USER" -d "$POSTGRES_DB"' < "$2"; then
    echo "failed to seed $1. Is the stack running, and has the service created its tables?" >&2
    exit 1
  fi
}

wait_for_table authordb   author
wait_for_table categorydb category
wait_for_table bookdb     book_category # created after book, so both exist
wait_for_table userdb     user

seed authordb   db/seed/author.sql
seed categorydb db/seed/category.sql
seed bookdb     db/seed/book.sql
seed userdb     db/seed/user.sql

echo "done. Demo logins (password: password123): librarian@example.com (admin), budi@example.com, siti@example.com, andi@example.com"
