package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/CookieG77/AppGDT-Server/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NoteRepository handles the persistence of notes.
type NoteRepository struct {
	pool *pgxpool.Pool
}

// NewNoteRepository creates a new NoteRepository using the given connection pool.
func NewNoteRepository(pool *pgxpool.Pool) *NoteRepository {
	return &NoteRepository{pool: pool}
}

// Create inserts a new note owned by the given user in the given space with the given details.
// It returns domain.ErrNotFound if no space matches.
func (repo *NoteRepository) Create(ctx context.Context, userID, spaceID int64, title, content string, status domain.NoteStatus) (domain.Note, error) {
	var note domain.Note
	err := repo.pool.QueryRow(ctx, `
		INSERT INTO notes (space_id, title, content, status)
		SELECT s.id, $3::text, $4::text, $5::note_status
		FROM spaces s
		WHERE s.id = $1 AND s.user_id = $2
		RETURNING id, space_id, title, content, status, created_at, updated_at`,
		spaceID, userID, title, content, string(status),
	).Scan(&note.ID, &note.SpaceID, &note.Title, &note.Content, &note.Status, &note.CreatedAt, &note.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Note{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Note{}, fmt.Errorf("creating note: %w", err)
	}
	return note, nil
}

// ListBySpace returns all the notes owned by the given user that are in the given space.
func (repo *NoteRepository) ListBySpace(ctx context.Context, userID, spaceID int64) ([]domain.Note, error) {
	rows, err := repo.pool.Query(ctx, `
		SELECT n.id, n.space_id, n.title, n.content, n.status, n.created_at, n.updated_at
		FROM notes n
		JOIN spaces s ON s.id = n.space_id
		WHERE s.user_id = $1 AND s.id = $2
		ORDER BY n.created_at DESC`,
		userID, spaceID,
	)
	if err != nil {
		return []domain.Note{}, fmt.Errorf("listing notes: %w", err)
	}
	defer rows.Close()

	var notes = make([]domain.Note, 0)
	for rows.Next() {
		var note domain.Note
		if err := rows.Scan(&note.ID, &note.SpaceID, &note.Title, &note.Content, &note.Status, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return []domain.Note{}, fmt.Errorf("scanning notes: %w", err)
		}
		notes = append(notes, note)
	}
	if err := rows.Err(); err != nil {
		return []domain.Note{}, fmt.Errorf("iterating notes: %w", err)
	}
	return notes, nil
}

// ListByUser returns all the notes owned by the given user, in all their spaces.
func (repo *NoteRepository) ListByUser(ctx context.Context, userID int64) ([]domain.Note, error) {
	rows, err := repo.pool.Query(ctx, `
		SELECT n.id, n.space_id, n.title, n.content, n.status, n.created_at, n.updated_at
		FROM notes n
		JOIN spaces s ON s.id = n.space_id
		WHERE s.user_id = $1
		ORDER BY n.created_at DESC`,
		userID,
	)
	if err != nil {
		return []domain.Note{}, fmt.Errorf("listing user notes: %w", err)
	}
	defer rows.Close()

	var notes = make([]domain.Note, 0)
	for rows.Next() {
		var note domain.Note
		if err := rows.Scan(&note.ID, &note.SpaceID, &note.Title, &note.Content, &note.Status, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return []domain.Note{}, fmt.Errorf("scanning user notes: %w", err)
		}
		notes = append(notes, note)
	}
	if err := rows.Err(); err != nil {
		return []domain.Note{}, fmt.Errorf("iterating user notes: %w", err)
	}
	return notes, nil
}

// GetByID returns a note owned by the given user with the given ID.
// It returns domain.ErrNotFound if no note matches.
func (repo *NoteRepository) GetByID(ctx context.Context, userID, noteID int64) (domain.Note, error) {
	var note domain.Note
	err := repo.pool.QueryRow(ctx, `
		SELECT n.id, n.space_id, n.title, n.content, n.status, n.created_at, n.updated_at
		FROM notes n
		JOIN spaces s ON s.id = n.space_id 
		WHERE s.user_id = $1 AND n.id = $2`,
		userID, noteID,
	).Scan(&note.ID, &note.SpaceID, &note.Title, &note.Content, &note.Status, &note.CreatedAt, &note.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Note{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Note{}, fmt.Errorf("getting note by id: %w", err)
	}
	return note, nil
}

// Update returns the updated note owned by the given user with the new settings from the given ones.
// It returns domain.ErrNotFound if no note matches.
func (repo *NoteRepository) Update(ctx context.Context, userID, noteID int64, title, content string, status domain.NoteStatus) (domain.Note, error) {
	var note domain.Note
	err := repo.pool.QueryRow(ctx, `
		UPDATE notes n
		SET title = $3, content = $4, status = $5
		FROM spaces s
		WHERE n.space_id = s.id AND s.user_id = $1 AND n.id = $2
		RETURNING n.id, n.space_id, n.title, n.content, n.status, n.created_at, n.updated_at`,
		userID, noteID, title, content, string(status),
	).Scan(&note.ID, &note.SpaceID, &note.Title, &note.Content, &note.Status, &note.CreatedAt, &note.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Note{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Note{}, fmt.Errorf("updating note: %w", err)
	}
	return note, nil
}

// Delete removes a note if it exists and belongs to the given user.
// It returns domain.ErrNotFound if no note matches.
func (repo *NoteRepository) Delete(ctx context.Context, userID, noteID int64) error {
	tag, err := repo.pool.Exec(ctx, `
		DELETE FROM notes n
		USING spaces s
		WHERE n.space_id = s.id AND s.user_id = $1 AND n.id = $2`,
		userID, noteID,
	)
	if err != nil {
		return fmt.Errorf("deleting note: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
