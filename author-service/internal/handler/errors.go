package handler

import (
	"errors"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// toStatusError converts a repository error to a gRPC status error.
//
//	gorm.ErrRecordNotFound -> codes.NotFound -> HTTP 404
//	anything else          -> codes.Internal -> HTTP 500
//
// The real error is logged here but never sent to the client, because it can
// contain SQL or connection details.
func toStatusError(op string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return status.Error(codes.NotFound, "author not found")
	}
	log.Printf("[ERROR] %s: %v", op, err)
	return status.Error(codes.Internal, "internal error")
}
