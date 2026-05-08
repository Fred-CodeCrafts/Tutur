package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines data access for storage-related phrase updates.
type Repository interface {
	UpdatePhraseAudioURL(ctx context.Context, phraseID uuid.UUID, audioURL string, durationSeconds float64) error
	UpdatePhraseImageURL(ctx context.Context, phraseID uuid.UUID, imageURL string) error
	GetPhraseAudioURL(ctx context.Context, phraseID uuid.UUID) (string, error)
	GetPhraseImageURL(ctx context.Context, phraseID uuid.UUID) (string, error)
}

type repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new storage repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) UpdatePhraseAudioURL(ctx context.Context, phraseID uuid.UUID, audioURL string, durationSeconds float64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE phrases SET audio_url = $1, audio_duration_seconds = $2, audio_status = 'pending', updated_at = NOW() WHERE id = $3`,
		audioURL, durationSeconds, phraseID,
	)
	if err != nil {
		return fmt.Errorf("update phrase audio url: %w", err)
	}
	return nil
}

func (r *repository) UpdatePhraseImageURL(ctx context.Context, phraseID uuid.UUID, imageURL string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE phrases SET native_script_image_url = $1, updated_at = NOW() WHERE id = $2`,
		imageURL, phraseID,
	)
	if err != nil {
		return fmt.Errorf("update phrase image url: %w", err)
	}
	return nil
}

func (r *repository) GetPhraseAudioURL(ctx context.Context, phraseID uuid.UUID) (string, error) {
	var url string
	err := r.pool.QueryRow(ctx, `SELECT COALESCE(audio_url, '') FROM phrases WHERE id = $1`, phraseID).Scan(&url)
	if err != nil {
		return "", fmt.Errorf("get phrase audio url: %w", err)
	}
	return url, nil
}

func (r *repository) GetPhraseImageURL(ctx context.Context, phraseID uuid.UUID) (string, error) {
	var url string
	err := r.pool.QueryRow(ctx, `SELECT COALESCE(native_script_image_url, '') FROM phrases WHERE id = $1`, phraseID).Scan(&url)
	if err != nil {
		return "", fmt.Errorf("get phrase image url: %w", err)
	}
	return url, nil
}