package search

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
)

// Repository defines data access for search.
type Repository interface {
	Search(ctx context.Context, req SearchRequest) ([]domain.Phrase, int, error)
}

type repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new search repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

// Search returns approved phrases matching the query, with offset pagination.
func (r *repository) Search(ctx context.Context, req SearchRequest) ([]domain.Phrase, int, error) {
	q := "%" + strings.ToLower(req.Query) + "%"
	args := []any{q, req.LanguageCode}
	argIdx := 3

	baseWhere := `
		WHERE p.status = 'approved'
		  AND p.language_code = $2
		  AND (
		      LOWER(p.text_latin) LIKE $1
		   OR LOWER(p.translation) LIKE $1`

	// Optionally search by root form via subquery
	if req.SearchByRoot {
		baseWhere += fmt.Sprintf(`
		   OR p.id IN (
		       SELECT pw.phrase_id FROM phrase_words pw
		       JOIN words w ON w.id = pw.word_id
		       WHERE LOWER(w.root_form_latin) LIKE $1
		         AND w.language_code = $%d
		   )`, argIdx)
		args = append(args, req.LanguageCode)
		argIdx++
	}
	baseWhere += ")"

	// Count query
	countQuery := "SELECT COUNT(*) FROM phrases p" + baseWhere
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count search: %w", err)
	}

	// Main query with pagination
	limit := 50
	args = append(args, limit, req.Offset)
	dataQuery := `
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
		FROM phrases p` + baseWhere + fmt.Sprintf(`
		ORDER BY p.created_at DESC
		LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("search query: %w", err)
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
			return nil, 0, fmt.Errorf("scan search result: %w", err)
		}
		phrases = append(phrases, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate search: %w", err)
	}
	if phrases == nil {
		phrases = []domain.Phrase{}
	}
	return phrases, total, nil
}