# Naming conventions

This project follows the [Go style](https://go.dev/doc/effective_go#names) and the
[protobuf style guide](https://protobuf.dev/programming-guides/style/). When adding code, copy
the patterns below.

## Go

| Thing | Convention | Example |
|---|---|---|
| Package | short, lowercase, no `_` | `handler`, `repository`, `authorpb` |
| Constructor | `New` + type | `NewAuthorRepository`, `NewAuthorHandler` |
| Interface | what it *is* | `AuthorRepository` |
| Private implementation | lowercase version | `authorRepository` |
| Methods inside a typed package | no repeated noun | `repo.GetByID`, not `repo.GetAuthorById` |
| Initialisms | all caps | `ID`, `JWT`, `HTTP`, `GRPC` → `GetByID`, `AuthorID`, `serveGRPC` |
| Receiver | 1–2 letters, the type's initial | `h *AuthorHandler`, `r *authorRepository` |
| Unexported helpers | lowerCamel | `toProto`, `toStatusError`, `loadDotenv` |
| Test fakes | `fake` + role | `fakeRepo`, `fakeAuthors`, `spyUsers` |

The one exception is generated protobuf code, which keeps protoc's naming: the field
`author_id` becomes `AuthorId` (lowercase `d`) in `*.pb.go`. Database models use our own
`AuthorID`.

## Protobuf

| Thing | Convention | Example |
|---|---|---|
| Package | `library.<service>` | `library.author` |
| Go package | `<service>pb` | `authorpb`, `bookpb` |
| Message | PascalCase noun | `Author`, `BookCategory` |
| Field | snake_case (JSON becomes camelCase) | `author_id` → `"authorId"` |
| RPC | Verb + noun (`Create`, `Get`, `List`, `Update`, `Delete`) | `GetAuthor`, `ListBooks` |
| Request/response | `<Rpc>Request` / `<Rpc>Response` | `GetAuthorRequest` |
| REST path | lowercase, `{id}` for path params | `/api/author/{id}` |

Why one request message per RPC? Adding a field to `GetAuthorRequest` later cannot affect
`DeleteAuthor`. Sharing one message (e.g. `Int64Value`) for both would tie them together.

## Files

| Path | Holds |
|---|---|
| `main.go` | only `cmd.Run()` |
| `cmd/server.go` | startup, wiring, shutdown |
| `cmd/caller.go` | reading the caller's identity from metadata |
| `internal/<layer>/<entity>.go` | one file per entity per layer, e.g. `handler/author.go` |
| `internal/handler/errors.go` | error → gRPC status mapping |
| `protobuf/<service>.proto` | the contract; generated code in `protobuf/<service>pb/` |
| `Makefile` | `proto`, `test` (and `clients` / `gateways` / `openapi` where needed) |

## Environment variables

Unchanged on purpose, so existing `.env` files keep working:

| Variable | Meaning |
|---|---|
| `IP_PORT` | user-service public HTTP port (gateway) |
| `GRPC_PORT` | the service's own gRPC port |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` | PostgreSQL connection |
| `AUTHOR_GRPC_ADDR`, `CATEGORY_GRPC_ADDR`, `BOOK_GRPC_ADDR` | `host:port` of services to call |
| `JWT_SECRET`, `JWT_TTL`, `ADMIN_EMAIL`, `ADMIN_PASSWORD` | user-service auth settings |

## What was renamed

| Before | After |
|---|---|
| `protobuf_author`, `protobuf_book`, ... | `authorpb`, `bookpb`, `categorypb`, `userpb` |
| proto package `protobuf_author` | `library.author` |
| `ProtoAuthorRepo.ProtoAuthor` | `Author` |
| `UpdateRequest` (shared name) | `UpdateAuthorRequest`, `UpdateBookRequest`, ... |
| `google.protobuf.Int64Value` as an ID | `GetAuthorRequest{id}`, `DeleteAuthorRequest{id}` |
| `GetAuthorById`, `GetAuthorList` (RPCs) | `GetAuthor`, `ListAuthors` |
| `GetMe` | `GetCurrentUser` (path still `/api/me`) |
| `StockRequest` | `AdjustStockRequest` |
| `RepositoryAuthor(db)` | `NewAuthorRepository(db)` |
| `repository.AuthorService` (interface) | `repository.AuthorRepository` |
| `CreateAuthor`, `GetAuthorById`, ... (repository) | `Create`, `GetByID`, `List`, `Update`, `Delete` |
| `HandlerAuthor(repo)`, `authorServices` | `NewAuthorHandler(repo)`, `AuthorHandler` |
| `ToProtoAuthor` | `toProto` |
| `config.ConnectPostgresGORM()` + global `config.DB` | `db := config.ConnectDatabase()` |
| `MigrationModel` | `migrate` |
| `cmd.RunServer`, `StartRPCServer`, `dotenv` | `cmd.Run`, `serveGRPC`, `loadDotenv` |
| `cmd/main.go` | `cmd/server.go` |
| `makefile` | `Makefile` |
| `Book.AuthorId`, `Book.Author` (model) | `Book.AuthorID`, `Book.AuthorName` (DB columns unchanged) |

### API changes visible to clients

REST paths are the same. JSON changed in these places:

| Before | After |
|---|---|
| `"ID": 1` | `"id": 1` |
| book `"author": "Jane"` | `"authorName": "Jane"` |
| book `"categories": ["Fiction"]`, `"categoryIds": [1]` | `"categories": [{"id": 1, "name": "Fiction"}]` |
| create book with `"categoryIds"` | unchanged |
