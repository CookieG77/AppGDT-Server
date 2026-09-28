package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/CookieG77/AppGDT-Server/internal/auth"
	"github.com/CookieG77/AppGDT-Server/internal/domain"
	"github.com/CookieG77/AppGDT-Server/internal/logging"
	"github.com/CookieG77/AppGDT-Server/internal/repository"
)

// AccountService implements the data protection rights of the users
// (GDPR): the right to erasure (deleting the account and all its data) and
// the right to data portability (exporting all the data of the account).
type AccountService struct {
	users  *repository.UserRepository
	spaces *repository.SpaceRepository
	notes  *repository.NoteRepository
	hasher *auth.PasswordHasher
	limits LoginLimiters
}

// NewAccountService creates an AccountService. The login limiters are shared
// with the AuthService, so that the password attempts made to delete an
// account count towards the same limit as the failed logins.
func NewAccountService(users *repository.UserRepository, spaces *repository.SpaceRepository, notes *repository.NoteRepository, hasher *auth.PasswordHasher, limits LoginLimiters) *AccountService {
	return &AccountService{users: users, spaces: spaces, notes: notes, hasher: hasher, limits: limits}
}

// DeleteAccount deletes the account of the given user, along with all their
// spaces and notes. The password is required, so that a stolen token alone
// is not enough to erase an account.
// It returns a *domain.ValidationError if the password is missing,
// domain.ErrWrongPassword if it is wrong, a *domain.TooManyAttemptsError if
// too many password attempts failed recently, and domain.ErrNotFound if the
// account no longer exists.
func (s *AccountService) DeleteAccount(ctx context.Context, userID int64, password string) error {
	if password == "" {
		vErr := &domain.ValidationError{}
		vErr.Add("password", "Le mot de passe est obligatoire.")
		return vErr
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("deleting account: %w", err)
	}

	// Checked before the password, as for the login: otherwise this route
	// could be used to guess the password of an account with a stolen token.
	if allowed, wait := s.limits.ByEmail.Allow(user.Email); !allowed {
		logging.Security(ctx, slog.LevelWarn, "account_deletion_blocked", "retryAfterSeconds", int(wait.Seconds()))
		return &domain.TooManyAttemptsError{RetryAfter: wait}
	}

	if err := s.checkPassword(user.PasswordHash, password); err != nil {
		if errors.Is(err, domain.ErrWrongPassword) {
			if s.limits.ByEmail.Fail(user.Email) {
				logging.Security(ctx, slog.LevelWarn, "login_locked", "scope", "email")
			}
			logging.Security(ctx, slog.LevelWarn, "account_deletion_failed", "reason", "wrong password")
		}
		return err
	}

	if err := s.users.Delete(ctx, userID); err != nil {
		return fmt.Errorf("deleting account: %w", err)
	}

	s.limits.ByEmail.Reset(user.Email)
	logging.Security(ctx, slog.LevelInfo, "account_deleted", "targetUserID", userID)
	return nil
}

// checkPassword returns domain.ErrWrongPassword if the password does not
// match the hash.
func (s *AccountService) checkPassword(hash, password string) error {
	// No account can have a password longer than bcrypt's limit.
	if len(password) > passwordMaxBytes {
		return domain.ErrWrongPassword
	}

	err := s.hasher.CheckPassword(hash, password)
	if errors.Is(err, auth.ErrPasswordMismatch) {
		return domain.ErrWrongPassword
	}
	if err != nil {
		return fmt.Errorf("checking password: %w", err)
	}
	return nil
}

// AccountExport holds all the data stored about a user.
type AccountExport struct {
	ExportedAt time.Time     `json:"exportedAt"`
	User       domain.User   `json:"user"`
	Spaces     []SpaceExport `json:"spaces"`
}

// SpaceExport is a space along with its notes.
type SpaceExport struct {
	domain.Space
	Notes []domain.Note `json:"notes"`
}

// Export returns all the data of the given user: their profile, their
// spaces and the notes of each space. The password hash is never exported.
// It returns domain.ErrNotFound if the account no longer exists.
func (s *AccountService) Export(ctx context.Context, userID int64) (AccountExport, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return AccountExport{}, fmt.Errorf("exporting account: %w", err)
	}

	spaces, err := s.spaces.ListByUser(ctx, userID)
	if err != nil {
		return AccountExport{}, fmt.Errorf("exporting account: %w", err)
	}

	// All the notes are loaded in a single query, then grouped by space.
	notes, err := s.notes.ListByUser(ctx, userID)
	if err != nil {
		return AccountExport{}, fmt.Errorf("exporting account: %w", err)
	}
	notesBySpace := make(map[int64][]domain.Note)
	for _, note := range notes {
		notesBySpace[note.SpaceID] = append(notesBySpace[note.SpaceID], note)
	}

	export := AccountExport{
		ExportedAt: time.Now().UTC(),
		User:       user,
		Spaces:     make([]SpaceExport, 0, len(spaces)),
	}
	for _, space := range spaces {
		spaceNotes := notesBySpace[space.ID]
		if spaceNotes == nil {
			spaceNotes = make([]domain.Note, 0) // Exported as [] rather than null
		}
		export.Spaces = append(export.Spaces, SpaceExport{Space: space, Notes: spaceNotes})
	}

	logging.Security(ctx, slog.LevelInfo, "account_exported", "spaces", len(spaces), "notes", len(notes))
	return export, nil
}
