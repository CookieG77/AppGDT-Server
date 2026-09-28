// Package logging enriches the structured logs of the server with the
// context of the HTTP request being handled (request ID, authenticated user),
// and provides a helper to record security events in a uniform way.
//
// Personal data is never logged: no email, password, token or request body.
// Users are identified by their ID only.
package logging

import (
	"context"
	"log/slog"
)

// contextKey is unexported so that no other package can read or overwrite
// the request information stored in the context.
type contextKey struct{}

// requestInfo holds what is known about the request being handled. It is
// stored as a pointer so that a middleware running later in the chain (the
// authentication middleware) can complete it, and the change is seen by the
// request logger that created it.
type requestInfo struct {
	id       string
	clientIP string
	userID   int64
}

// WithRequest returns a copy of ctx carrying the ID and client IP of the
// request being handled.
func WithRequest(ctx context.Context, requestID, clientIP string) context.Context {
	return context.WithValue(ctx, contextKey{}, &requestInfo{id: requestID, clientIP: clientIP})
}

// SetUserID records the authenticated user of the request, so that every
// log written afterward for this request includes it. It does nothing if
// ctx was not created by WithRequest.
func SetUserID(ctx context.Context, userID int64) {
	if info := fromContext(ctx); info != nil {
		info.userID = userID
	}
}

// RequestID returns the ID of the request, or "" if there is none.
func RequestID(ctx context.Context) string {
	if info := fromContext(ctx); info != nil {
		return info.id
	}
	return ""
}

// ClientIP returns the IP address of the client, or "" if unknown.
func ClientIP(ctx context.Context) string {
	if info := fromContext(ctx); info != nil {
		return info.clientIP
	}
	return ""
}

// UserID returns the authenticated user of the request. The boolean is
// false if the request is not authenticated.
func UserID(ctx context.Context) (int64, bool) {
	if info := fromContext(ctx); info != nil && info.userID != 0 {
		return info.userID, true
	}
	return 0, false
}

func fromContext(ctx context.Context) *requestInfo {
	if ctx == nil {
		return nil
	}
	info, _ := ctx.Value(contextKey{}).(*requestInfo)
	return info
}

// Security records a security event (failed login, rejected token...).
// All security events share the same message and an "event" attribute, so
// that they can be filtered easily; the client IP is always included.
func Security(ctx context.Context, level slog.Level, event string, args ...any) {
	attrs := append([]any{"event", event, "clientIP", ClientIP(ctx)}, args...)
	slog.Log(ctx, level, "security event", attrs...)
}

// ContextHandler is a slog.Handler that adds the request ID and the
// authenticated user ID to every record logged with a request context
// (slog.InfoContext, slog.ErrorContext...).
type ContextHandler struct {
	next slog.Handler
}

// NewContextHandler wraps the given handler.
func NewContextHandler(next slog.Handler) *ContextHandler {
	return &ContextHandler{next: next}
}

// Enabled reports whether the wrapped handler handles records at the given level.
func (h *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

// Handle adds the request attributes to the record, then passes it on.
func (h *ContextHandler) Handle(ctx context.Context, record slog.Record) error {
	if id := RequestID(ctx); id != "" {
		record.AddAttrs(slog.String("requestID", id))
	}
	if userID, ok := UserID(ctx); ok {
		record.AddAttrs(slog.Int64("userID", userID))
	}
	return h.next.Handle(ctx, record)
}

// WithAttrs returns a ContextHandler wrapping the handler with the given attributes.
func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{next: h.next.WithAttrs(attrs)}
}

// WithGroup returns a ContextHandler wrapping the handler with the given group.
func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{next: h.next.WithGroup(name)}
}
