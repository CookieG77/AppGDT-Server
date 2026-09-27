package handler

import (
	"log/slog"
	"net/http"

	"github.com/CookieG77/AppGDT-Server/internal/httpjson"
	"github.com/CookieG77/AppGDT-Server/internal/middleware"
	"github.com/CookieG77/AppGDT-Server/internal/service"
)

// UserHandler exposes the routes of the authenticated user.
type UserHandler struct {
	auth *service.AuthService
}

// NewUserHandler creates a UserHandler using the given service.
func NewUserHandler(auth *service.AuthService) *UserHandler {
	return &UserHandler{auth: auth}
}

// Me handles GET /users/me.
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		// Only possible if the route was registered without the middleware
		slog.Error("missing user ID in context", "method", r.Method, "path", r.URL.Path)
		httpjson.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Une erreur interne est survenue.")
		return
	}

	user, err := h.auth.GetCurrentUser(r.Context(), userID)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, user)
}
