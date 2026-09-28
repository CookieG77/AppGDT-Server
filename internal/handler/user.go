package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/CookieG77/AppGDT-Server/internal/httpjson"
	"github.com/CookieG77/AppGDT-Server/internal/middleware"
	"github.com/CookieG77/AppGDT-Server/internal/service"
)

// UserHandler exposes the routes of the authenticated user.
type UserHandler struct {
	auth    *service.AuthService
	account *service.AccountService
}

// NewUserHandler creates a UserHandler using the given services.
func NewUserHandler(auth *service.AuthService, account *service.AccountService) *UserHandler {
	return &UserHandler{auth: auth, account: account}
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

// deleteAccountInput is the request body of DELETE /users/me.
type deleteAccountInput struct {
	Password string `json:"password"`
}

// DeleteMe handles DELETE /users/me: it deletes the account of the
// authenticated user and all their data, after checking their password.
func (h *UserHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	var input deleteAccountInput
	if err := httpjson.DecodeJSON(w, r, &input); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	if err := h.account.DeleteAccount(r.Context(), userID, input.Password); err != nil {
		writeServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Export handles GET /users/me/export: it returns all the data of the
// authenticated user as a JSON file to download.
func (h *UserHandler) Export(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	export, err := h.account.Export(r.Context(), userID)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	// Asks browsers to download the response as a file rather than display it.
	filename := fmt.Sprintf("gdt-export-%s.json", export.ExportedAt.Format("2006-01-02"))
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	httpjson.WriteJSON(w, http.StatusOK, export)
}
