# Architecture

## The big picture

The system is four Go services. Each one owns one PostgreSQL database. Clients only ever talk
to **user-service**, which acts as the API gateway. Everything behind it uses gRPC.

```mermaid
flowchart LR
    client([Client<br/>curl / frontend])

    subgraph public[" public "]
        gw["user-service :6006<br/>REST/JSON + JWT check"]
    end

    subgraph internal[" internal Docker network (no published ports) "]
        ugrpc["user-service gRPC :6007"]
        author["author-service :6001"]
        category["category-service :6003"]
        book["book-service :6005"]
    end

    udb[(user_mst)]
    adb[(author_mst)]
    cdb[(category_mst)]
    bdb[(book_mst)]

    client -- "HTTP + Bearer token" --> gw
    gw -- gRPC --> ugrpc
    gw -- gRPC --> author
    gw -- gRPC --> category
    gw -- gRPC --> book
    book -- "gRPC: does author exist?" --> author
    book -- "gRPC: do categories exist?" --> category
    ugrpc --> udb
    author --> adb
    category --> cdb
    book --> bdb
```

| Service          | Owns                       | Talks to                    |
|------------------|----------------------------|-----------------------------|
| user-service     | users, login, JWT, gateway | all services (gateway)      |
| author-service   | authors                    | nothing                     |
| category-service | categories                 | nothing                     |
| book-service     | books, stock, book↔category links | author, category     |

## Why one gateway?

- **One place for security.** The JWT is checked once, at the edge. The internal services don't
  need the secret and don't repeat the check.
- **A small attack surface.** Only port 6006 is reachable. The databases and gRPC ports are not
  published (the database ports are published in `docker-compose.yml` only for local
  debugging).
- **REST outside, gRPC inside.** Browsers and curl speak JSON. Services talk to each other with
  gRPC, which is typed, faster, and generated from the same `.proto` contract. This pattern comes
  from [gRPC with REST and Open APIs](https://grpc.io/blog/coreos/).

## Life of a request

`POST /api/book` with an admin token:

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant G as user-service<br/>gateway
    participant B as book-service
    participant A as author-service
    participant K as category-service
    participant DB as book_mst

    C->>G: POST /api/book {name, authorId, categoryIds}<br/>Authorization: Bearer <JWT>
    G->>G: RequireJWT: verify signature, expiry, role=admin
    G->>G: grpc-gateway: JSON → CreateBookRequest<br/>add metadata x-user-id, x-user-role
    G->>B: gRPC CreateBook
    B->>A: gRPC GetAuthor(id)
    A-->>B: Author{name}
    B->>K: gRPC GetCategory(id) for each category
    K-->>B: Category{name}
    B->>DB: INSERT book + book_category (one transaction)
    B-->>G: Book
    G-->>C: 200 {"id":1,"authorName":"...","categories":[...]}
```

If something fails, the error travels back the same way as a **gRPC status code**, and the
gateway turns it into an **HTTP status**:

| Where it fails | gRPC code | HTTP |
|---|---|---|
| no or invalid token (gateway) | `UNAUTHENTICATED` | 401 |
| member tries to write (gateway) | `PERMISSION_DENIED` | 403 |
| author or category doesn't exist (book-service) | `INVALID_ARGUMENT` | 400 |
| author-service is down (book-service) | `UNAVAILABLE` | 503 |
| book id not found | `NOT_FOUND` | 404 |
| stock would go below zero | `FAILED_PRECONDITION` | 400 |
| database error | `INTERNAL` | 500 (generic message) |

## Who the caller is: gRPC metadata

After verifying the token, the gateway adds two gRPC metadata entries. Metadata works like HTTP
headers for gRPC:

```
x-user-id:   42
x-user-role: member
```

Services read them with `metadata.FromIncomingContext(ctx)`. user-service's `GetCurrentUser`
uses them, and every service logs `user=<id>` per call.

Services **trust** these values without checking a token. That is only safe because:

1. the gateway removes any `Grpc-Metadata-*` header a client sends, so clients cannot set them
2. the gRPC ports are not reachable from outside

If you ever expose a gRPC port publicly, this trust breaks. That is why user-service serves gRPC
on a separate internal port (6007), not on the public one.

## Data ownership: no cross-database JOINs

Each service has its own database, so book-service **cannot** `JOIN author`. Instead it keeps a
copy of what it needs, taken when the book is saved:

| book_mst column | copied from |
|---|---|
| `book.author` (`AuthorName`) | author-service `Author.name` |
| `book_category.categoryname` | category-service `Category.name` |

Trade-off: reads are fast and still work while author-service is down, but if an author is
renamed, old books keep the old name until they are updated. This is a classic microservice
trade-off: **availability over immediate consistency**. A later improvement could push rename
events to book-service.

## Database design

See [`erd.png`](../erd.png) for the full planned ERD. The `CREATE TABLE` statements, with comments,
are in [`db/schema/`](../db/schema/). Implemented tables:

```mermaid
erDiagram
    USER {
        bigint id PK
        varchar username
        varchar email UK
        varchar password "bcrypt hash"
        smallint level "1=admin 2=member"
    }
    AUTHOR {
        bigint id PK
        varchar name
        varchar biography
    }
    CATEGORY {
        bigint id PK
        varchar name
        varchar description
    }
    BOOK {
        bigint id PK
        varchar name
        varchar description
        int author_id "author-service id"
        varchar author "copied name"
        int stock "CHECK >= 0"
    }
    BOOK_CATEGORY {
        bigint book_id PK,FK
        bigint category_id PK "category-service id"
        varchar categoryname "copied name"
    }
    BOOK ||--o{ BOOK_CATEGORY : has
```

Every table except `BOOK_CATEGORY` also has `created_at`, `updated_at` and `deleted_at` from
`gorm.Model` (soft delete). `BOOK_CATEGORY` is a plain link table.
`USER`, `AUTHOR` and `CATEGORY` live in different databases, so the lines between services in
the ERD are logical links, not foreign keys.
