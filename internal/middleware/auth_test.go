package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CookieG77/AppGDT-Server/internal/auth"
)

// fakeUsers says whether a user exists without a database.
type fakeUsers struct {
	existing map[int64]bool
	err      error
}

func (f fakeUsers) Exists(_ context.Context, userID int64) (bool, error) {
	return f.existing[userID], f.err
}

func TestAuthenticate(t *testing.T) {
	tokens := auth.NewTokenManager("a-secret-of-at-least-32-characters!!", time.Hour)
	token7, _ := tokens.Generate(7)
	token8, _ := tokens.Generate(8) // user 8 was deleted

	var gotUserID int64
	protected := Authenticate(tokens, fakeUsers{existing: map[int64]bool{7: true}})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotUserID, _ = UserIDFromContext(r.Context())
			w.WriteHeader(http.StatusNoContent)
		}))

	tests := []struct {
		name   string
		header string
		status int
	}{
		{"valid token", "Bearer " + token7, http.StatusNoContent},
		{"no header", "", http.StatusUnauthorized},
		{"wrong scheme", "Basic " + token7, http.StatusUnauthorized},
		{"invalid token", "Bearer abc", http.StatusUnauthorized},
		{"deleted user", "Bearer " + token8, http.StatusUnauthorized},
	}
	for _, tt := range tests {
		gotUserID = 0
		req := httptest.NewRequest(http.MethodGet, "/spaces", nil)
		if tt.header != "" {
			req.Header.Set("Authorization", tt.header)
		}
		rec := httptest.NewRecorder()
		protected.ServeHTTP(rec, req)

		if rec.Code != tt.status {
			t.Errorf("%s: status = %d, want %d", tt.name, rec.Code, tt.status)
		}
		if tt.status == http.StatusNoContent && gotUserID != 7 {
			t.Errorf("%s: user ID in context = %d, want 7", tt.name, gotUserID)
		}
		if tt.status == http.StatusUnauthorized && gotUserID != 0 {
			t.Errorf("%s: the handler must not be called", tt.name)
		}
	}
}

func TestAuthenticateDatabaseError(t *testing.T) {
	tokens := auth.NewTokenManager("a-secret-of-at-least-32-characters!!", time.Hour)
	token, _ := tokens.Generate(7)
	protected := Authenticate(tokens, fakeUsers{err: errors.New("db down")})(http.NotFoundHandler())

	req := httptest.NewRequest(http.MethodGet, "/spaces", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	protected.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}
