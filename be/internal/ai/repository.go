package ai

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
)

// Repository defines data access for AI pipeline results.
type Repository interface {
	SaveAIResults(ctx context.Context, phraseID uuid.UUID, tone domain.Tone, words []domain.Word) error
	UpdatePhraseAIFailed(ctx context.Context, phraseID uuid.UUID) error
}

type repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new AI repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

// SaveAIResults persists words and tone in a single transaction.
func (r *repository) SaveAIResults(ctx context.Context, phraseID uuid.UUID, tone domain.Tone, words []domain.Word) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx,
		`UPDATE phrases SET tone = $1, updated_at = NOW() WHERE id = $2`,
		tone, phraseID,
	)
	if err != nil {
		return fmt.Errorf("update phrase tone: %w", err)
	}

	var languageCode string
	if err = tx.QueryRow(ctx, `SELECT language_code FROM phrases WHERE id = $1`, phraseID).Scan(&languageCode); err != nil {
		return fmt.Errorf("get language code: %w", err)
	}

	for i, word := range words {
		wordID := uuid.New()
		if _, err = tx.Exec(ctx,
			`INSERT INTO words (id, surface_form_latin, root_form_latin, part_of_speech, language_code, created_at)
			 VALUES ($1, $2, $3, $4, $5, NOW())`,
			wordID, word.SurfaceFormLatin, word.RootFormLatin, word.PartOfSpeech, languageCode,
		); err != nil {
			return fmt.Errorf("insert word: %w", err)
		}

		if _, err = tx.Exec(ctx,
			`INSERT INTO phrase_words (phrase_id, word_id, position) VALUES ($1, $2, $3)`,
			phraseID, wordID, i,
		); err != nil {
			return fmt.Errorf("insert phrase_word: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// UpdatePhraseAIFailed marks a phrase as ai_failed.
func (r *repository) UpdatePhraseAIFailed(ctx context.Context, phraseID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE phrases SET status = 'ai_failed', updated_at = NOW() WHERE id = $1`,
		phraseID,
	)
	return err
}