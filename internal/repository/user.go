// Contains the repository for the 'users' table

package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/CookieG77/AppGDT-Server/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository handles the persistence of users.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository using the given connection pool.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create inserts a new user. The email is expected to be already normalized
// and the password already hashed by the caller.
// It returns domain.ErrEmailUsed if the email is already taken.
func (repo *UserRepository) Create(ctx context.Context, email, username, passwordHash string) (domain.User, error) {
	var user domain.User
	err := repo.pool.QueryRow(ctx, `
		INSERT INTO users (email, username, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, email, username, password_hash, created_at, updated_at`,
		email, username, passwordHash,
	).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if isUniqueViolation(err) {
		return domain.User{}, domain.ErrEmailUsed
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("creating user: %w", err)
	}
	return user, nil
}

// GetByEmail returns the user with the given email.
// It returns domain.ErrNotFound if no user matches.
func (repo *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	var user domain.User
	err := repo.pool.QueryRow(ctx, `
		SELECT id, email, username, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("getting user by email: %w", err)
	}
	return user, nil
}

// GetByID returns the user with the given ID.
// It returns domain.ErrNotFound if no user matches.
func (repo *UserRepository) GetByID(ctx context.Context, userID int64) (domain.User, error) {
	var user domain.User
	err := repo.pool.QueryRow(ctx, `
		SELECT id, email, username, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("getting user by id: %w", err)
	}
	return user, nil
}
