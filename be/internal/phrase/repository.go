package phrase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
)

// ErrNotFound is returned when a phrase is not found.
var ErrNotFound = errors.New("phrase not found")

// Repository defines the data access interface for phrase operations.
type Repository interface {
	CreatePhrase(ctx context.Context, p *domain.Phrase) error
	GetPhraseByID(ctx context.Context, id uuid.UUID) (*domain.Phrase, error)
	ListPendingPhrases(ctx context.Context) ([]domain.Phrase, error)
	ListPhrasesByContributor(ctx context.Context, contributorID uuid.UUID) ([]domain.Phrase, error)
	IsLanguageActive(ctx context.Context, code string) (bool, error)
}

type repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new phrase repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

// CreatePhrase inserts a new phrase into the database.
func (r *repository) CreatePhrase(ctx context.Context, p *domain.Phrase) error {
	query := `
		INSERT INTO phrases (
			id, text_latin, text_native_script, script_type, translation,
			language_code, status, script_status, audio_status,
			contributor_id, cultural_context_id,
			upvote_count, downvote_count, flag_count,
			audio_upvote_count, audio_downvote_count,
			script_upvote_count, script_downvote_count,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11,
			0, 0, 0,
			0, 0,
			0, 0,
			NOW(), NOW()
		)
		RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		p.ID,
		p.TextLatin,
		p.TextNativeScript,
		p.ScriptType,
		p.Translation,
		p.LanguageCode,
		p.Status,
		p.ScriptStatus,
		p.AudioStatus,
		p.ContributorID,
		p.CulturalContextID,
	).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create phrase: %w", err)
	}
	return nil
}

// GetPhraseByID retrieves a phrase by its UUID, including vote counts.
// Returns ErrNotFound if no phrase exists with that ID.
func (r *repository) GetPhraseByID(ctx context.Context, id uuid.UUID) (*domain.Phrase, error) {
	query := `
		SELECT
			p.id, p.text_latin, p.text_native_script, p.script_type,
			p.translation, p.language_code, p.tone, p.status, p.script_status,
			p.contributor_id, p.cultural_context_id,
			p.audio_url, p.audio_duration_seconds, p.audio_status,
			p.native_script_image_url,
			p.moderated_by, p.moderated_at,
			p.upvote_count, p.downvote_count, p.flag_count,
			p.audio_upvote_count, p.audio_downvote_count,
			p.script_upvote_count, p.script_downvote_count,
			p.created_at, p.updated_at
		FROM phrases p
		WHERE p.id = $1`

	p := &domain.Phrase{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.TextLatin, &p.TextNativeScript, &p.ScriptType,
		&p.Translation, &p.LanguageCode, &p.Tone, &p.Status, &p.ScriptStatus,
		&p.ContributorID, &p.CulturalContextID,
		&p.AudioURL, &p.AudioDurationSeconds, &p.AudioStatus,
		&p.NativeScriptImageURL,
		&p.ModeratedBy, &p.ModeratedAt,
		&p.UpvoteCount, &p.DownvoteCount, &p.FlagCount,
		&p.AudioUpvoteCount, &p.AudioDownvoteCount,
		&p.ScriptUpvoteCount, &p.ScriptDownvoteCount,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get phrase by id: %w", err)
	}
	return p, nil
}

// ListPendingPhrases returns all phrases with status 'pending', ordered by creation time.
// Used by contributors to find phrases available for voting.
func (r *repository) ListPendingPhrases(ctx context.Context) ([]domain.Phrase, error) {
	query := `
		SELECT
			p.id, p.text_latin, p.text_native_script, p.script_type,
			p.translation, p.language_code, p.tone, p.status, p.script_status,
			p.contributor_id, p.cultural_context_id,
			p.audio_url, p.audio_duration_seconds, p.audio_status,
			p.native_script_image_url,
			p.moderated_by, p.moderated_at,
			p.upvote_count, p.downvote_count, p.flag_count,
			p.audio_upvote_count, p.audio_downvote_count,
			p.script_upvote_count, p.script_downvote_count,
			p.created_at, p.updated_at
		FROM phrases p
		WHERE p.status = 'pending'
		ORDER BY p.created_at ASC`

	return r.scanPhrases(ctx, query)
}

// ListPhrasesByContributor returns all phrases submitted by a specific contributor,
// ordered by creation time descending (newest first).
func (r *repository) ListPhrasesByContributor(ctx context.Context, contributorID uuid.UUID) ([]domain.Phrase, error) {
	query := `
		SELECT
			p.id, p.text_latin, p.text_native_script, p.script_type,
			p.translation, p.language_code, p.tone, p.status, p.script_status,
			p.contributor_id, p.cultural_context_id,
			p.audio_url, p.audio_duration_seconds, p.audio_status,
			p.native_script_image_url,
			p.moderated_by, p.moderated_at,
			p.upvote_count, p.downvote_count, p.flag_count,
			p.audio_upvote_count, p.audio_downvote_count,
			p.script_upvote_count, p.script_downvote_count,
			p.created_at, p.updated_at
		FROM phrases p
		WHERE p.contributor_id = $1
		ORDER BY p.created_at DESC`

	return r.scanPhrases(ctx, query, contributorID)
}

// IsLanguageActive checks whether a language code exists and is active.
func (r *repository) IsLanguageActive(ctx context.Context, code string) (bool, error) {
	query := `SELECT is_active FROM languages WHERE code = $1`

	var isActive bool
	err := r.pool.QueryRow(ctx, query, code).Scan(&isActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil // language doesn't exist → treat as inactive
		}
		return false, fmt.Errorf("check language active: %w", err)
	}
	return isActive, nil
}

// scanPhrases is a helper that executes a query and scans the result rows into a slice of Phrase.
func (r *repository) scanPhrases(ctx context.Context, query string, args ...any) ([]domain.Phrase, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query phrases: %w", err)
	}
	defer rows.Close()

	var phrases []domain.Phrase
	for rows.Next() {
		var p domain.Phrase
		if err := rows.Scan(
			&p.ID, &p.TextLatin, &p.TextNativeScript, &p.ScriptType,
			&p.Translation, &p.LanguageCode, &p.Tone, &p.Status, &p.ScriptStatus,
			&p.ContributorID, &p.CulturalContextID,
			&p.AudioURL, &p.AudioDurationSeconds, &p.AudioStatus,
			&p.NativeScriptImageURL,
			&p.ModeratedBy, &p.ModeratedAt,
			&p.UpvoteCount, &p.DownvoteCount, &p.FlagCount,
			&p.AudioUpvoteCount, &p.AudioDownvoteCount,
			&p.ScriptUpvoteCount, &p.ScriptDownvoteCount,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan phrase: %w", err)
		}
		phrases = append(phrases, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate phrases: %w", err)
	}

	if phrases == nil {
		phrases = []domain.Phrase{}
	}
	return phrases, nil
}
