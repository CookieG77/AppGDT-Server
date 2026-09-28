package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/CookieG77/AppGDT-Server/internal/httpjson"
)

// Recover returns a middleware that turns a panic in a handler into a 500
// response in the API error format, instead of letting net/http close the
// connection without any response. The panic value and the stack trace are
// logged, but never sent to the client.
//
// It must be placed inside LogRequests, so that the request is logged with
// its 500 status and its request ID.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w}

		defer func() {
			v := recover()
			if v == nil {
				return
			}
			// http.ErrAbortHandler is used on purpose to abort a response:
			// net/http handles it silently, so it is passed on unchanged.
			if v == http.ErrAbortHandler {
				panic(v)
			}

			slog.ErrorContext(r.Context(), "panic recovered",
				"panic", fmt.Sprint(v),
				"stack", string(debug.Stack()),
				"method", r.Method,
				"path", r.URL.Path,
			)

			// If the handler already started its response, the status can no
			// longer be changed: the client gets a truncated response.
			if rec.status != 0 {
				return
			}
			httpjson.WriteError(rec, http.StatusInternalServerError, "INTERNAL_ERROR", "Une erreur interne est survenue.")
		}()

		next.ServeHTTP(rec, r)
	})
}
