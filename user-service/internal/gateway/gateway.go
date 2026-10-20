// Package gateway is the single public REST entry point. It verifies the JWT,
// then proxies JSON requests to the internal services over gRPC.
//
// For every request:
//  1. RequireJWT checks the "Authorization: Bearer <token>" header and the
//     caller's role, and stores the verified claims on the request context.
//  2. The grpc-gateway mux matches the URL to an RPC (from google.api.http in
//     the .proto files) and converts the JSON body into a protobuf message.
//  3. identityMetadata copies the verified user ID and role into gRPC metadata.
//  4. The RPC is called; its protobuf reply (or error code) becomes the JSON
//     response (or HTTP status).
package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/fnxr21/user-service/internal/auth"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

// publicRoutes need no token. Everything else does.
var publicRoutes = map[string]bool{
	http.MethodPost + " /api/register": true,
	http.MethodPost + " /api/login":    true,
}

// NewMux returns a grpc-gateway mux that forwards the verified identity, and
// nothing the client could use to fake it, as gRPC metadata.
func NewMux() *runtime.ServeMux {
	return runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(headerMatcher),
		runtime.WithMetadata(identityMetadata),
	)
}

// headerMatcher drops Grpc-Metadata-* headers so clients cannot set x-user-id themselves.
func headerMatcher(key string) (string, bool) {
	if strings.HasPrefix(textproto.CanonicalMIMEHeaderKey(key), "Grpc-Metadata-") {
		return "", false
	}
	return runtime.DefaultHeaderMatcher(key)
}

// identityMetadata passes the caller's verified ID and role to the backend services.
func identityMetadata(ctx context.Context, _ *http.Request) metadata.MD {
	claims, ok := auth.ClaimsFrom(ctx)
	if !ok {
		return nil
	}
	return metadata.Pairs(auth.MetadataUserID, claims.Subject, auth.MetadataUserRole, claims.Role)
}

// RequireJWT rejects requests without a valid bearer token. Members may only
// read (GET); every write needs the admin role.
func RequireJWT(tokens *auth.Manager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if publicRoutes[r.Method+" "+r.URL.Path] {
			r.Header.Del("Authorization")
			next.ServeHTTP(w, r)
			return
		}

		raw, ok := bearerToken(r)
		if !ok {
			writeError(w, http.StatusUnauthorized, codes.Unauthenticated, "missing bearer token")
			return
		}
		claims, err := tokens.Verify(raw)
		if err != nil {
			writeError(w, http.StatusUnauthorized, codes.Unauthenticated, "invalid or expired token")
			return
		}
		if r.Method != http.MethodGet && claims.Role != auth.RoleAdmin {
			writeError(w, http.StatusForbidden, codes.PermissionDenied, "admin role required")
			return
		}

		// grpc-gateway always forwards Authorization; the backends only need the verified identity.
		r.Header.Del("Authorization")
		next.ServeHTTP(w, r.WithContext(auth.WithClaims(r.Context(), claims)))
	})
}

func bearerToken(r *http.Request) (string, bool) {
	scheme, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
		return "", false
	}
	return token, true
}

// writeError uses the same JSON shape as grpc-gateway errors.
func writeError(w http.ResponseWriter, httpStatus int, code codes.Code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	if httpStatus == http.StatusUnauthorized {
		w.Header().Set("WWW-Authenticate", `Bearer realm="library"`)
	}
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    int(code),
		"message": msg,
		"details": []interface{}{},
	})
}

// Healthz reports that the gateway process is up.
func Healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}` + "\n"))
}
