# Docs

Explanations of how this project works and why it is built this way. Recommended reading order:

1. **[architecture.md](architecture.md)**: the services, how a request travels through them, who owns which data.
2. **[learning-guide.md](learning-guide.md)**: the concepts used in the code (protobuf, gRPC, grpc-gateway, GORM, JWT, race conditions, testing), each with links to the code that uses it. Ends with a step-by-step "add a new endpoint" walkthrough and exercises.
3. **[naming.md](naming.md)**: the naming conventions, and a table of what was renamed.

The [main README](../README.md) covers running the project and the REST API reference.

## Where to start in the code

Read one service end to end before the others. **author-service** is the simplest and has the most comments:

| Step | File | What you learn |
|------|------|----------------|
| 1 | [`author-service/protobuf/author.proto`](../author-service/protobuf/author.proto) | The API contract |
| 2 | [`author-service/main.go`](../author-service/main.go) → [`cmd/server.go`](../author-service/cmd/server.go) | Startup and wiring |
| 3 | [`internal/handler/author.go`](../author-service/internal/handler/author.go) | Validation and RPC logic |
| 4 | [`internal/handler/errors.go`](../author-service/internal/handler/errors.go) | Error → status code mapping |
| 5 | [`internal/repository/author.go`](../author-service/internal/repository/author.go) | Database access |
| 6 | [`internal/model/author.go`](../author-service/internal/model/author.go) | The table |
| 7 | [`internal/handler/author_test.go`](../author-service/internal/handler/author_test.go) | Testing with a fake repository |

Then read:

- **book-service**, for calls between services ([`handler/clients.go`](../book-service/internal/handler/clients.go)) and the race-safe stock update ([`repository/book.go`](../book-service/internal/repository/book.go)).
- **user-service**, for JWT ([`internal/auth`](../user-service/internal/auth/jwt.go)) and the gateway ([`internal/gateway`](../user-service/internal/gateway/gateway.go)).
