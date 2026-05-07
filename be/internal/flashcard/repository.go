package flashcard

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
)

// Repository defines data access for flashcard operations.
type Repository interface {
	ListFlashcards(ctx context.Context, filter FlashcardFilter) ([]domain.Phrase, string, error)
	ListConversationPhrases(ctx context.Context, languageCode string) ([]domain.Phrase, error)
	SavePracticeResult(ctx context.Context, result domain.PhrasePracticeResult) error
}

type repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new flashcard repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

// ListFlashcards returns up to 20 approved phrases with cursor-based pagination.
// Results include joined words and cultural context.
func (r *repository) ListFlashcards(ctx context.Context, f FlashcardFilter) ([]domain.Phrase, string, error) {
	args := []any{f.LanguageCode}
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
			p.created_at, p.updated_at,
			cc.id, cc.language_code, cc.region, cc.usage_situation, cc.created_at
		FROM phrases p
		LEFT JOIN cultural_contexts cc ON cc.id = p.cultural_context_id
		WHERE p.status = 'approved'
		  AND p.language_code = $1`

	argIdx := 2
	if f.Tone != "" {
		args = append(args, string(f.Tone))
		query += fmt.Sprintf(" AND p.tone = $%d", argIdx)
		argIdx++
	}
	if f.Cursor != "" {
		args = append(args, f.Cursor)
		query += fmt.Sprintf(" AND p.id > $%d", argIdx)
		argIdx++
	}

	query += " ORDER BY p.id LIMIT 21" // fetch 21 to detect next page

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("list flashcards: %w", err)
	}
	defer rows.Close()

	var phrases []domain.Phrase
	for rows.Next() {
		var p domain.Phrase
		var cc domain.CulturalContext
		var ccID, ccLangCode, ccRegion, ccSituation *string
		var ccCreatedAt *interface{}

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
			&ccID, &ccLangCode, &ccRegion, &ccSituation, &ccCreatedAt,
		); err != nil {
			return nil, "", fmt.Errorf("scan flashcard: %w", err)
		}

		if ccID != nil {
			cc.LanguageCode = derefStr(ccLangCode)
			cc.Region = derefStr(ccRegion)
			cc.UsageSituation = derefStr(ccSituation)
			p.CulturalContext = &cc
		}

		// Only include native script if approved
		if p.ScriptStatus != domain.ScriptStatusApproved {
			p.TextNativeScript = nil
			p.ScriptType = nil
		}
		// Only include audio if approved
		if p.AudioStatus != domain.AudioStatusApproved {
			p.AudioURL = nil
			p.AudioDurationSeconds = nil
		}

		phrases = append(phrases, p)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterate flashcards: %w", err)
	}

	// Determine next cursor
	var nextCursor string
	if len(phrases) == 21 {
		nextCursor = phrases[20].ID.String()
		phrases = phrases[:20]
	}

	// Load words for each phrase
	for i := range phrases {
		words, err := r.loadWords(ctx, phrases[i].ID.String())
		if err != nil {
			return nil, "", err
		}
		phrases[i].Words = words
	}

	if phrases == nil {
		phrases = []domain.Phrase{}
	}
	return phrases, nextCursor, nil
}

// ListConversationPhrases returns approved phrases for a language, grouped by cultural context.
func (r *repository) ListConversationPhrases(ctx context.Context, languageCode string) ([]domain.Phrase, error) {
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
			p.created_at, p.updated_at,
			cc.id, cc.language_code, cc.region, cc.usage_situation, cc.created_at
		FROM phrases p
		LEFT JOIN cultural_contexts cc ON cc.id = p.cultural_context_id
		WHERE p.status = 'approved'
		  AND p.language_code = $1
		ORDER BY p.cultural_context_id NULLS LAST, p.created_at ASC`

	rows, err := r.pool.Query(ctx, query, languageCode)
	if err != nil {
		return nil, fmt.Errorf("list conversation phrases: %w", err)
	}
	defer rows.Close()

	var phrases []domain.Phrase
	for rows.Next() {
		var p domain.Phrase
		var ccID, ccLangCode, ccRegion, ccSituation *string
		var ccCreatedAt *interface{}
		var cc domain.CulturalContext

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
			&ccID, &ccLangCode, &ccRegion, &ccSituation, &ccCreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan conversation phrase: %w", err)
		}

		if ccID != nil {
			cc.LanguageCode = derefStr(ccLangCode)
			cc.Region = derefStr(ccRegion)
			cc.UsageSituation = derefStr(ccSituation)
			p.CulturalContext = &cc
		}
		if p.AudioStatus != domain.AudioStatusApproved {
			p.AudioURL = nil
		}

		phrases = append(phrases, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate conversation phrases: %w", err)
	}
	if phrases == nil {
		phrases = []domain.Phrase{}
	}
	return phrases, nil
}

// SavePracticeResult persists a learner's self-assessment result.
func (r *repository) SavePracticeResult(ctx context.Context, result domain.PhrasePracticeResult) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO phrase_practice_results (id, learner_id, phrase_id, result, created_at)
		 VALUES ($1, $2, $3, $4, NOW())`,
		result.ID, result.LearnerID, result.PhraseID, result.Result,
	)
	if err != nil {
		return fmt.Errorf("save practice result: %w", err)
	}
	return nil
}

// loadWords fetches the words associated with a phrase.
func (r *repository) loadWords(ctx context.Context, phraseID string) ([]domain.Word, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT w.id, w.surface_form_latin, w.root_form_latin, w.part_of_speech,
		       w.language_code, pw.position, w.created_at
		FROM words w
		JOIN phrase_words pw ON pw.word_id = w.id
		WHERE pw.phrase_id = $1
		ORDER BY pw.position ASC`, phraseID)
	if err != nil {
		return nil, fmt.Errorf("load words: %w", err)
	}
	defer rows.Close()

	var words []domain.Word
	for rows.Next() {
		var w domain.Word
		if err := rows.Scan(&w.ID, &w.SurfaceFormLatin, &w.RootFormLatin,
			&w.PartOfSpeech, &w.LanguageCode, &w.Position, &w.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan word: %w", err)
		}
		words = append(words, w)
	}
	return words, rows.Err()
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}