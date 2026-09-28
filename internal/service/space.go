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

// Validation limits for spaces, identical to those of the API contract.
const (
	spaceNameMaxLength        = 100
	spaceDescriptionMaxLength = 1000
)

// SpaceService handles the business rules of spaces. Every method takes the
// ID of the authenticated user, so that a user can only access their own spaces.
type SpaceService struct {
	spaces *repository.SpaceRepository
}

// NewSpaceService creates a SpaceService using the given repository.
func NewSpaceService(spaces *repository.SpaceRepository) *SpaceService {
	return &SpaceService{spaces: spaces}
}

// Create validates the input and creates a space owned by the given user.
// It returns a *domain.ValidationError if the input is invalid.
func (s *SpaceService) Create(ctx context.Context, userID int64, name, description string) (domain.Space, error) {
	name, description, err := validateSpaceInput(name, description)
	if err != nil {
		return domain.Space{}, err
	}

	space, err := s.spaces.Create(ctx, userID, name, description)
	if err != nil {
		return domain.Space{}, fmt.Errorf("creating space: %w", err)
	}
	slog.InfoContext(ctx, "space created", "spaceID", space.ID)
	return space, nil
}

// List returns all the spaces owned by the given user.
func (s *SpaceService) List(ctx context.Context, userID int64) ([]domain.Space, error) {
	spaces, err := s.spaces.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing spaces: %w", err)
	}
	return spaces, nil
}

// Get returns a space owned by the given user.
// It returns domain.ErrNotFound if the space does not exist or belongs to
// another user.
func (s *SpaceService) Get(ctx context.Context, userID, spaceID int64) (domain.Space, error) {
	space, err := s.spaces.GetByID(ctx, userID, spaceID)
	if err != nil {
		return domain.Space{}, fmt.Errorf("getting space: %w", err)
	}
	return space, nil
}

// Update validates the input and replaces the name and description of a
// space owned by the given user.
// It returns a *domain.ValidationError if the input is invalid, and
// domain.ErrNotFound if the space does not exist or belongs to another user.
func (s *SpaceService) Update(ctx context.Context, userID, spaceID int64, name, description string) (domain.Space, error) {
	name, description, err := validateSpaceInput(name, description)
	if err != nil {
		return domain.Space{}, err
	}

	space, err := s.spaces.Update(ctx, userID, spaceID, name, description)
	if err != nil {
		return domain.Space{}, fmt.Errorf("updating space: %w", err)
	}
	slog.InfoContext(ctx, "space updated", "spaceID", space.ID)
	return space, nil
}

// Delete removes a space owned by the given user, along with all its notes.
// It returns domain.ErrNotFound if the space does not exist or belongs to
// another user.
func (s *SpaceService) Delete(ctx context.Context, userID, spaceID int64) error {
	if err := s.spaces.Delete(ctx, userID, spaceID); err != nil {
		return fmt.Errorf("deleting space: %w", err)
	}
	slog.InfoContext(ctx, "space deleted", "spaceID", spaceID)
	return nil
}

// validateSpaceInput trims the name and description, then checks them.
// It returns the cleaned values, or a *domain.ValidationError.
func validateSpaceInput(name, description string) (string, string, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	vErr := &domain.ValidationError{}

	switch {
	case name == "":
		vErr.Add("name", "Le nom est obligatoire.")
	case utf8.RuneCountInString(name) > spaceNameMaxLength:
		vErr.Add("name", fmt.Sprintf("Le nom ne doit pas dépasser %d caractères.", spaceNameMaxLength))
	}

	if utf8.RuneCountInString(description) > spaceDescriptionMaxLength {
		vErr.Add("description", fmt.Sprintf("La description ne doit pas dépasser %d caractères.", spaceDescriptionMaxLength))
	}

	if vErr.HasErrors() {
		return "", "", vErr
	}
	return name, description, nil
}
