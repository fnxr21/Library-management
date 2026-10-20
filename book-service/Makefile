.PHONY: proto clients test

# Regenerate Go code from the .proto file. Run after every .proto change,
# then run `make gateways openapi` in user-service so the REST side matches.
proto:
	protoc --go_out=protobuf \
	  --go-grpc_out=protobuf --go-grpc_opt=require_unimplemented_servers=false \
	  protobuf/book.proto

# gRPC clients for the services book-service calls. They are generated from
# the other services' own .proto files, so the contract is never copied.
clients:
	protoc -I . -I ../author-service --go_out=protobuf \
	  --go-grpc_out=protobuf --go-grpc_opt=require_unimplemented_servers=false \
	  ../author-service/protobuf/author.proto
	protoc -I . -I ../category-service --go_out=protobuf \
	  --go-grpc_out=protobuf --go-grpc_opt=require_unimplemented_servers=false \
	  ../category-service/protobuf/category.proto

test:
	go test ./...
