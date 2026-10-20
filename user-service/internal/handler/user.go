// Package handler implements the gRPC service defined in user.proto:
// registration, login (issuing JWTs) and "who am I".
package handler

import (
	"context"
	"errors"
	"net/mail"
	"strconv"
	"strings"

	"github.com/fnxr21/user-service/internal/auth"
	"github.com/fnxr21/user-service/internal/model"
	"github.com/fnxr21/user-service/internal/repository"
	"github.com/fnxr21/user-service/protobuf/userpb"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

const (
	minPasswordLen = 8
	maxPasswordLen = 72 // bcrypt ignores bytes after 72
)

// dummyHash is compared against when the email is unknown. bcrypt is slow on
// purpose, so without this an attacker could tell "no such account" (fast)
// from "wrong password" (slow) by timing the response.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)

// UserHandler implements userpb.UserServiceServer.
type UserHandler struct {
	userpb.UnimplementedUserServiceServer
	repo   repository.UserRepository
	tokens *auth.Manager
}

func NewUserHandler(repo repository.UserRepository, tokens *auth.Manager) *UserHandler {
	return &UserHandler{repo: repo, tokens: tokens}
}

// Register creates a member account. Admins are only created by EnsureAdmin.
func (h *UserHandler) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.User, error) {
	username := strings.TrimSpace(req.Username)
	email, err := normalizeEmail(req.Email)
	switch {
	case username == "" || len(username) > 100:
		return nil, status.Error(codes.InvalidArgument, "username must be 1-100 characters")
	case err != nil:
		return nil, status.Error(codes.InvalidArgument, "email is not valid")
	case len(req.Password) < minPasswordLen || len(req.Password) > maxPasswordLen:
		return nil, status.Errorf(codes.InvalidArgument, "password must be %d-%d characters", minPasswordLen, maxPasswordLen)
	}

	user, err := h.createUser(ctx, username, email, req.Password, model.LevelMember)
	if err != nil {
		return nil, err
	}
	return toProto(user), nil
}

// Login checks the password and returns a signed JWT.
//
// Every failure returns the same "invalid email or password" message, so the
// response never reveals which emails have accounts.
func (h *UserHandler) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	invalid := status.Error(codes.Unauthenticated, "invalid email or password")

	email, err := normalizeEmail(req.Email)
	if err != nil {
		return nil, invalid
	}
	user, err := h.repo.GetByEmail(ctx, email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(req.Password))
		return nil, invalid
	}
	if err != nil {
		return nil, toStatusError("Login", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		return nil, invalid
	}

	token, expiresAt, err := h.tokens.Issue(uint64(user.ID), roleOf(user))
	if err != nil {
		return nil, toStatusError("Login", err)
	}
	return &userpb.LoginResponse{
		Token:     token,
		ExpiresAt: timestamppb.New(expiresAt),
		User:      toProto(user),
	}, nil
}

// GetCurrentUser reads the user ID that the gateway put in gRPC metadata
// after verifying the JWT. The handler trusts that value, which is why the
// gRPC port must never be reachable from outside.
func (h *UserHandler) GetCurrentUser(ctx context.Context, _ *emptypb.Empty) (*userpb.User, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	ids := md.Get(auth.MetadataUserID)
	if len(ids) != 1 {
		return nil, status.Error(codes.Unauthenticated, "no authenticated user")
	}
	id, err := strconv.ParseUint(ids[0], 10, 64)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "no authenticated user")
	}

	user, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return nil, toStatusError("GetCurrentUser", err)
	}
	return toProto(user), nil
}

// EnsureAdmin creates an admin account with the given email if none exists
// yet. It runs once at startup (see cmd), so there is always a way to log in
// as admin without editing the database by hand.
func (h *UserHandler) EnsureAdmin(ctx context.Context, email, password string) error {
	email, err := normalizeEmail(email)
	if err != nil {
		return errors.New("admin email is not valid")
	}
	if len(password) < minPasswordLen || len(password) > maxPasswordLen {
		return errors.New("admin password must be 8-72 characters")
	}
	if _, err := h.repo.GetByEmail(ctx, email); err == nil {
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	_, err = h.createUser(ctx, "admin", email, password, model.LevelAdmin)
	return err
}

// createUser hashes the password with bcrypt and stores the user.
// bcrypt adds a random salt and is deliberately slow, which makes stolen
// hashes very expensive to crack.
func (h *UserHandler) createUser(ctx context.Context, username, email, password string, level int16) (model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, toStatusError("createUser", err)
	}
	user, err := h.repo.Create(ctx, model.User{
		Username: username,
		Email:    email,
		Password: string(hash),
		Level:    level,
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return model.User{}, status.Error(codes.AlreadyExists, "email is already registered")
	}
	if err != nil {
		return model.User{}, toStatusError("createUser", err)
	}
	return user, nil
}

// normalizeEmail validates an email and lowercases it, so "A@x.com" and
// "a@x.com" are the same account.
func normalizeEmail(raw string) (string, error) {
	addr, err := mail.ParseAddress(strings.TrimSpace(raw))
	if err != nil || addr.Name != "" {
		return "", errors.New("invalid email")
	}
	return strings.ToLower(addr.Address), nil
}

func roleOf(user model.User) string {
	if user.Level == model.LevelAdmin {
		return auth.RoleAdmin
	}
	return auth.RoleMember
}

// toProto converts the database model to the API message. The password hash
// is deliberately left out.
func toProto(user model.User) *userpb.User {
	return &userpb.User{
		Id:       uint64(user.ID),
		Username: user.Username,
		Email:    user.Email,
		Role:     roleOf(user),
	}
}
