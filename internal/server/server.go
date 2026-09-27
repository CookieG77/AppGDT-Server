// Handles the routing of all API routes

package server

import (
	"net/http"
	"time"

	"github.com/CookieG77/AppGDT-Server/internal/handler"
)

type Handlers struct {
	Auth *handler.AuthHandler
	User *handler.UserHandler
	// Space *handler.SpaceHandler
	// Note *handler.NoteHandler
}

func New(addr string, h Handlers, requireAuth func(http.Handler) http.Handler) *http.Server {
	mux := http.NewServeMux()



	// Public routes
	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("POST /auth/register", h.Auth.Register)
	mux.HandleFunc("POST /auth/login", h.Auth.Login)

	// Protected routes
	mux.Handle("GET /users/me", requireAuth(http.HandlerFunc(h.User.Me)))

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
