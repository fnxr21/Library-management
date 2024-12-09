package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/fnxr21/book-service/internal/config"
	"github.com/fnxr21/book-service/internal/handler"
	"github.com/fnxr21/book-service/internal/repository"
	"github.com/fnxr21/book-service/protobuf/protobuf_book"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
)

func RunServer() {
	dotenv()
	config.ConnectPostgresGORM()

	// push RPC server as goroutine
	go StartRPCServer()

	// push gRPC-Gateway generated server as goroutine
	go StartRPCGatewayServer()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan
	log.Println("Termination signal received. Exiting...")

}

func dotenv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}
func StartRPCGatewayServer() {
	gwmux := runtime.NewServeMux()
	IPPORT := os.Getenv("IP_PORT")
	GRPCPORT := os.Getenv("GRPC_PORT")

	err := protobuf_book.RegisterBookServiceHandlerFromEndpoint(context.Background(), gwmux, ":"+GRPCPORT, []grpc.DialOption{grpc.WithInsecure()})
	if err != nil {
		log.Fatal(err)
	}
	gwServer := &http.Server{
		Addr:    ":" + IPPORT,
		Handler: gwmux,
	}

	log.Printf("Serving gRPC-Gateway on http://localhost:%s\n", IPPORT)
	log.Fatalln(gwServer.ListenAndServe())
}

func StartRPCServer() {
	GRPCPORT := os.Getenv("GRPC_PORT")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", GRPCPORT))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	bookRepository := repository.RepositoryBook(config.DB)

	h := handler.HandlerBook(bookRepository)

	protobuf_book.RegisterBookServiceServer(s, h)

	log.Printf("gRPC server listening on port %v\n", GRPCPORT)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
