.PHONY: proto gateways openapi test

# Regenerate Go code (and REST handlers) from user.proto.
proto:
	protoc --go_out=protobuf \
	  --go-grpc_out=protobuf --go-grpc_opt=require_unimplemented_servers=false \
	  --grpc-gateway_out=protobuf \
	  protobuf/user.proto

# REST handlers for the internal services, generated from their own .proto
# files. user-service is the only public REST entry point: these handlers turn
# JSON requests into gRPC calls to author, category and book services.
# Run this after changing any of those .proto files.
gateways:
	protoc -I . -I ../author-service --go_out=protobuf \
	  --go-grpc_out=protobuf --go-grpc_opt=require_unimplemented_servers=false \
	  --grpc-gateway_out=protobuf \
	  ../author-service/protobuf/author.proto
	protoc -I . -I ../category-service --go_out=protobuf \
	  --go-grpc_out=protobuf --go-grpc_opt=require_unimplemented_servers=false \
	  --grpc-gateway_out=protobuf \
	  ../category-service/protobuf/category.proto
	protoc -I . -I ../book-service --go_out=protobuf \
	  --go-grpc_out=protobuf --go-grpc_opt=require_unimplemented_servers=false \
	  --grpc-gateway_out=protobuf \
	  ../book-service/protobuf/book.proto

# One OpenAPI spec covering every public endpoint, served at /openapi/library.swagger.json.
openapi:
	protoc -I . -I ../author-service -I ../category-service -I ../book-service \
	  --openapiv2_out=openapi --openapiv2_opt=allow_merge=true,merge_file_name=library \
	  protobuf/user.proto ../author-service/protobuf/author.proto \
	  ../category-service/protobuf/category.proto ../book-service/protobuf/book.proto

test:
	go test ./...
