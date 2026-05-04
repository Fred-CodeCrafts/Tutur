package language

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
)

// ErrNotFound is returned when a language is not found.
var ErrNotFound = errors.New("language not found")

// ErrDuplicateCode is returned when a language with the given code already exists.
var ErrDuplicateCode = errors.New("language code already exists")

// Repository defines the data access interface for language operations.
type Repository interface {
	ListLanguages(ctx context.Context) ([]domain.Language, error)
	GetLanguageByCode(ctx context.Context, code string) (*domain.Language, error)
	CreateLanguage(ctx context.Context, lang *domain.Language) error
	SetLanguageActive(ctx context.Context, code string, isActive bool) error
}

type repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new language repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

// ListLanguages returns all languages ordered by name.
func (r *repository) ListLanguages(ctx context.Context) ([]domain.Language, error) {
	query := `
		SELECT code, name, region, is_active, created_at
		FROM languages
		ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list languages: %w", err)
	}
	defer rows.Close()

	var langs []domain.Language
	for rows.Next() {
		var l domain.Language
		if err := rows.Scan(&l.Code, &l.Name, &l.Region, &l.IsActive, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan language: %w", err)
		}
		langs = append(langs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate languages: %w", err)
	}

	if langs == nil {
		langs = []domain.Language{}
	}
	return langs, nil
}

// GetLanguageByCode retrieves a single language by its code.
// Returns ErrNotFound if no language exists with that code.
func (r *repository) GetLanguageByCode(ctx context.Context, code string) (*domain.Language, error) {
	query := `
		SELECT code, name, region, is_active, created_at
		FROM languages
		WHERE code = $1`

	var l domain.Language
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&l.Code, &l.Name, &l.Region, &l.IsActive, &l.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get language by code: %w", err)
	}
	return &l, nil
}

// CreateLanguage inserts a new language.
// Returns ErrDuplicateCode if the code is already taken.
func (r *repository) CreateLanguage(ctx context.Context, lang *domain.Language) error {
	query := `
		INSERT INTO languages (code, name, region, is_active, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		lang.Code, lang.Name, lang.Region, lang.IsActive,
	).Scan(&lang.CreatedAt)
	if err != nil {
		if isDuplicateKeyError(err) {
			return ErrDuplicateCode
		}
		return fmt.Errorf("create language: %w", err)
	}
	return nil
}

// SetLanguageActive toggles the is_active flag for a language.
// Returns ErrNotFound if no language exists with that code.
func (r *repository) SetLanguageActive(ctx context.Context, code string, isActive bool) error {
	query := `UPDATE languages SET is_active = $1 WHERE code = $2`

	ct, err := r.pool.Exec(ctx, query, isActive, code)
	if err != nil {
		return fmt.Errorf("set language active: %w", err)
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
	type pgErr interface {
		SQLState() string
	}
	var pe pgErr
	if errors.As(err, &pe) {
		return pe.SQLState() == "23505"
	}
	return false
}
