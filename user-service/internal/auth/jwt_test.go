package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-test-secret-test-secret!"

func newTestManager(t *testing.T) *Manager {
	m, err := NewManager(testSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestIssueAndVerify(t *testing.T) {
	m := newTestManager(t)
	token, _, err := m.Issue(7, RoleAdmin)
	if err != nil {
		t.Fatal(err)
	}

	claims, err := m.Verify(token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if id, _ := claims.UserID(); id != 7 || claims.Role != RoleAdmin {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestNewManagerRejectsShortSecret(t *testing.T) {
	if _, err := NewManager("short", time.Hour); err == nil {
		t.Fatal("want error for short secret")
	}
}

func TestVerifyRejectsExpiredToken(t *testing.T) {
	m := newTestManager(t)
	token, _, _ := m.Issue(1, RoleMember)

	m.now = func() time.Time { return time.Now().Add(2 * time.Hour) }
	if _, err := m.Verify(token); err == nil {
		t.Fatal("want error for expired token")
	}
}

func TestVerifyRejectsOtherSecret(t *testing.T) {
	token, _, _ := newTestManager(t).Issue(1, RoleMember)

	other, _ := NewManager("another-secret-another-secret-another", time.Hour)
	if _, err := other.Verify(token); err == nil {
		t.Fatal("want error for token signed with another secret")
	}
}

func TestVerifyRejectsNoneAlgorithm(t *testing.T) {
	claims := Claims{Role: RoleAdmin, RegisteredClaims: jwt.RegisteredClaims{
		Issuer: issuer, Subject: "1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}}
	token, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := newTestManager(t).Verify(token); err == nil {
		t.Fatal("want error for alg=none token")
	}
}

func TestVerifyRejectsUnknownRole(t *testing.T) {
	token, _, _ := newTestManager(t).Issue(1, "superuser")
	if _, err := newTestManager(t).Verify(token); err == nil {
		t.Fatal("want error for unknown role")
	}
}
