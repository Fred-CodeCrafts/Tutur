package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
)

// ErrDuplicateEmail is returned when a user with the given email already exists.
var ErrDuplicateEmail = errors.New("duplicate email")

// ErrNotFound is returned when a user is not found.
var ErrNotFound = errors.New("user not found")

// Repository defines the data access interface for auth operations.
type Repository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	UpdateUserRole(ctx context.Context, userID uuid.UUID, role domain.Role) error
}

type repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new auth repository backed by the given pool.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

// CreateUser inserts a new user into the database.
// Returns ErrDuplicateEmail if the email is already taken.
func (r *repository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, name, email, password_hash, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		user.ID,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.IsActive,
	).Scan(&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		// pgx error code 23505 = unique_violation
		if isDuplicateKeyError(err) {
			return ErrDuplicateEmail
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// GetUserByEmail retrieves a user by their email address.
// Returns ErrNotFound if no user exists with that email.
func (r *repository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE email = $1`

	user := &domain.User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

// GetUserByID retrieves a user by their UUID.
// Returns ErrNotFound if no user exists with that ID.
func (r *repository) GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1`

	user := &domain.User{}
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

// UpdateUserRole updates the role of a user identified by userID.
func (r *repository) UpdateUserRole(ctx context.Context, userID uuid.UUID, role domain.Role) error {
	query := `
		UPDATE users
		SET role = $1, updated_at = NOW()
		WHERE id = $2`

	ct, err := r.pool.Exec(ctx, query, role, userID)
	if err != nil {
		return fmt.Errorf("update user role: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// isDuplicateKeyError checks whether the error is a PostgreSQL unique violation.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	// pgconn.PgError has Code field; check via error string as a fallback
	type pgErr interface {
		SQLState() string
	}
	var pe pgErr
	if errors.As(err, &pe) {
		return pe.SQLState() == "23505"
	}
	return false
}
