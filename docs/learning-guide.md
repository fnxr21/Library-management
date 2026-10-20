# Learning guide

Each section explains one concept, shows where the project uses it, and lists what to try.
Read [architecture.md](architecture.md) first for the big picture.

- [1. Protocol Buffers and code generation](#1-protocol-buffers-and-code-generation)
- [2. gRPC](#2-grpc)
- [3. grpc-gateway: REST from the same contract](#3-grpc-gateway-rest-from-the-same-contract)
- [4. The layers inside a service](#4-the-layers-inside-a-service)
- [5. GORM and PostgreSQL](#5-gorm-and-postgresql)
- [6. Error handling](#6-error-handling)
- [7. Authentication: bcrypt and JWT](#7-authentication-bcrypt-and-jwt)
- [8. Race conditions and the stock update](#8-race-conditions-and-the-stock-update)
- [9. Testing with fakes](#9-testing-with-fakes)
- [10. Walkthrough: adding a new endpoint](#10-walkthrough-adding-a-new-endpoint)
- [11. Exercises](#11-exercises)

---

## 1. Protocol Buffers and code generation

A `.proto` file is the **contract** of a service: its messages (data) and RPCs (functions).
`protoc` reads it and generates Go code, so the client and server can never disagree about field
names or types.

```proto
message GetAuthorRequest {
    uint64 id = 1;          // "= 1" is the field number used on the wire, not a default value
}

service AuthorService {
    rpc GetAuthor(GetAuthorRequest) returns (Author);
}
```

What `make proto` generates in `protobuf/authorpb/`:

| File | Contains |
|---|---|
| `author.pb.go` | Go structs: `authorpb.Author`, `authorpb.GetAuthorRequest`, ... |
| `author_grpc.pb.go` | `AuthorServiceServer` (interface you implement) and `AuthorServiceClient` (to call it) |
| `author.pb.gw.go` (user-service only) | REST handlers that turn JSON into gRPC calls |

Rules worth remembering:

- **Never edit generated files** (`*.pb.go`). Change the `.proto` and run `make proto`.
- **Never reuse or renumber field numbers** once clients exist; old clients would misread data.
- proto3 has no "unset" for plain fields: an empty string is `""`, a missing number is `0`. That
  is why `UpdateAuthor` treats `""` as "keep the old value".

**Try it:** add `string nationality = 4;` to `Author` in `author.proto`, run `make proto`, and
look at the new `GetNationality()` method in `author.pb.go`.

## 2. gRPC

gRPC calls a function on another server as if it were local. It runs over HTTP/2 and sends
protobuf (binary), not JSON.

**Server side**: implement the generated interface and register it.
See [`author-service/cmd/server.go`](../author-service/cmd/server.go):

```go
grpcServer := grpc.NewServer(grpc.UnaryInterceptor(logUnary))
authorpb.RegisterAuthorServiceServer(grpcServer, authorHandler)
```

**Client side**: dial once, then call methods.
See [`book-service/cmd/server.go`](../book-service/cmd/server.go) and
[`handler/clients.go`](../book-service/internal/handler/clients.go):

```go
conn := dialGRPC("AUTHOR_GRPC_ADDR")             // grpc.NewClient(...)
authors := authorpb.NewAuthorServiceClient(conn)
author, err := authors.GetAuthor(ctx, &authorpb.GetAuthorRequest{Id: 1})
```

Concepts used in this project:

| Concept | What it is | Where |
|---|---|---|
| Unary RPC | one request, one response | `GetAuthor`, `CreateBook`, ... |
| Server streaming | one request, many responses | `ListAuthors` (`returns (stream Author)`) |
| Status codes | gRPC's version of HTTP status (`NotFound`, `InvalidArgument`, ...) | every `handler/errors.go` |
| Metadata | key/value headers on a call | `x-user-id`, see `cmd/caller.go` |
| Interceptor | middleware around every call | `logUnary` in `cmd/server.go` |
| Deadline | max time a call may take | `context.WithTimeout` in `clients.go` |
| Graceful stop | finish running calls before exiting | `grpcServer.GracefulStop()` |

**Try it:** stop author-service and create a book. book-service returns `UNAVAILABLE` (HTTP
503), and `downstreamError` in `clients.go` is why.

## 3. grpc-gateway: REST from the same contract

The `google.api.http` option in a `.proto` maps an RPC to a REST route:

```proto
rpc GetAuthor(GetAuthorRequest) returns (Author) {
    option (google.api.http) = { get: "/api/author/{id}" };   // {id} fills GetAuthorRequest.id
};
```

`protoc-gen-grpc-gateway` generates a handler that:

1. matches `GET /api/author/7`
2. builds `GetAuthorRequest{id: 7}` from the path, query and JSON body
3. calls the gRPC method
4. turns the reply into JSON (`author_id` becomes `authorId`) and the gRPC code into an HTTP status

In this project only **user-service** generates gateway code, for all four services
(`make gateways` in [`user-service/Makefile`](../user-service/Makefile)), and serves it behind the
JWT middleware. `make openapi` generates the API spec from the same files, so the docs can never
drift from the code.

## 4. The layers inside a service

Every service has the same folders. Each layer only calls the one below it:

```
main.go                 calls cmd.Run()
cmd/                    wiring: config → repository → handler → gRPC server
internal/handler/       validate input, apply rules, map errors, model ↔ protobuf
internal/repository/    SQL through GORM; nothing else touches the DB
internal/model/         structs = tables
internal/config/        DB connection + migrations
```

Why bother?

- **Handlers never see SQL**, and repositories never see protobuf. Each layer can change
  without breaking the others.
- **Interfaces make testing easy.** The handler depends on `repository.AuthorRepository` (an
  interface), so tests pass in an in-memory fake (section 9).
- **`internal/`** is enforced by Go: other modules cannot import these packages.

The wiring in `cmd/server.go` is plain **dependency injection**: create the dependency, then pass
it in.

```go
db := config.ConnectDatabase()
repo := repository.NewAuthorRepository(db)
authorHandler := handler.NewAuthorHandler(repo)
```

## 5. GORM and PostgreSQL

| Feature | Where | Note |
|---|---|---|
| `gorm.Model` | all models | adds `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt` |
| Soft delete | `Delete()` | sets `deleted_at`; normal queries skip those rows |
| `AutoMigrate` | `config/postgres.go` | creates missing tables and columns; never drops them |
| `WithContext(ctx)` | every repository method | cancels the query if the client goes away |
| `Preload("Categories")` | `book` repository | loads the has-many relation in a second query |
| Transaction | `bookRepository.Update` | several statements, all or nothing |
| `?` placeholders | `GetByEmail` | values sent separately, so no SQL injection |
| `TranslateError` | user-service config | unique violation → `gorm.ErrDuplicatedKey` |
| `check:` tag | `Book.Stock` | database `CHECK (stock >= 0)` |

**Try it:** delete an author through the API, then run
`SELECT id, deleted_at FROM author;` in the database. The row is still there.

## 6. Error handling

One rule: **log the details, return a safe code.** See
[`author-service/internal/handler/errors.go`](../author-service/internal/handler/errors.go).

```go
if errors.Is(err, gorm.ErrRecordNotFound) {
    return status.Error(codes.NotFound, "author not found")   // client sees 404
}
log.Printf("[ERROR] %s: %v", op, err)                          // you see the real cause
return status.Error(codes.Internal, "internal error")          // client sees a generic 500
```

Database errors can contain SQL, table names or hostnames, so they never go to the client.

## 7. Authentication: bcrypt and JWT

**Passwords** ([`user-service/internal/handler/user.go`](../user-service/internal/handler/user.go)):

- stored as **bcrypt** hashes, which are salted and deliberately slow, so stolen hashes are expensive to crack
- login gives the same error for "no such email" and "wrong password", and even takes the same
  time (`dummyHash`), so attackers cannot discover which emails have accounts

**JWT** ([`user-service/internal/auth/jwt.go`](../user-service/internal/auth/jwt.go)):

```
eyJhbGciOiJIUzI1NiJ9 . eyJzdWIiOiI0MiIsInJvbGUiOiJtZW1iZXIiLCJleHAiOjE3... . 3q2-7w...
      header                        payload (claims)                       signature
```

- The payload is **readable by anyone** (base64), so never put secrets in it.
- The signature proves the payload came from someone holding `JWT_SECRET`. Change one character
  and verification fails.
- `Verify` also checks the algorithm (rejects `alg: none`), issuer, expiry and role.

**Authorization** ([`user-service/internal/gateway/gateway.go`](../user-service/internal/gateway/gateway.go)):
public routes → no token; `GET` → any valid token; everything else → `admin`.

**Try it:** log in and decode the claims locally (avoid pasting tokens into websites):

```sh
echo "$TOKEN" | cut -d. -f2 | base64 -d 2>/dev/null; echo
```

Then change one character of the token and call `/api/me`: you get 401.

## 8. Race conditions and the stock update

The naive version is broken:

```go
book := repo.GetByID(id)        // request A reads stock=1   request B reads stock=1
if book.Stock < 1 { fail }      // A: ok                     B: ok
book.Stock--                    // A: writes 0               B: writes 0
repo.Save(book)                 // two copies lent out, only one existed
```

The fix is in [`book-service/internal/repository/book.go`](../book-service/internal/repository/book.go):
let the database check and update in **one statement**.

```sql
UPDATE book SET stock = stock + $delta
WHERE id = $id AND stock + $delta >= 0
```

PostgreSQL locks the row while it updates, so concurrent requests run one after another. If no
row was updated, the stock was too low. A `CHECK (stock >= 0)` constraint is a second safety
net.

[`repository/book_test.go`](../book-service/internal/repository/book_test.go) proves it: 20
goroutines borrow 5 copies at once, and exactly 5 succeed.

Also note that `UpdateBook` uses `Select("name", "description", ...)`, so updating a book's name
can never overwrite a stock change that happened in between.

## 9. Testing with fakes

Handlers depend on interfaces, so tests replace the real database or service with a small
in-memory version:

| Test | Fake | Proves |
|---|---|---|
| `author_test.go` | `fakeRepo` (map) | validation, 404s, partial update, generic 500 |
| `book_test.go` | `fakeRepo`, `fakeAuthors`, `fakeCategories` | lookups across services, service down → 503 |
| `gateway_test.go` | `spyUsers` gRPC server | 401/403 rules, identity can't be faked, token not forwarded |
| `jwt_test.go` | none | expiry, wrong secret, `alg: none` all rejected |
| `repository/book_test.go` | **real** Postgres (`BOOK_TEST_DSN`) | the race-safe stock update |

A trick in `book_test.go`: `fakeAuthors` embeds `authorpb.AuthorServiceClient`, so it only has to
implement the one method the handler calls.

```sh
cd author-service && go test ./...          # unit tests, no database needed
cd book-service && go test -race ./...      # -race detects data races between goroutines
```

## 10. Walkthrough: adding a new endpoint

Goal: `GET /api/author/{id}/books`, which lists one author's books. It belongs in
**book-service**, which owns books.

1. **Contract.** In `book-service/protobuf/book.proto`, add:
   ```proto
   message ListBooksByAuthorRequest { uint64 author_id = 1; }

   rpc ListBooksByAuthor(ListBooksByAuthorRequest) returns (stream Book) {
       option (google.api.http) = { get: "/api/author/{author_id}/books" };
   };
   ```
2. **Generate.** Run `make proto` in book-service, then `make gateways openapi` in user-service.
3. **Repository.** Add `ListByAuthor(ctx, authorID uint64) ([]model.Book, error)` to the
   `BookRepository` interface and implement it with
   `Preload("Categories").Where("author_id = ?", authorID).Find(&books)`.
4. **Handler.** Add `ListBooksByAuthor` to `BookHandler`, following `ListBooks`. Validate
   `author_id != 0`.
5. **Test.** Add `ListByAuthor` to `fakeRepo` in `book_test.go` (the compiler will insist) and
   write a test.
6. **Permissions.** It is a `GET`, so any logged-in user may call it. No gateway change needed.
7. **Docs.** Add the route to the API table in the main README.

## 11. Exercises

From easy to hard:

1. Add `nationality` to authors (section 1), including the update RPC.
2. Return HTTP 404 instead of 400 when a book's author does not exist. Which function decides that?
3. Add `GET /api/books?q=` search by name (`ILIKE`), with paging (`limit`, `offset`).
4. Block deleting a category that books still use. Which service must ask which?
5. Build borrowing: a `loan` table and `POST /api/book/{id}/borrow` for members. Use
   `x-user-id` for the borrower and `AdjustStock(-1)` for stock. What happens if saving the loan
   fails after the stock was decreased?
6. Cache `GetBook` in Redis and clear the entry when the book changes.
