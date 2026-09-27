package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/CookieG77/AppGDT-Server/internal/httpjson"
	"github.com/CookieG77/AppGDT-Server/internal/middleware"
)

// currentUserID returns the ID of the authenticated user. If it is missing,
// which only happens when a route is registered without the authentication
// middleware, it writes a 500 response and returns false.
func currentUserID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		slog.Error("missing user ID in context", "method", r.Method, "path", r.URL.Path)
		httpjson.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Une erreur interne est survenue.")
		return 0, false
	}
	return userID, true
}

// pathID reads a positive integer ID from the given path parameter.
// If it is missing or invalid, it writes a 400 response and returns false.
func pathID(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id < 1 {
		httpjson.WriteError(w, http.StatusBadRequest, "INVALID_ID", "L'identifiant doit être un entier positif.")
		return 0, false
	}
	return id, true
}
