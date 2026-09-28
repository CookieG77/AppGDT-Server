// Package handler contains the HTTP layer of the API: each handler reads
// the request (path parameters, JSON body, authenticated user), calls the
// matching service, and translates the result into a JSON response or an
// HTTP error, following the API contract.
package handler

import (
	"net/http"

	"github.com/CookieG77/AppGDT-Server/internal/httpjson"
	"github.com/CookieG77/AppGDT-Server/internal/logging"
	"github.com/CookieG77/AppGDT-Server/internal/service"
)

// AuthHandler exposes the routes for a user to authenticate himself and register an account.
type AuthHandler struct {
	auth *service.AuthService
}

// NewAuthHandler creates a AuthHandler using the given service.
func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := httpjson.DecodeJSON(w, r, &input); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	user, err := h.auth.Register(r.Context(), input.Email, input.Username, input.Password)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	httpjson.WriteJSON(w, http.StatusCreated, user)
}

// authTokenResponse is the login response, as defined by the AuthToken
// schema of the API contract.
type authTokenResponse struct {
	Token     string `json:"token"`
	TokenType string `json:"tokenType"`
	ExpiresIn int    `json:"expiresIn"`
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := httpjson.DecodeJSON(w, r, &input); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	// The client IP set by the request logger is used to limit failed logins.
	clientIP := logging.ClientIP(r.Context())
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}

	token, err := h.auth.Login(r.Context(), input.Email, input.Password, clientIP)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, authTokenResponse{
		Token:     token,
		TokenType: "Bearer",
		ExpiresIn: int(h.auth.TokenTTL().Seconds()),
	})
}
