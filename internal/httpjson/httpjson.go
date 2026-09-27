// Package httpjson provides helpers to read JSON request bodies and write
// JSON responses in the format defined by the API contract.
package httpjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/CookieG77/AppGDT-Server/internal/domain"
)

// maxBodySize limits the size of request bodies (1 MB) to prevent a client
// from exhausting the server memory with a huge payload.
const maxBodySize = 1 << 20

// errorResponse is the error format shared by every route of the API,
// as defined by the Error schema of the OpenAPI contract.
// 'omitempty' for Details allows us to prevent empty details from appearing in the json response
type errorResponse struct {
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Details []domain.FieldError `json:"details,omitempty"`
}

// WriteJSON writes data as a JSON response with the given status code.
// If data cannot be encoded, a 500 error is sent instead.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	body, err := json.Marshal(data)
	if err != nil {
		slog.Error("failed to marshal response", "error", err)
		status = http.StatusInternalServerError
		body = []byte(`{"code":"INTERNAL_ERROR","message":"Une erreur interne est survenue."}`)
	}

	// Fall back to 500 if the given status code is not a known HTTP status
	if http.StatusText(status) == "" {
		status = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(body); err != nil {
		slog.Warn("failed to write response", "error", err)
	}
}

// WriteError writes an error response in the API error format.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	errResp := errorResponse{
		Code:    code,
		Message: message,
	}
	WriteJSON(w, status, errResp)
}

// WriteValidationError writes a 400 response listing the invalid fields.
func WriteValidationError(w http.ResponseWriter, vErr *domain.ValidationError) {
	errResp := errorResponse{
		Code:    "VALIDATION_ERROR",
		Message: "Les données envoyées sont invalides.",
		Details: vErr.Fields,
	}
	WriteJSON(w, http.StatusBadRequest, errResp)
}

// DecodeJSON reads the request body into dst.
// It rejects bodies that are empty, too large, malformed, contain unknown
// fields or contain more than one JSON value.
// The returned error message is safe to send back to the client.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return translateDecodeError(err)
	}

	// The body must contain a single JSON value: anything after it
	// (apart from whitespace) is rejected.
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return translateDecodeError(err)
		}
		return errors.New("Le corps de la requête ne doit contenir qu'un seul objet JSON.")
	}

	return nil
}

// translateDecodeError converts a JSON decoding error into a message that can
// be shown to the client without exposing implementation details.
func translateDecodeError(err error) error {
	var (
		syntaxErr        *json.SyntaxError
		typeErr          *json.UnmarshalTypeError
		maxBytesErr      *http.MaxBytesError
		invalidTargetErr *json.InvalidUnmarshalError
	)

	switch {
	// Empty body
	case errors.Is(err, io.EOF):
		return errors.New("Le corps de la requête ne doit pas être vide.")

	// Body larger than maxBodySize
	case errors.As(err, &maxBytesErr):
		return fmt.Errorf("Le corps de la requête ne doit pas dépasser %d octets.", maxBytesErr.Limit)

	// Invalid JSON syntax, or JSON cut off before its end
	case errors.As(err, &syntaxErr), errors.Is(err, io.ErrUnexpectedEOF):
		return errors.New("Le corps de la requête contient du JSON mal formé.")

	// Wrong type for a field (e.g. a number instead of a string),
	// or a body that is not an object at all (e.g. an array)
	case errors.As(err, &typeErr):
		if typeErr.Field != "" {
			return fmt.Errorf("Le champ %q a un type invalide.", typeErr.Field)
		}
		return errors.New("Le corps de la requête doit être un objet JSON.")

	// Field not declared in the target struct (DisallowUnknownFields).
	// encoding/json has no dedicated error type for this case.
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		field := strings.TrimPrefix(err.Error(), "json: unknown field ")
		return fmt.Errorf("Le champ %s n'est pas autorisé.", field)

	// dst is not a non-nil pointer: this is a programming error, not a
	// client error, so it must be fixed in the code rather than reported.
	case errors.As(err, &invalidTargetErr):
		panic(err)

	// Any other read error (e.g. connection closed by the client)
	default:
		return errors.New("Le corps de la requête est invalide.")
	}
}
