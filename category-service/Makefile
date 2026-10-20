.PHONY: proto test

# Regenerate Go code from the .proto file. Run after every .proto change,
# then run `make gateways openapi` in user-service so the REST side matches.
proto:
	protoc --go_out=protobuf \
	  --go-grpc_out=protobuf --go-grpc_opt=require_unimplemented_servers=false \
	  protobuf/category.proto

test:
	go test ./...
