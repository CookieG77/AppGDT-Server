// Package service contains the business logic of the application: it
// validates the input, applies the business rules and orchestrates the
// repositories.
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/CookieG77/AppGDT-Server/internal/auth"
	"github.com/CookieG77/AppGDT-Server/internal/domain"
	"github.com/CookieG77/AppGDT-Server/internal/repository"
)

// Validation limits, identical to those of the API contract.
const (
	emailMaxLength    = 254
	usernameMaxLength = 50
	passwordMinLength = 8
	passwordMaxBytes  = 72 // bcrypt only uses the first 72 bytes of a password
)

// AuthService handles registration, login and access to the current user.
type AuthService struct {
	users  *repository.UserRepository
	hasher *auth.PasswordHasher
	tokens *auth.TokenManager

	// dummyHash is compared against when a login targets an unknown email,
	// so that the response time does not reveal whether an account exists.
	dummyHash string
}

// NewAuthService creates an AuthService. It returns an error if the dummy
// hash used for timing protection cannot be computed.
func NewAuthService(users *repository.UserRepository, hasher *auth.PasswordHasher, tokens *auth.TokenManager) (*AuthService, error) {
	dummyHash, err := hasher.HashPassword("dummy-password-for-timing-protection")
	if err != nil {
		return nil, fmt.Errorf("creating auth service: %w", err)
	}

	return &AuthService{
		users:     users,
		hasher:    hasher,
		tokens:    tokens,
		dummyHash: dummyHash,
	}, nil
}

// Register validates the input, hashes the password and creates the user.
// It returns a *domain.ValidationError if the input is invalid, and
// domain.ErrEmailUsed if the email already belongs to an account.
func (s *AuthService) Register(ctx context.Context, email, username, password string) (domain.User, error) {
	email = normalizeEmail(email)
	username = strings.TrimSpace(username)
	// The password is intentionally not trimmed: spaces may be part of it.

	vErr := &domain.ValidationError{}
	validateEmail(vErr, email)
	validateUsername(vErr, username)
	validatePassword(vErr, password)
	if vErr.HasErrors() {
		return domain.User{}, vErr
	}

	hash, err := s.hasher.HashPassword(password)
	if err != nil {
		return domain.User{}, fmt.Errorf("registering user: %w", err)
	}

	// domain.ErrEmailUsed is wrapped, so errors.Is still recognizes it.
	user, err := s.users.Create(ctx, email, username, hash)
	if err != nil {
		return domain.User{}, fmt.Errorf("registering user: %w", err)
	}

	slog.Info("user registered", "userID", user.ID)
	return user, nil
}

// Login checks the credentials and returns a signed token.
// It returns domain.ErrInvalidCredentials whether the email is unknown or
// the password is wrong, so that the caller cannot tell which one failed.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	email = normalizeEmail(email)

	vErr := &domain.ValidationError{}
	if email == "" {
		vErr.Add("email", "L'email est obligatoire.")
	}
	if password == "" {
		vErr.Add("password", "Le mot de passe est obligatoire.")
	}
	if vErr.HasErrors() {
		return "", vErr
	}

	// No account can have a password longer than bcrypt's limit.
	if len(password) > passwordMaxBytes {
		s.simulatePasswordCheck(password)
		slog.Info("failed login attempt", "reason", "password too long")
		return "", domain.ErrInvalidCredentials
	}

	user, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		s.simulatePasswordCheck(password)
		slog.Info("failed login attempt", "reason", "unknown email")
		return "", domain.ErrInvalidCredentials
	}
	if err != nil {
		return "", fmt.Errorf("logging in: %w", err)
	}

	err = s.hasher.CheckPassword(user.PasswordHash, password)
	if errors.Is(err, auth.ErrPasswordMismatch) {
		slog.Info("failed login attempt", "reason", "wrong password", "userID", user.ID)
		return "", domain.ErrInvalidCredentials
	}
	if err != nil {
		return "", fmt.Errorf("logging in: %w", err)
	}

	token, err := s.tokens.Generate(user.ID)
	if err != nil {
		return "", fmt.Errorf("logging in: %w", err)
	}

	slog.Info("user logged in", "userID", user.ID)
	return token, nil
}

// GetCurrentUser returns the user with the given ID.
// It returns domain.ErrNotFound if the user no longer exists.
func (s *AuthService) GetCurrentUser(ctx context.Context, userID int64) (domain.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return domain.User{}, fmt.Errorf("getting current user: %w", err)
	}
	return user, nil
}

// TokenTTL returns the lifetime of the tokens issued by Login.
func (s *AuthService) TokenTTL() time.Duration {
	return s.tokens.TTL()
}

// simulatePasswordCheck runs a bcrypt comparison against the dummy hash, so
// that a failed login takes as long as a real password check.
func (s *AuthService) simulatePasswordCheck(password string) {
	if len(password) > passwordMaxBytes {
		password = password[:passwordMaxBytes]
	}
	_ = s.hasher.CheckPassword(s.dummyHash, password)
}

// normalizeEmail trims the email and lowercases it, so that the same address
// typed with a different case matches the same account.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validateEmail(vErr *domain.ValidationError, email string) {
	switch {
	case email == "":
		vErr.Add("email", "L'email est obligatoire.")
	case utf8.RuneCountInString(email) > emailMaxLength:
		vErr.Add("email", fmt.Sprintf("L'email ne doit pas dépasser %d caractères.", emailMaxLength))
	case !isValidEmail(email):
		vErr.Add("email", "Le format de l'email est invalide.")
	}
}

// isValidEmail reports whether email is a bare address such as
// "jean@example.com". mail.ParseAddress also accepts forms like
// "Jean <jean@example.com>", which are rejected by the equality check.
func isValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}

func validateUsername(vErr *domain.ValidationError, username string) {
	switch {
	case username == "":
		vErr.Add("username", "Le nom d'utilisateur est obligatoire.")
	case utf8.RuneCountInString(username) > usernameMaxLength:
		vErr.Add("username", fmt.Sprintf("Le nom d'utilisateur ne doit pas dépasser %d caractères.", usernameMaxLength))
	}
}

func validatePassword(vErr *domain.ValidationError, password string) {
	switch {
	case utf8.RuneCountInString(password) < passwordMinLength:
		vErr.Add("password", fmt.Sprintf("Le mot de passe doit contenir au moins %d caractères.", passwordMinLength))
	case len(password) > passwordMaxBytes:
		vErr.Add("password", fmt.Sprintf("Le mot de passe ne doit pas dépasser %d octets.", passwordMaxBytes))
	}
}
