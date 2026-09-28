package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/CookieG77/AppGDT-Server/internal/domain"
	"github.com/CookieG77/AppGDT-Server/internal/repository"
)

// NoteService handles the business rules of notes. Every method takes the
// ID of the authenticated user, so that a user can only access the notes
// stored in their own spaces.
type NoteService struct {
	notes  *repository.NoteRepository
	spaces *repository.SpaceRepository
}

// NewNoteService creates a NoteService using the given repositories.
func NewNoteService(notes *repository.NoteRepository, spaces *repository.SpaceRepository) *NoteService {
	return &NoteService{notes: notes, spaces: spaces}
}

// NoteInput holds the fields a client can set on a note. Status is a
// pointer so that an omitted status (nil) can be told apart from an empty
// one: the first defaults to domain.Todo, the second is invalid.
type NoteInput struct {
	Title   string
	Content string
	Status  *domain.NoteStatus
}

// Create validates the input and creates a note in a space owned by the
// given user.
// It returns a *domain.ValidationError if the input is invalid, and
// domain.ErrNotFound if the space does not exist or belongs to another user.
func (s *NoteService) Create(ctx context.Context, userID, spaceID int64, input NoteInput) (domain.Note, error) {
	title, content, status, err := validateNoteInput(input)
	if err != nil {
		return domain.Note{}, err
	}

	note, err := s.notes.Create(ctx, userID, spaceID, title, content, status)
	if err != nil {
		return domain.Note{}, fmt.Errorf("creating note: %w", err)
	}
	slog.InfoContext(ctx, "note created", "noteID", note.ID, "spaceID", note.SpaceID)
	return note, nil
}

// List returns all the notes of a space owned by the given user.
// It returns domain.ErrNotFound if the space does not exist or belongs to
// another user, so that an empty list always means an existing, empty space.
func (s *NoteService) List(ctx context.Context, userID, spaceID int64) ([]domain.Note, error) {
	if _, err := s.spaces.GetByID(ctx, userID, spaceID); err != nil {
		return nil, fmt.Errorf("checking space: %w", err)
	}

	notes, err := s.notes.ListBySpace(ctx, userID, spaceID)
	if err != nil {
		return nil, fmt.Errorf("listing notes: %w", err)
	}
	return notes, nil
}

// Get returns a note stored in a space owned by the given user.
// It returns domain.ErrNotFound if the note does not exist or belongs to
// another user.
func (s *NoteService) Get(ctx context.Context, userID, noteID int64) (domain.Note, error) {
	note, err := s.notes.GetByID(ctx, userID, noteID)
	if err != nil {
		return domain.Note{}, fmt.Errorf("getting note: %w", err)
	}
	return note, nil
}

// Update validates the input and replaces the title, content and status of
// a note stored in a space owned by the given user. An omitted status is
// reset to domain.Todo, as the whole note is replaced.
// It returns a *domain.ValidationError if the input is invalid, and
// domain.ErrNotFound if the note does not exist or belongs to another user.
func (s *NoteService) Update(ctx context.Context, userID, noteID int64, input NoteInput) (domain.Note, error) {
	title, content, status, err := validateNoteInput(input)
	if err != nil {
		return domain.Note{}, err
	}

	note, err := s.notes.Update(ctx, userID, noteID, title, content, status)
	if err != nil {
		return domain.Note{}, fmt.Errorf("updating note: %w", err)
	}
	slog.InfoContext(ctx, "note updated", "noteID", note.ID)
	return note, nil
}

// Delete removes a note stored in a space owned by the given user.
// It returns domain.ErrNotFound if the note does not exist or belongs to
// another user.
func (s *NoteService) Delete(ctx context.Context, userID, noteID int64) error {
	if err := s.notes.Delete(ctx, userID, noteID); err != nil {
		return fmt.Errorf("deleting note: %w", err)
	}
	slog.InfoContext(ctx, "note deleted", "noteID", noteID)
	return nil
}

// validateNoteInput trims the title, then checks every field. The content is
// kept as is, since its leading and trailing whitespace can be meaningful
// (indentation, blank lines).
// It returns the cleaned values, or a *domain.ValidationError.
func validateNoteInput(input NoteInput) (string, string, domain.NoteStatus, error) {
	title := strings.TrimSpace(input.Title)
	content := input.Content
	status := domain.Todo
	if input.Status != nil {
		status = *input.Status
	}

	vErr := &domain.ValidationError{}

	switch {
	case title == "":
		vErr.Add("title", "Le titre est obligatoire.")
	case utf8.RuneCountInString(title) > domain.NoteTitleMaxLength:
		vErr.Add("title", fmt.Sprintf("Le titre ne doit pas dépasser %d caractères.", domain.NoteTitleMaxLength))
	}

	if utf8.RuneCountInString(content) > domain.NoteContentMaxLength {
		vErr.Add("content", fmt.Sprintf("Le contenu ne doit pas dépasser %d caractères.", domain.NoteContentMaxLength))
	}

	if !status.IsValid() {
		vErr.Add("status", "L'état doit valoir todo, in_progress ou done.")
	}

	if vErr.HasErrors() {
		return "", "", "", vErr
	}
	return title, content, status, nil
}
