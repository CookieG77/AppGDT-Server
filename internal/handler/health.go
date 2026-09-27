// Package handler provides the basic route to check if you the api is running
package handler

import (
	"log/slog"
	"net/http"
)

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(`{"status": "ok"}`)); err != nil {
		slog.Warn("failed to write health response", "error", err)
	}
}
