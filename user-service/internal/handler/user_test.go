package handler

import (
	"context"
	"testing"
	"time"

	"github.com/fnxr21/user-service/internal/auth"
	"github.com/fnxr21/user-service/internal/model"
	"github.com/fnxr21/user-service/protobuf/userpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

// fakeRepo is an in-memory repository.UserRepository.
type fakeRepo struct {
	rows   map[uint]model.User
	nextID uint
}

func newFakeRepo() *fakeRepo { return &fakeRepo{rows: map[uint]model.User{}, nextID: 1} }

func (f *fakeRepo) Create(_ context.Context, u model.User) (model.User, error) {
	for _, existing := range f.rows {
		if existing.Email == u.Email {
			return model.User{}, gorm.ErrDuplicatedKey
		}
	}
	u.ID = f.nextID
	f.nextID++
	f.rows[u.ID] = u
	return u, nil
}

func (f *fakeRepo) GetByID(_ context.Context, id uint64) (model.User, error) {
	u, ok := f.rows[uint(id)]
	if !ok {
		return model.User{}, gorm.ErrRecordNotFound
	}
	return u, nil
}

func (f *fakeRepo) GetByEmail(_ context.Context, email string) (model.User, error) {
	for _, u := range f.rows {
		if u.Email == email {
			return u, nil
		}
	}
	return model.User{}, gorm.ErrRecordNotFound
}

const testSecret = "test-secret-test-secret-test-secret!"

func newHandler(t *testing.T) (*UserHandler, *auth.Manager) {
	tokens, err := auth.NewManager(testSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return NewUserHandler(newFakeRepo(), tokens), tokens
}

func register(t *testing.T, h *UserHandler, email string) *userpb.User {
	u, err := h.Register(context.Background(), &userpb.RegisterRequest{Username: "reader", Email: email, Password: "correct-horse"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	return u
}

func TestRegisterCreatesMember(t *testing.T) {
	h, _ := newHandler(t)

	u := register(t, h, "Reader@Example.com")
	if u.Role != auth.RoleMember || u.Email != "reader@example.com" {
		t.Fatalf("unexpected user: %+v", u)
	}
	stored, _ := h.repo.GetByID(context.Background(), u.Id)
	if stored.Password == "correct-horse" {
		t.Fatal("password stored in plain text")
	}
}

func TestRegisterValidation(t *testing.T) {
	h, _ := newHandler(t)
	cases := map[string]*userpb.RegisterRequest{
		"no username":    {Email: "a@example.com", Password: "correct-horse"},
		"bad email":      {Username: "a", Email: "not-an-email", Password: "correct-horse"},
		"short password": {Username: "a", Email: "a@example.com", Password: "short"},
	}
	for name, req := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := h.Register(context.Background(), req); status.Code(err) != codes.InvalidArgument {
				t.Fatalf("want InvalidArgument, got %v", err)
			}
		})
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	h, _ := newHandler(t)
	register(t, h, "reader@example.com")

	_, err := h.Register(context.Background(), &userpb.RegisterRequest{Username: "x", Email: "READER@example.com", Password: "correct-horse"})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("want AlreadyExists, got %v", err)
	}
}

func TestLogin(t *testing.T) {
	h, tokens := newHandler(t)
	u := register(t, h, "reader@example.com")

	resp, err := h.Login(context.Background(), &userpb.LoginRequest{Email: "reader@example.com", Password: "correct-horse"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	claims, err := tokens.Verify(resp.Token)
	if err != nil {
		t.Fatalf("issued token does not verify: %v", err)
	}
	if id, _ := claims.UserID(); id != u.Id || claims.Role != auth.RoleMember {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestLoginFailuresLookTheSame(t *testing.T) {
	h, _ := newHandler(t)
	register(t, h, "reader@example.com")

	_, wrongPassword := h.Login(context.Background(), &userpb.LoginRequest{Email: "reader@example.com", Password: "wrong-password"})
	_, unknownEmail := h.Login(context.Background(), &userpb.LoginRequest{Email: "nobody@example.com", Password: "correct-horse"})
	if status.Code(wrongPassword) != codes.Unauthenticated || wrongPassword.Error() != unknownEmail.Error() {
		t.Fatalf("errors differ: %v / %v", wrongPassword, unknownEmail)
	}
}

func TestGetCurrentUserUsesMetadataIdentity(t *testing.T) {
	h, _ := newHandler(t)
	u := register(t, h, "reader@example.com")

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(auth.MetadataUserID, "1"))
	got, err := h.GetCurrentUser(ctx, &emptypb.Empty{})
	if err != nil || got.Id != u.Id {
		t.Fatalf("got %+v, %v", got, err)
	}

	if _, err := h.GetCurrentUser(context.Background(), &emptypb.Empty{}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("no metadata: want Unauthenticated, got %v", err)
	}
}

func TestEnsureAdminIsIdempotent(t *testing.T) {
	h, _ := newHandler(t)

	for i := 0; i < 2; i++ {
		if err := h.EnsureAdmin(context.Background(), "admin@example.com", "admin-password"); err != nil {
			t.Fatalf("EnsureAdmin #%d: %v", i+1, err)
		}
	}
	admin, err := h.repo.GetByEmail(context.Background(), "admin@example.com")
	if err != nil || admin.Level != model.LevelAdmin {
		t.Fatalf("admin = %+v, %v", admin, err)
	}
}
