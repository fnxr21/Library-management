// Package auth issues and verifies JSON Web Tokens (JWT).
//
// A JWT has three base64 parts: header.payload.signature. The payload holds
// "claims" (user ID, role, expiry). Anyone can read the payload, but only
// someone with JWT_SECRET can produce a valid signature, so the server can
// trust the claims without looking anything up in the database.
//
// Flow: Login -> Issue(token) -> client stores it -> client sends
// "Authorization: Bearer <token>" -> gateway calls Verify on every request.
package auth

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	RoleAdmin  = "admin"
	RoleMember = "member"

	issuer = "library-user-service"
	// minSecretLen is the shortest HS256 secret accepted (256 bits).
	minSecretLen = 32
)

// Claims is the JWT payload: the subject is the user ID.
type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// UserID returns the numeric user ID stored in the subject.
func (c *Claims) UserID() (uint64, error) {
	return strconv.ParseUint(c.Subject, 10, 64)
}

// Manager issues and verifies HS256 tokens. HS256 uses one shared secret
// both to sign and to verify, which is fine because only user-service does both.
type Manager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewManager(secret string, ttl time.Duration) (*Manager, error) {
	if len(secret) < minSecretLen {
		return nil, errors.New("JWT secret must be at least 32 bytes")
	}
	if ttl <= 0 {
		return nil, errors.New("JWT TTL must be positive")
	}
	return &Manager{secret: []byte(secret), ttl: ttl, now: time.Now}, nil
}

// Issue signs a token for the user and returns it with its expiry time.
func (m *Manager) Issue(userID uint64, role string) (string, time.Time, error) {
	now := m.now()
	expiresAt := now.Add(m.ttl)
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   strconv.FormatUint(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	return token, expiresAt, err
}

// Verify checks the signature, algorithm, issuer and expiry of a token.
//
// WithValidMethods matters: without it, a token whose header says
// "alg": "none" (no signature at all) could be accepted.
func (m *Manager) Verify(token string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(*jwt.Token) (interface{}, error) {
		return m.secret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(issuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil {
		return nil, err
	}
	if _, err := claims.UserID(); err != nil {
		return nil, errors.New("token subject is not a user ID")
	}
	if claims.Role != RoleAdmin && claims.Role != RoleMember {
		return nil, errors.New("token has an unknown role")
	}
	return claims, nil
}

type claimsKey struct{}

// WithClaims stores verified claims on a request context.
func WithClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, claimsKey{}, c)
}

// ClaimsFrom returns the verified claims on ctx, if any.
func ClaimsFrom(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(claimsKey{}).(*Claims)
	return c, ok
}

// gRPC metadata keys the gateway sets from verified claims for downstream services.
const (
	MetadataUserID   = "x-user-id"
	MetadataUserRole = "x-user-role"
)
