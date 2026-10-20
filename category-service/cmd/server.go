// Package cmd wires the service together and runs it.
//
// category-service speaks gRPC only. Clients never call it directly: REST
// requests enter through user-service (the gateway), which checks the JWT and
// then calls this service over gRPC.
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

	"github.com/fnxr21/category-service/internal/config"
	"github.com/fnxr21/category-service/internal/handler"
	"github.com/fnxr21/category-service/internal/repository"
	"github.com/fnxr21/category-service/protobuf/categorypb"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// Run starts the gRPC server and blocks until Ctrl+C or SIGTERM.
func Run() {
	loadDotenv()
	db := config.ConnectDatabase()

	// Dependency wiring: database -> repository -> handler -> gRPC server.
	repo := repository.NewCategoryRepository(db)
	categoryHandler := handler.NewCategoryHandler(repo)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(logUnary))
	categorypb.RegisterCategoryServiceServer(grpcServer, categoryHandler)

	go serveGRPC(grpcServer, os.Getenv("GRPC_PORT"))

	waitForShutdown()
	// GracefulStop finishes in-flight requests before returning.
	grpcServer.GracefulStop()
}

// loadDotenv loads .env for local runs. In Docker the variables come from
// docker-compose instead, so a missing file is not an error.
func loadDotenv() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("no .env file found, using environment variables")
	}
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

// logUnary is a gRPC interceptor: middleware that runs around every unary
// (single request, single response) call. It logs the method, the caller,
// the status code and how long the call took.
func logUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := next(ctx, req)
	log.Printf("[gRPC] %s user=%s code=%s took=%s", info.FullMethod, callerID(ctx), status.Code(err), time.Since(start))
	return resp, err
}
