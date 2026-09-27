package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/CookieG77/AppGDT-Server/internal/domain"
	"github.com/CookieG77/AppGDT-Server/internal/httpjson"
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
		httpjson.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Ressource introuvable.")

	case errors.Is(err, domain.ErrEmailUsed):
		httpjson.WriteError(w, http.StatusConflict, "EMAIL_ALREADY_USED", "Cette adresse email est déjà utilisée.")

	default:
		slog.Error("unexpected error", "error", err, "method", r.Method, "path", r.URL.Path)
		httpjson.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Une erreur interne est survenue.")
	}
}
