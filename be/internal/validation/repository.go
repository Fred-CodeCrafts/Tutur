package validation

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
)

// pgUniqueViolation is the PostgreSQL error code for unique constraint violations.
const pgUniqueViolation = "23505"

// Repository defines the data access interface for validation operations.
type Repository interface {
	// Phrase votes
	GetPhraseContributorID(ctx context.Context, phraseID uuid.UUID) (uuid.UUID, error)
	InsertVoteAndUpdateCount(ctx context.Context, phraseID, contributorID uuid.UUID, voteType domain.VoteType) (upvotes, downvotes int, err error)
	UpdatePhraseStatus(ctx context.Context, phraseID uuid.UUID, status domain.PhraseStatus) error

	// Flags
	InsertFlagAndUpdateCount(ctx context.Context, phraseID, userID uuid.UUID, reason domain.FlagReason) (flagCount int, err error)

	// Audio votes
	InsertAudioVoteAndUpdateCount(ctx context.Context, phraseID, contributorID uuid.UUID, voteType domain.VoteType) (upvotes, downvotes int, err error)
	UpdateAudioStatus(ctx context.Context, phraseID uuid.UUID, status domain.AudioStatus) error

	// Script votes
	InsertScriptVoteAndUpdateCount(ctx context.Context, phraseID, contributorID uuid.UUID, voteType domain.VoteType) (upvotes, downvotes int, err error)
	UpdateScriptStatus(ctx context.Context, phraseID uuid.UUID, status domain.ScriptStatus) error
}

type repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new validation repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

// GetPhraseContributorID returns the contributor_id of a phrase.
// Returns ErrPhraseNotFound if the phrase does not exist.
func (r *repository) GetPhraseContributorID(ctx context.Context, phraseID uuid.UUID) (uuid.UUID, error) {
	var contributorID uuid.UUID
	err := r.pool.QueryRow(ctx,
		`SELECT contributor_id FROM phrases WHERE id = $1`, phraseID,
	).Scan(&contributorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrPhraseNotFound
		}
		return uuid.Nil, fmt.Errorf("get phrase contributor: %w", err)
	}
	return contributorID, nil
}

// InsertVoteAndUpdateCount inserts a text vote and atomically increments the
// corresponding count on the phrases row. Returns the updated counts.
// Returns ErrDuplicateVote if the contributor already voted on this phrase.
func (r *repository) InsertVoteAndUpdateCount(
	ctx context.Context,
	phraseID, contributorID uuid.UUID,
	voteType domain.VoteType,
) (upvotes, downvotes int, err error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Insert vote — unique constraint (phrase_id, contributor_id) prevents duplicates
	_, err = tx.Exec(ctx,
		`INSERT INTO votes (id, phrase_id, contributor_id, vote_type, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, NOW())`,
		phraseID, contributorID, voteType,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return 0, 0, ErrDuplicateVote
		}
		return 0, 0, fmt.Errorf("insert vote: %w", err)
	}

	// Atomically increment the appropriate count and return updated values
	var upvoteCount, downvoteCount int
	if voteType == domain.VoteUpvote {
		err = tx.QueryRow(ctx,
			`UPDATE phrases SET upvote_count = upvote_count + 1, updated_at = NOW()
			 WHERE id = $1
			 RETURNING upvote_count, downvote_count`,
			phraseID,
		).Scan(&upvoteCount, &downvoteCount)
	} else {
		err = tx.QueryRow(ctx,
			`UPDATE phrases SET downvote_count = downvote_count + 1, updated_at = NOW()
			 WHERE id = $1
			 RETURNING upvote_count, downvote_count`,
			phraseID,
		).Scan(&upvoteCount, &downvoteCount)
	}
	if err != nil {
		return 0, 0, fmt.Errorf("update vote count: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, 0, fmt.Errorf("commit tx: %w", err)
	}
	return upvoteCount, downvoteCount, nil
}

// UpdatePhraseStatus updates the status of a phrase.
func (r *repository) UpdatePhraseStatus(ctx context.Context, phraseID uuid.UUID, status domain.PhraseStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE phrases SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, phraseID,
	)
	if err != nil {
		return fmt.Errorf("update phrase status: %w", err)
	}
	return nil
}

