// Handles the routing of all API routes

package server

import (
	"net/http"
	"time"

	"github.com/CookieG77/AppGDT-Server/internal/handler"
	"github.com/CookieG77/AppGDT-Server/internal/middleware"
)

type Handlers struct {
	Auth  *handler.AuthHandler
	User  *handler.UserHandler
	Space *handler.SpaceHandler
	Note  *handler.NoteHandler
}

func New(addr string, h Handlers, requireAuth func(http.Handler) http.Handler) *http.Server {
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("POST /auth/register", h.Auth.Register)
	mux.HandleFunc("POST /auth/login", h.Auth.Login)

	// Protected routes
	mux.Handle("GET /users/me", requireAuth(http.HandlerFunc(h.User.Me)))
	mux.Handle("GET /spaces", requireAuth(http.HandlerFunc(h.Space.List)))
	mux.Handle("POST /spaces", requireAuth(http.HandlerFunc(h.Space.Create)))
	mux.Handle("GET /spaces/{spaceId}", requireAuth(http.HandlerFunc(h.Space.Get)))
	mux.Handle("PUT /spaces/{spaceId}", requireAuth(http.HandlerFunc(h.Space.Update)))
	mux.Handle("DELETE /spaces/{spaceId}", requireAuth(http.HandlerFunc(h.Space.Delete)))
	mux.Handle("GET /spaces/{spaceId}/notes", requireAuth(http.HandlerFunc(h.Note.List)))
	mux.Handle("POST /spaces/{spaceId}/notes", requireAuth(http.HandlerFunc(h.Note.Create)))
	mux.Handle("GET /notes/{noteId}", requireAuth(http.HandlerFunc(h.Note.Get)))
	mux.Handle("PUT /notes/{noteId}", requireAuth(http.HandlerFunc(h.Note.Update)))
	mux.Handle("DELETE /notes/{noteId}", requireAuth(http.HandlerFunc(h.Note.Delete)))

	return &http.Server{
		Addr:              addr,
		Handler:           middleware.LogRequests(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
