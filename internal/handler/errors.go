package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/CookieG77/AppGDT-Server/internal/domain"
	"github.com/CookieG77/AppGDT-Server/internal/httpjson"
	"github.com/CookieG77/AppGDT-Server/internal/logging"
)

// writeServiceError translates an error returned by a service into the
// matching HTTP response. Unexpected errors are logged in detail and
// answered with a generic 500, so that no internal detail reaches the client.
func writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var vErr *domain.ValidationError

	switch {
	case errors.As(err, &vErr):
		httpjson.WriteValidationError(w, vErr)

	case errors.Is(err, domain.ErrInvalidCredentials):
		httpjson.WriteError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Email ou mot de passe incorrect.")

	case errors.Is(err, domain.ErrNotFound):
		// A burst of these events from the same user may reveal an attempt
		// to guess the IDs of resources belonging to other users.
		logging.Security(r.Context(), slog.LevelInfo, "resource_not_found", "method", r.Method, "path", r.URL.Path)
		httpjson.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Ressource introuvable.")

	case errors.Is(err, domain.ErrEmailUsed):
		httpjson.WriteError(w, http.StatusConflict, "EMAIL_ALREADY_USED", "Cette adresse email est déjà utilisée.")

	default:
		slog.ErrorContext(r.Context(), "unexpected error", "error", err, "method", r.Method, "path", r.URL.Path)
		httpjson.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Une erreur interne est survenue.")
	}
}
