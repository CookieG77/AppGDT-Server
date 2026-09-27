package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// ErrPasswordMismatch is returned when a password does not match its hash.
var ErrPasswordMismatch = errors.New("password does not match")

// PasswordHasher hashes and checks passwords with bcrypt.
type PasswordHasher struct {
	cost int
}

// NewPasswordHasher creates a PasswordHasher using the given bcrypt cost.
func NewPasswordHasher(cost int) *PasswordHasher {
	return &PasswordHasher{
		cost: cost,
	}
}

// HashPassword return a new hash from the given password.
// If the hashing of the password failed it returns an error.
func (h *PasswordHasher) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword returns nil if the given password matchs with the given hash.
// Otherwise, returns an error.
func (h *PasswordHasher) CheckPassword(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrPasswordMismatch
	}
	if err != nil {
		return fmt.Errorf("checking password: %w", err)
	}
	return nil
}
