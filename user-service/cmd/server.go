// Package cmd wires user-service together and runs it.
//
// user-service runs two servers:
//
//	:IP_PORT   (6006) public REST/JSON gateway: JWT check -> grpc-gateway -> gRPC
//	:GRPC_PORT (6007) internal gRPC server for UserService
//
// The gateway forwards each REST call as a gRPC call: user routes to this
// process's own gRPC server, author/category/book routes to those services.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fnxr21/user-service/internal/auth"
	"github.com/fnxr21/user-service/internal/config"
	"github.com/fnxr21/user-service/internal/gateway"
	"github.com/fnxr21/user-service/internal/handler"
	"github.com/fnxr21/user-service/internal/repository"
	"github.com/fnxr21/user-service/openapi"
	"github.com/fnxr21/user-service/protobuf/authorpb"
	"github.com/fnxr21/user-service/protobuf/bookpb"
	"github.com/fnxr21/user-service/protobuf/categorypb"
	"github.com/fnxr21/user-service/protobuf/userpb"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// Run starts both servers and blocks until Ctrl+C or SIGTERM.
func Run() {
	loadDotenv()
	db := config.ConnectDatabase()

	tokens, err := auth.NewManager(os.Getenv("JWT_SECRET"), jwtTTL())
	if err != nil {
		log.Fatalf("JWT config: %v", err)
	}

	userHandler := handler.NewUserHandler(repository.NewUserRepository(db), tokens)
	if email := os.Getenv("ADMIN_EMAIL"); email != "" {
		if err := userHandler.EnsureAdmin(context.Background(), email, os.Getenv("ADMIN_PASSWORD")); err != nil {
			log.Fatalf("admin bootstrap: %v", err)
		}
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(logUnary))
	userpb.RegisterUserServiceServer(grpcServer, userHandler)
	go serveGRPC(grpcServer, os.Getenv("GRPC_PORT"))

	httpServer := newGatewayServer(tokens)
	go func() {
		log.Printf("Serving REST gateway on http://localhost%s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("gateway stopped: %v", err)
		}
	}()

	waitForShutdown()

	// Stop accepting new HTTP requests and wait up to 10s for running ones,
	// then stop the gRPC server the same way.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("gateway shutdown: %v", err)
	}
	grpcServer.GracefulStop()
}

// loadDotenv loads .env for local runs. In Docker the variables come from
// docker-compose instead, so a missing file is not an error.
func loadDotenv() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("no .env file found, using environment variables")
	}
}

// jwtTTL reads how long tokens stay valid, e.g. JWT_TTL=24h or 30m.
func jwtTTL() time.Duration {
	raw := os.Getenv("JWT_TTL")
	if raw == "" {
		return 24 * time.Hour
	}
	ttl, err := time.ParseDuration(raw)
	if err != nil {
		log.Fatalf("invalid JWT_TTL %q: %v", raw, err)
	}
	return ttl
}

// newGatewayServer builds the only public HTTP server.
//
// Request flow: http.ServeMux -> gateway.RequireJWT -> grpc-gateway mux ->
// gRPC call to the right service -> JSON response.
func newGatewayServer(tokens *auth.Manager) *http.Server {
	ctx := context.Background()
	gwmux := gateway.NewMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Each Register...HandlerFromEndpoint call is generated code (*.pb.gw.go).
	// It adds the REST routes from that service's .proto and a gRPC client
	// that sends the converted requests to addr.
	endpoints := []struct {
		name     string
		addr     string
		register func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error
	}{
		{"user", "localhost:" + os.Getenv("GRPC_PORT"), userpb.RegisterUserServiceHandlerFromEndpoint},
		{"author", requireEnv("AUTHOR_GRPC_ADDR"), authorpb.RegisterAuthorServiceHandlerFromEndpoint},
		{"category", requireEnv("CATEGORY_GRPC_ADDR"), categorypb.RegisterCategoryServiceHandlerFromEndpoint},
		{"book", requireEnv("BOOK_GRPC_ADDR"), bookpb.RegisterBookServiceHandlerFromEndpoint},
	}
	for _, e := range endpoints {
		if err := e.register(ctx, gwmux, e.addr, opts); err != nil {
			log.Fatalf("register %s gateway: %v", e.name, err)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", gateway.Healthz)
	mux.HandleFunc("/openapi/library.swagger.json", openapi.Handler)
	mux.Handle("/", gateway.RequireJWT(tokens, gwmux))

	return &http.Server{
		Addr:    ":" + os.Getenv("IP_PORT"),
		Handler: mux,
		// Protects against clients that open a connection and send headers
		// very slowly to tie up the server (Slowloris).
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("%s is not set", key)
	}
	return v
}

func serveGRPC(s *grpc.Server, port string) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("gRPC server listening on port %v\n", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func waitForShutdown() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("Termination signal received. Shutting down...")
}

// logUnary logs every unary call with its status code and duration.
func logUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := next(ctx, req)
	log.Printf("[gRPC] %s code=%s took=%s", info.FullMethod, status.Code(err), time.Since(start))
	return resp, err
}
