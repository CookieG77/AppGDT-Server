package auth

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestPasswordHasher(t *testing.T) {
	h := NewPasswordHasher(10) // lowest accepted cost, to keep the test fast

	hash, err := h.HashPassword("motdepasse")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if strings.Contains(hash, "motdepasse") || !strings.HasPrefix(hash, "$2") {
		t.Fatalf("the hash must be a bcrypt hash, got %q", hash)
	}
	if err := h.CheckPassword(hash, "motdepasse"); err != nil {
		t.Errorf("the right password must be accepted: %v", err)
	}
	if err := h.CheckPassword(hash, "MotDePasse"); err == nil {
		t.Error("a wrong password must be refused")
	}

	// Two hashes of the same password differ (random salt)
	other, _ := h.HashPassword("motdepasse")
	if other == hash {
		t.Error("each hash must use its own salt")
	}
}

func TestTokenRoundTrip(t *testing.T) {
	m := NewTokenManager("a-secret-of-at-least-32-characters!!", time.Hour)
	token, err := m.Generate(42)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	userID, err := m.Parse(token)
	if err != nil || userID != 42 {
		t.Fatalf("Parse = %d, %v; want 42, nil", userID, err)
	}
}

func TestTokenRejected(t *testing.T) {
	m := NewTokenManager("a-secret-of-at-least-32-characters!!", time.Hour)
	valid, _ := m.Generate(42)

	other := NewTokenManager("another-secret-of-32-characters!!!!!", time.Hour)
	forged, _ := other.Generate(42)

	expired, _ := NewTokenManager("a-secret-of-at-least-32-characters!!", -time.Minute).Generate(42)

	// Token signed with the "none" algorithm: must never be accepted
	none, _ := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{
		Subject:   "42",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)

	// Token without expiration date
	noExp, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: "42"}).
		SignedString([]byte("a-secret-of-at-least-32-characters!!"))

	tests := map[string]string{
		"empty":            "",
		"garbage":          "not.a.token",
		"other secret":     forged,
		"expired":          expired,
		"alg none":         none,
		"no expiration":    noExp,
		"modified payload": valid[:len(valid)-2] + "xx",
	}
	for name, token := range tests {
		if _, err := m.Parse(token); !errors.Is(err, ErrInvalidToken) {
			t.Errorf("%s: expected ErrInvalidToken, got %v", name, err)
		}
	}
}
