package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/CookieG77/AppGDT-Server/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SpaceRepository handles the persistence of spaces.
type SpaceRepository struct {
	pool *pgxpool.Pool
}

// NewSpaceRepository creates a new SpaceRepository using the given connection pool.
func NewSpaceRepository(pool *pgxpool.Pool) *SpaceRepository {
	return &SpaceRepository{pool: pool}
}

// Create inserts a new space owned by the given user with the given details.
func (repo *SpaceRepository) Create(ctx context.Context, userID int64, name, description string) (domain.Space, error) {
	var space domain.Space
	err := repo.pool.QueryRow(ctx, `
		INSERT INTO spaces (user_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, name, description, created_at, updated_at`,
		userID, name, description,
	).Scan(&space.ID, &space.UserID, &space.Name, &space.Description, &space.CreatedAt, &space.UpdatedAt)
	if err != nil {
		return domain.Space{}, fmt.Errorf("create space: %w", err)
	}
	return space, nil
}

// ListByUser returns all the spaces owned by a given user.
func (repo *SpaceRepository) ListByUser(ctx context.Context, userID int64) ([]domain.Space, error) {
	rows, err := repo.pool.Query(ctx, `
		SELECT id, user_id, name, description, created_at, updated_at
		FROM spaces
		WHERE user_id = $1
		ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return []domain.Space{}, fmt.Errorf("listing spaces: %w", err)
	}
	defer rows.Close()

	var spaces = make([]domain.Space, 0)
	for rows.Next() {
		var space domain.Space
		if err := rows.Scan(&space.ID, &space.UserID, &space.Name, &space.Description, &space.CreatedAt, &space.UpdatedAt); err != nil {
			return []domain.Space{}, fmt.Errorf("scanning spaces: %w", err)
		}
		spaces = append(spaces, space)
	}
	if err := rows.Err(); err != nil {
		return []domain.Space{}, fmt.Errorf("iterating spaces: %w", err)
	}
	return spaces, nil
}

// GetByID returns a space with the given ID if it belongs to the given user.
// It returns domain.ErrNotFound if no space matches.
func (repo *SpaceRepository) GetByID(ctx context.Context, userID, spaceID int64) (domain.Space, error) {
	var space domain.Space
	err := repo.pool.QueryRow(ctx, `
		SELECT id, user_id, name, description, created_at, updated_at
		FROM spaces
		WHERE id = $1 AND user_id = $2`,
		spaceID, userID,
	).Scan(&space.ID, &space.UserID, &space.Name, &space.Description, &space.CreatedAt, &space.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Space{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Space{}, fmt.Errorf("getting a space by id: %w", err)
	}
	return space, nil
}

// Update returns the updated space with the new settings from the given ones.
// It returns domain.ErrNotFound if no space matches.
func (repo *SpaceRepository) Update(ctx context.Context, userID, spaceID int64, name, description string) (domain.Space, error) {
	var space domain.Space
	err := repo.pool.QueryRow(ctx, `
		UPDATE spaces
		SET name = $1, description = $2
		WHERE id = $3 AND user_id = $4
		RETURNING id, user_id, name, description, created_at, updated_at`,
		name, description, spaceID, userID,
	).Scan(&space.ID, &space.UserID, &space.Name, &space.Description, &space.CreatedAt, &space.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Space{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Space{}, fmt.Errorf("updating space: %w", err)
	}
	return space, nil
}

// Delete removes a space (and its notes) if it exists and belongs to the given user.
// It returns domain.ErrNotFound if no space matches.
func (repo *SpaceRepository) Delete(ctx context.Context, userID, spaceID int64) error {
	tag, err := repo.pool.Exec(ctx, `
		DELETE FROM spaces WHERE id = $1 AND user_id = $2`,
		spaceID, userID,
	)
	if err != nil {
		return fmt.Errorf("deleting space: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
