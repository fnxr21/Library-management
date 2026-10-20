package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fnxr21/user-service/internal/auth"
	"github.com/fnxr21/user-service/protobuf/userpb"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

const testSecret = "test-secret-test-secret-test-secret!"

// spyUsers records the gRPC metadata each call receives.
type spyUsers struct {
	userpb.UnimplementedUserServiceServer
	md metadata.MD
}

func (s *spyUsers) GetCurrentUser(ctx context.Context, _ *emptypb.Empty) (*userpb.User, error) {
	s.md, _ = metadata.FromIncomingContext(ctx)
	return &userpb.User{Id: 1}, nil
}

func (s *spyUsers) Login(ctx context.Context, _ *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	s.md, _ = metadata.FromIncomingContext(ctx)
	return &userpb.LoginResponse{Token: "t"}, nil
}

func setup(t *testing.T) (http.Handler, *spyUsers, *auth.Manager) {
	tokens, err := auth.NewManager(testSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	spy := &spyUsers{}
	mux := NewMux()
	if err := userpb.RegisterUserServiceHandlerServer(context.Background(), mux, spy); err != nil {
		t.Fatal(err)
	}
	// A stand-in for a proxied write route, so role checks can be tested without a backend.
	inner := http.NewServeMux()
	inner.Handle("/api/", mux)
	inner.HandleFunc("/api/book", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusCreated) })
	return RequireJWT(tokens, inner), spy, tokens
}

func do(h http.Handler, method, path, token string, extra map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(`{}`))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	for k, v := range extra {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestPublicRoutesNeedNoToken(t *testing.T) {
	h, _, _ := setup(t)
	if rec := do(h, http.MethodPost, "/api/login", "", nil); rec.Code != http.StatusOK {
		t.Fatalf("login: got %d %s", rec.Code, rec.Body)
	}
}

func TestProtectedRoutesRejectBadTokens(t *testing.T) {
	h, _, _ := setup(t)
	other, _ := auth.NewManager("another-secret-another-secret-another", time.Hour)
	forged, _, _ := other.Issue(1, auth.RoleAdmin)

	for name, token := range map[string]string{"missing": "", "garbage": "not-a-jwt", "wrong secret": forged} {
		t.Run(name, func(t *testing.T) {
			rec := do(h, http.MethodGet, "/api/me", token, nil)
			if rec.Code != http.StatusUnauthorized || !strings.Contains(rec.Body.String(), `"code":16`) {
				t.Fatalf("got %d %s", rec.Code, rec.Body)
			}
		})
	}
}

func TestRolePolicy(t *testing.T) {
	h, _, tokens := setup(t)
	member, _, _ := tokens.Issue(2, auth.RoleMember)
	admin, _, _ := tokens.Issue(1, auth.RoleAdmin)

	if rec := do(h, http.MethodGet, "/api/me", member, nil); rec.Code != http.StatusOK {
		t.Fatalf("member read: got %d", rec.Code)
	}
	if rec := do(h, http.MethodPost, "/api/book", member, nil); rec.Code != http.StatusForbidden {
		t.Fatalf("member write: got %d, want 403", rec.Code)
	}
	if rec := do(h, http.MethodPost, "/api/book", admin, nil); rec.Code != http.StatusCreated {
		t.Fatalf("admin write: got %d", rec.Code)
	}
}

func TestIdentityForwardedAndCannotBeSpoofed(t *testing.T) {
	h, spy, tokens := setup(t)
	member, _, _ := tokens.Issue(2, auth.RoleMember)

	rec := do(h, http.MethodGet, "/api/me", member, map[string]string{
		"Grpc-Metadata-X-User-Id":   "1",
		"Grpc-Metadata-X-User-Role": "admin",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
	if got := spy.md.Get(auth.MetadataUserID); len(got) != 1 || got[0] != "2" {
		t.Fatalf("x-user-id = %v, want [2]", got)
	}
	if got := spy.md.Get(auth.MetadataUserRole); len(got) != 1 || got[0] != auth.RoleMember {
		t.Fatalf("x-user-role = %v, want [member]", got)
	}
	if got := spy.md.Get("authorization"); len(got) != 0 {
		t.Fatal("raw token was forwarded to the backend")
	}
}

func TestPublicRouteIgnoresIdentityHeaders(t *testing.T) {
	h, spy, _ := setup(t)

	do(h, http.MethodPost, "/api/login", "", map[string]string{"Grpc-Metadata-X-User-Id": "1"})
	if got := spy.md.Get(auth.MetadataUserID); len(got) != 0 {
		t.Fatalf("x-user-id = %v on a public route, want none", got)
	}
}
