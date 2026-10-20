// Package cmd wires the service together and runs it.
//
// book-service speaks gRPC only, like author-service (see its cmd package).
// It is also a gRPC *client* of author-service and category-service.
package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fnxr21/book-service/internal/config"
	"github.com/fnxr21/book-service/internal/handler"
	"github.com/fnxr21/book-service/internal/repository"
	"github.com/fnxr21/book-service/protobuf/authorpb"
	"github.com/fnxr21/book-service/protobuf/bookpb"
	"github.com/fnxr21/book-service/protobuf/categorypb"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// Run starts the gRPC server and blocks until Ctrl+C or SIGTERM.
func Run() {
	loadDotenv()
	db := config.ConnectDatabase()

	authorConn := dialGRPC("AUTHOR_GRPC_ADDR")
	defer authorConn.Close()
	categoryConn := dialGRPC("CATEGORY_GRPC_ADDR")
	defer categoryConn.Close()

	// Dependency wiring: database + gRPC clients -> handler -> gRPC server.
	bookHandler := handler.NewBookHandler(
		repository.NewBookRepository(db),
		authorpb.NewAuthorServiceClient(authorConn),
		categorypb.NewCategoryServiceClient(categoryConn),
	)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(logUnary))
	bookpb.RegisterBookServiceServer(grpcServer, bookHandler)

	go serveGRPC(grpcServer, os.Getenv("GRPC_PORT"))

	waitForShutdown()
	grpcServer.GracefulStop()
}

// loadDotenv loads .env for local runs. In Docker the variables come from
// docker-compose instead, so a missing file is not an error.
func loadDotenv() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("no .env file found, using environment variables")
	}
}

// dialGRPC creates a client connection to the address in the given env var.
// grpc.NewClient does not connect yet; it connects on the first call and
// reconnects by itself, so the other service may start later.
//
// insecure = no TLS. Fine inside a private Docker network for a lab; use TLS
// (or mTLS) between services in production.
func dialGRPC(envKey string) *grpc.ClientConn {
	addr := os.Getenv(envKey)
	if addr == "" {
		log.Fatalf("%s is not set", envKey)
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("invalid %s %q: %v", envKey, addr, err)
	}
	return conn
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

// logUnary logs every unary call: method, caller, status code and duration.
func logUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := next(ctx, req)
	log.Printf("[gRPC] %s user=%s code=%s took=%s", info.FullMethod, callerID(ctx), status.Code(err), time.Since(start))
	return resp, err
}
