package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/CookieG77/AppGDT-Server/internal/httpjson"
)

// healthCheckTimeout bounds the database check, so that the health route
// answers quickly even when the database hangs.
const healthCheckTimeout = 2 * time.Second

// Pinger is implemented by anything that can check its connection, such as
// the database connection pool. Depending on this small interface rather
// than on the pool itself also allows testing the handler with a fake.
type Pinger interface {
	Ping(ctx context.Context) error
}

// HealthHandler exposes the route checking that the API can serve requests.
type HealthHandler struct {
	db Pinger
}

// NewHealthHandler creates a HealthHandler checking the given database.
func NewHealthHandler(db Pinger) *HealthHandler {
	return &HealthHandler{db: db}
}

// Check handles GET /health. It answers 200 if the database responds, and
// 503 otherwise, so that a monitoring tool or an orchestrator can tell that
// the API is running but cannot serve requests.
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), healthCheckTimeout)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		// The cause is logged, but not sent to the client.
		slog.ErrorContext(r.Context(), "health check failed", "error", err)
		httpjson.WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Le service est temporairement indisponible.")
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
