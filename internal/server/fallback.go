package server

import (
	"net/http"

	"github.com/CookieG77/AppGDT-Server/internal/httpjson"
)

// withJSONFallback wraps the router so that requests matching no route get
// an error in the API format, like every other error, instead of the plain
// text responses of http.ServeMux:
//   - 404 ROUTE_NOT_FOUND when no route exists for the path;
//   - 405 METHOD_NOT_ALLOWED when the path exists with other methods, along
//     with the Allow header listing them.
func withJSONFallback(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// An empty pattern means that no route matches the request: the
		// returned handler is then the one ServeMux uses to reject it.
		handler, pattern := mux.Handler(r)
		if pattern != "" {
			mux.ServeHTTP(w, r)
			return
		}

		// Run the rejection handler on a recorder, only to learn which
		// status it chose and which methods it allows.
		rec := &headerRecorder{header: http.Header{}}
		handler.ServeHTTP(rec, r)

		switch rec.status {
		case http.StatusMethodNotAllowed:
			w.Header().Set("Allow", rec.header.Get("Allow"))
			httpjson.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "La méthode HTTP n'est pas autorisée pour cette route.")
		case http.StatusNotFound:
			httpjson.WriteError(w, http.StatusNotFound, "ROUTE_NOT_FOUND", "La route demandée n'existe pas.")
		default:
			// Any other rejection (e.g. a redirect) is served as is.
			mux.ServeHTTP(w, r)
		}
	})
}

// headerRecorder is a minimal http.ResponseWriter that keeps the headers and
// the status code, and discards the body.
type headerRecorder struct {
	header http.Header
	status int
}

func (rec *headerRecorder) Header() http.Header { return rec.header }

func (rec *headerRecorder) WriteHeader(status int) {
	if rec.status == 0 {
		rec.status = status
	}
}

func (rec *headerRecorder) Write(b []byte) (int, error) {
	if rec.status == 0 {
		rec.status = http.StatusOK
	}
	return len(b), nil
}
