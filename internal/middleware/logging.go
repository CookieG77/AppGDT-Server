package middleware

import (
	"crypto/rand"
	"log/slog"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/CookieG77/AppGDT-Server/internal/logging"
)

// requestIDHeader carries the request ID. A valid ID sent by the client (for
// instance the web client forwarding its own ID) is reused, so that a request
// can be followed from the client logs to the server logs. Otherwise a new ID
// is generated. In both cases, it is sent back in the response.
const requestIDHeader = "X-Request-ID"

// validRequestID limits the IDs accepted from the client, so that a forged
// header cannot inject arbitrary content into the logs.
var validRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// LogRequests returns a middleware that assigns an ID to every request,
// makes it available to the logs written while handling it, and writes one
// log line per request once the response is sent.
//
// The level depends on the response: Info for a success, Warn for a client
// error (4xx) and Error for a server error (5xx).
func LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := r.Header.Get(requestIDHeader)
		if !validRequestID.MatchString(requestID) {
			requestID = rand.Text()
		}
		w.Header().Set(requestIDHeader, requestID)

		ctx := logging.WithRequest(r.Context(), requestID, clientIP(r))
		rec := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(rec, r.WithContext(ctx))

		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}

		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		}

		// Only the path is logged, never the query string, the headers or
		// the body, which may contain personal data or secrets.
		slog.Log(ctx, level, "request handled",
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"bytes", rec.bytes,
			"durationMs", float64(time.Since(start).Microseconds())/1000,
			"clientIP", logging.ClientIP(ctx),
		)
	})
}

// clientIP returns the IP address of the client. X-Forwarded-For is
// deliberately ignored: without a trusted reverse proxy in front of the
// server, any client could forge it.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// statusRecorder wraps an http.ResponseWriter to remember the status code
// and the number of bytes written, which are not readable afterwards.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (rec *statusRecorder) WriteHeader(status int) {
	if rec.status == 0 {
		rec.status = status
	}
	rec.ResponseWriter.WriteHeader(status)
}

func (rec *statusRecorder) Write(b []byte) (int, error) {
	if rec.status == 0 {
		rec.status = http.StatusOK
	}
	n, err := rec.ResponseWriter.Write(b)
	rec.bytes += n
	return n, err
}

// Unwrap gives http.ResponseController access to the original writer.
func (rec *statusRecorder) Unwrap() http.ResponseWriter {
	return rec.ResponseWriter
}
