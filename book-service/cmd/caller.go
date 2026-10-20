package cmd

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// callerID returns the user ID that the gateway attached to the call.
//
// gRPC metadata works like HTTP headers. After verifying the JWT, the gateway
// sets "x-user-id" so services know who is calling without checking the
// token again. Calls from other services (e.g. book-service checking an
// author) have no user, so this returns "-".
func callerID(ctx context.Context) string {
	md, _ := metadata.FromIncomingContext(ctx)
	if ids := md.Get("x-user-id"); len(ids) == 1 {
		return ids[0]
	}
	return "-"
}
