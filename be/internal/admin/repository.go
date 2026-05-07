package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
)

var ErrNotFound = errors.New("not found")

// FlaggedPhrase is a phrase with its flag details for admin moderation.
type FlaggedPhrase struct {
	domain.Phrase
	Flags []FlagDetail `json:"flags"`
}

// FlagDetail holds info about a single flag.
type FlagDetail struct {
	UserID    uuid.UUID        `json:"user_id"`
	Reason    domain.FlagReason `json:"reason"`
	CreatedAt time.Time        `json:"created_at"`
}

// UserSummary is a user record for admin list views.
type UserSummary struct {
	ID        uuid.UUID   `json:"id"`
	Name      string      `json:"name"`
	Email     string      `json:"email"`
	Role      domain.Role `json:"role"`
	IsActive  bool        `json:"is_active"`
	CreatedAt time.Time   `json:"created_at"`
}

// Repository defines data access for admin operations.
type Repository interface {
	ListFlaggedPhrases(ctx context.Context) ([]FlaggedPhrase, error)
	UpdatePhraseStatus(ctx context.Context, phraseID, adminID uuid.UUID, status domain.PhraseStatus) error
	ListUsers(ctx context.Context, search string) ([]UserSummary, error)
	BanUser(ctx context.Context, userID uuid.UUID) error
	AssignRole(ctx context.Context, targetUserID, adminID uuid.UUID, role domain.Role) error
	DeletePhrase(ctx context.Context, phraseID uuid.UUID) error
	GetPhraseByID(ctx context.Context, phraseID uuid.UUID) (*domain.Phrase, error)
}

type repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new admin repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) ListFlaggedPhrases(ctx context.Context) ([]FlaggedPhrase, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			p.id, p.text_latin, p.text_native_script, p.script_type,
			p.translation, p.language_code, p.tone, p.status, p.script_status,
			p.contributor_id, p.cultural_context_id,
			p.audio_url, p.audio_duration_seconds, p.audio_status,
			p.native_script_image_url, p.moderated_by, p.moderated_at,
			p.upvote_count, p.downvote_count, p.flag_count,
			p.audio_upvote_count, p.audio_downvote_count,
			p.script_upvote_count, p.script_downvote_count,
			p.created_at, p.updated_at
		FROM phrases p
		WHERE p.status = 'flagged'
		ORDER BY p.updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list flagged phrases: %w", err)
	}
	defer rows.Close()

	var phrases []FlaggedPhrase
	for rows.Next() {
		var fp FlaggedPhrase
		p := &fp.Phrase
		if err := rows.Scan(
			&p.ID, &p.TextLatin, &p.TextNativeScript, &p.ScriptType,
			&p.Translation, &p.LanguageCode, &p.Tone, &p.Status, &p.ScriptStatus,
			&p.ContributorID, &p.CulturalContextID,
			&p.AudioURL, &p.AudioDurationSeconds, &p.AudioStatus,
			&p.NativeScriptImageURL, &p.ModeratedBy, &p.ModeratedAt,
			&p.UpvoteCount, &p.DownvoteCount, &p.FlagCount,
			&p.AudioUpvoteCount, &p.AudioDownvoteCount,
			&p.ScriptUpvoteCount, &p.ScriptDownvoteCount,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan flagged phrase: %w", err)
		}
		phrases = append(phrases, fp)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate flagged phrases: %w", err)
	}

	// Load flags for each phrase
	for i := range phrases {
		flags, err := r.loadFlags(ctx, phrases[i].ID)
		if err != nil {
			return nil, err
		}
		phrases[i].Flags = flags
	}

	if phrases == nil {
		phrases = []FlaggedPhrase{}
	}
	return phrases, nil
}

func (r *repository) loadFlags(ctx context.Context, phraseID uuid.UUID) ([]FlagDetail, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT user_id, reason, created_at FROM flags WHERE phrase_id = $1 ORDER BY created_at ASC`,
		phraseID,
	)
	if err != nil {
		return nil, fmt.Errorf("load flags: %w", err)
	}
	defer rows.Close()

	var flags []FlagDetail
	for rows.Next() {
		var f FlagDetail
		if err := rows.Scan(&f.UserID, &f.Reason, &f.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan flag: %w", err)
		}
		flags = append(flags, f)
	}
	return flags, rows.Err()
}

func (r *repository) UpdatePhraseStatus(ctx context.Context, phraseID, adminID uuid.UUID, status domain.PhraseStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE phrases SET status = $1, moderated_by = $2, moderated_at = NOW(), updated_at = NOW() WHERE id = $3`,
		status, adminID, phraseID,
	)
	if err != nil {
		return fmt.Errorf("update phrase status: %w", err)
	}
	return nil
}

func (r *repository) ListUsers(ctx context.Context, search string) ([]UserSummary, error) {
	args := []any{}
	where := ""
	if search != "" {
		args = append(args, "%"+strings.ToLower(search)+"%")
		where = "WHERE LOWER(name) LIKE $1 OR LOWER(email) LIKE $1"
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, name, email, role, is_active, created_at FROM users `+where+` ORDER BY created_at DESC`,
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []UserSummary
	for rows.Next() {
		var u UserSummary
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	if users == nil {
		users = []UserSummary{}
	}
	return users, rows.Err()
}

func (r *repository) BanUser(ctx context.Context, userID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `UPDATE users SET is_active = false, updated_at = NOW() WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("ban user: %w", err)
	}

	// Reject all pending phrases by this user
	_, err = tx.Exec(ctx,
		`UPDATE phrases SET status = 'rejected', updated_at = NOW()
		 WHERE contributor_id = $1 AND status = 'pending'`, userID,
	)
	if err != nil {
		return fmt.Errorf("reject pending phrases: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *repository) AssignRole(ctx context.Context, targetUserID, adminID uuid.UUID, role domain.Role) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2`,
		role, targetUserID,
	)
	if err != nil {
		return fmt.Errorf("assign role: %w", err)
	}
	return nil
}

func (r *repository) DeletePhrase(ctx context.Context, phraseID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM phrases WHERE id = $1`, phraseID)
	if err != nil {
		return fmt.Errorf("delete phrase: %w", err)
	}
	return nil
}

func (r *repository) GetPhraseByID(ctx context.Context, phraseID uuid.UUID) (*domain.Phrase, error) {
	var p domain.Phrase
	err := r.pool.QueryRow(ctx, `
		SELECT id, audio_url, native_script_image_url, contributor_id
		FROM phrases WHERE id = $1`, phraseID,
	).Scan(&p.ID, &p.AudioURL, &p.NativeScriptImageURL, &p.ContributorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get phrase: %w", err)
	}
	return &p, nil
}