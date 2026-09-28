package handler

import (
	"log/slog"
	"net/http"
)

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(`{"status": "ok"}`)); err != nil {
		slog.WarnContext(r.Context(), "failed to write health response", "error", err)
	}
}