// InsertFlagAndUpdateCount inserts a flag and atomically increments flag_count.
// Returns ErrDuplicateVote if the user already flagged this phrase.
func (r *repository) InsertFlagAndUpdateCount(
	ctx context.Context,
	phraseID, userID uuid.UUID,
	reason domain.FlagReason,
) (flagCount int, err error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx,
		`INSERT INTO flags (id, phrase_id, user_id, reason, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, NOW())`,
		phraseID, userID, reason,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return 0, ErrDuplicateVote
		}
		return 0, fmt.Errorf("insert flag: %w", err)
	}

	var count int
	err = tx.QueryRow(ctx,
		`UPDATE phrases SET flag_count = flag_count + 1, updated_at = NOW()
		 WHERE id = $1
		 RETURNING flag_count`,
		phraseID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("update flag count: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}
	return count, nil
}

// InsertAudioVoteAndUpdateCount inserts an audio vote and atomically updates counts.
// Returns ErrDuplicateVote if the contributor already voted on this phrase's audio.
func (r *repository) InsertAudioVoteAndUpdateCount(
	ctx context.Context,
	phraseID, contributorID uuid.UUID,
	voteType domain.VoteType,
) (upvotes, downvotes int, err error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx,
		`INSERT INTO audio_votes (id, phrase_id, contributor_id, vote_type, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, NOW())`,
		phraseID, contributorID, voteType,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return 0, 0, ErrDuplicateVote
		}
		return 0, 0, fmt.Errorf("insert audio vote: %w", err)
	}

	var upvoteCount, downvoteCount int
	if voteType == domain.VoteUpvote {
		err = tx.QueryRow(ctx,
			`UPDATE phrases SET audio_upvote_count = audio_upvote_count + 1, updated_at = NOW()
			 WHERE id = $1
			 RETURNING audio_upvote_count, audio_downvote_count`,
			phraseID,
		).Scan(&upvoteCount, &downvoteCount)
	} else {
		err = tx.QueryRow(ctx,
			`UPDATE phrases SET audio_downvote_count = audio_downvote_count + 1, updated_at = NOW()
			 WHERE id = $1
			 RETURNING audio_upvote_count, audio_downvote_count`,
			phraseID,
		).Scan(&upvoteCount, &downvoteCount)
	}
	if err != nil {
		return 0, 0, fmt.Errorf("update audio vote count: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, 0, fmt.Errorf("commit tx: %w", err)
	}
	return upvoteCount, downvoteCount, nil
}

// UpdateAudioStatus updates the audio_status of a phrase.
func (r *repository) UpdateAudioStatus(ctx context.Context, phraseID uuid.UUID, status domain.AudioStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE phrases SET audio_status = $1, updated_at = NOW() WHERE id = $2`,
		status, phraseID,
	)
	if err != nil {
		return fmt.Errorf("update audio status: %w", err)
	}
	return nil
}

// InsertScriptVoteAndUpdateCount inserts a script vote and atomically updates counts.
// Returns ErrDuplicateVote if the contributor already voted on this phrase's script.
func (r *repository) InsertScriptVoteAndUpdateCount(
	ctx context.Context,
	phraseID, contributorID uuid.UUID,
	voteType domain.VoteType,
) (upvotes, downvotes int, err error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx,
		`INSERT INTO script_votes (id, phrase_id, contributor_id, vote_type, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, NOW())`,
		phraseID, contributorID, voteType,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return 0, 0, ErrDuplicateVote
		}
		return 0, 0, fmt.Errorf("insert script vote: %w", err)
	}

	var upvoteCount, downvoteCount int
	if voteType == domain.VoteUpvote {
		err = tx.QueryRow(ctx,
			`UPDATE phrases SET script_upvote_count = script_upvote_count + 1, updated_at = NOW()
			 WHERE id = $1
			 RETURNING script_upvote_count, script_downvote_count`,
			phraseID,
		).Scan(&upvoteCount, &downvoteCount)
	} else {
		err = tx.QueryRow(ctx,
			`UPDATE phrases SET script_downvote_count = script_downvote_count + 1, updated_at = NOW()
			 WHERE id = $1
			 RETURNING script_upvote_count, script_downvote_count`,
			phraseID,
		).Scan(&upvoteCount, &downvoteCount)
	}
	if err != nil {
		return 0, 0, fmt.Errorf("update script vote count: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, 0, fmt.Errorf("commit tx: %w", err)
	}
	return upvoteCount, downvoteCount, nil
}

// UpdateScriptStatus updates the script_status of a phrase.
func (r *repository) UpdateScriptStatus(ctx context.Context, phraseID uuid.UUID, status domain.ScriptStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE phrases SET script_status = $1, updated_at = NOW() WHERE id = $2`,
		status, phraseID,
	)
	if err != nil {
		return fmt.Errorf("update script status: %w", err)
	}
	return nil
}
