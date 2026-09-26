// Handles the routing of all API routes

package server

import (
	"net/http"
	"time"

	"github.com/CookieG77/AppGDT-Server/internal/handler"
)

func New(addr string) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.Health)

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
