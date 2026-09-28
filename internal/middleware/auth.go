// Package middleware contains the HTTP middlewares, which run before the
// handlers to apply cross-cutting rules such as authentication.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/CookieG77/AppGDT-Server/internal/auth"
	"github.com/CookieG77/AppGDT-Server/internal/httpjson"
	"github.com/CookieG77/AppGDT-Server/internal/logging"
)

// contextKey is unexported so that no other package can read or overwrite
// the values stored by this middleware in the request context.
type contextKey struct{}

// userIDKey is the key under which the authenticated user ID is stored.
var userIDKey = contextKey{}

// UserChecker is implemented by anything that can tell whether a user still
// exists, such as the user repository.
type UserChecker interface {
	Exists(ctx context.Context, userID int64) (bool, error)
}

// Authenticate returns a middleware that rejects requests without a valid
// Bearer token, and stores the authenticated user ID in the request context.
//
// It also checks that the user still exists: a token stays valid until it
// expires, so without this check, the token of a deleted account would keep
// giving access to the API.
func Authenticate(tokens *auth.TokenManager, users UserChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || tokenString == "" {
				unauthorized(w)
				return
			}

			userID, err := tokens.Parse(tokenString)
			if err != nil {
				logging.Security(r.Context(), slog.LevelWarn, "token_rejected",
					"reason", err.Error(), "method", r.Method, "path", r.URL.Path)
				unauthorized(w)
				return
			}

			exists, err := users.Exists(r.Context(), userID)
			if err != nil {
				slog.ErrorContext(r.Context(), "checking user existence failed", "error", err)
				httpjson.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Une erreur interne est survenue.")
				return
			}
			if !exists {
				logging.Security(r.Context(), slog.LevelWarn, "token_rejected",
					"reason", "user no longer exists", "method", r.Method, "path", r.URL.Path)
				unauthorized(w)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			logging.SetUserID(ctx, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext returns the ID of the authenticated user stored by
// Authenticate. The boolean is false if the route is not protected.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	return userID, ok
}

// unauthorized writes a 401 response in the API error format.
func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	httpjson.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentification requise.")
}
