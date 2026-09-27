// Package domain contains the different errors that can be returned by an API calls and the struct types for the API errors
// As well as the data struct for the users, spaces and notes and their associated fields
package domain

import "errors"

var (
	ErrNotFound           = errors.New("resource not found")
	ErrEmailUsed          = errors.New("email already used")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// FieldError represents a validation failure of a single field
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationError represents all the validation failure of a request
type ValidationError struct {
	Fields []FieldError `json:"fields"`
}

func (e ValidationError) Error() string {
	return "validation failed"
}
