package flashcard

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
)

var ErrInsufficientContent = errors.New("insufficient approved content for this language")

// FlashcardFilter holds query parameters for listing flashcards.
type FlashcardFilter struct {
	LanguageCode string
	Tone         domain.Tone
	Cursor       string
}

// FlashcardPage is the paginated response for flashcards.
type FlashcardPage struct {
	Flashcards []domain.Phrase `json:"flashcards"`
	NextCursor string          `json:"next_cursor,omitempty"`
}

// ConversationScenario groups phrases by cultural context.
type ConversationScenario struct {
	LanguageCode   string          `json:"language_code"`
	UsageSituation string          `json:"usage_situation"`
	TotalPhrases   int             `json:"total_phrases"`
	Phrases        []domain.Phrase `json:"phrases"`
}

// PracticeResultRequest holds the input for recording a practice result.
type PracticeResultRequest struct {
	PhraseID string                `json:"phrase_id"`
	Result   domain.PracticeResult `json:"result"`
}

// Service defines business logic for flashcards, conversation scenarios, and practice.
type Service interface {
	ListFlashcards(ctx context.Context, filter FlashcardFilter) (*FlashcardPage, error)
	GetConversationScenario(ctx context.Context, languageCode string) (*ConversationScenario, error)
	SavePracticeResult(ctx context.Context, learnerID uuid.UUID, req PracticeResultRequest) error
}

type service struct {
	repo Repository
}

// NewService creates a new flashcard service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ListFlashcards returns paginated, randomised approved flashcards.
func (s *service) ListFlashcards(ctx context.Context, filter FlashcardFilter) (*FlashcardPage, error) {
	phrases, nextCursor, err := s.repo.ListFlashcards(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list flashcards: %w", err)
	}
	return &FlashcardPage{Flashcards: phrases, NextCursor: nextCursor}, nil
}

// GetConversationScenario builds a scenario of 3–8 phrases for the language.
func (s *service) GetConversationScenario(ctx context.Context, languageCode string) (*ConversationScenario, error) {
	phrases, err := s.repo.ListConversationPhrases(ctx, languageCode)
	if err != nil {
		return nil, fmt.Errorf("list conversation phrases: %w", err)
	}
	if len(phrases) < 3 {
		return nil, ErrInsufficientContent
	}

	// Pick the largest cultural_context group (or first 8 if no contexts)
	scenario := buildScenario(phrases, languageCode)
	return scenario, nil
}

// SavePracticeResult records a learner's self-assessment.
func (s *service) SavePracticeResult(ctx context.Context, learnerID uuid.UUID, req PracticeResultRequest) error {
	phraseID, err := uuid.Parse(req.PhraseID)
	if err != nil {
		return fmt.Errorf("invalid phrase_id: %w", err)
	}
	if req.Result != domain.PracticeResultTahu && req.Result != domain.PracticeResultBelumTahu {
		return fmt.Errorf("result must be 'tahu' or 'belum_tahu'")
	}

	result := domain.PhrasePracticeResult{
		ID:        uuid.New(),
		LearnerID: learnerID,
		PhraseID:  phraseID,
		Result:    req.Result,
	}
	return s.repo.SavePracticeResult(ctx, result)
}

// buildScenario picks the best 3–8 phrases for a conversation scenario.
func buildScenario(phrases []domain.Phrase, languageCode string) *ConversationScenario {
	// Group by cultural context
	type group struct {
		situation string
		phrases   []domain.Phrase
	}

	groups := map[string]*group{}
	noContext := &group{situation: "Umum", phrases: []domain.Phrase{}}

	for _, p := range phrases {
		if p.CulturalContext != nil && p.CulturalContext.UsageSituation != "" {
			sit := p.CulturalContext.UsageSituation
			if _, ok := groups[sit]; !ok {
				groups[sit] = &group{situation: sit}
			}
			groups[sit].phrases = append(groups[sit].phrases, p)
		} else {
			noContext.phrases = append(noContext.phrases, p)
		}
	}

	// Find the largest group with ≥ 3 phrases
	best := noContext
	for _, g := range groups {
		if len(g.phrases) > len(best.phrases) {
			best = g
		}
	}

	// Clamp to 3–8
	selected := best.phrases
	if len(selected) > 8 {
		selected = selected[:8]
	}

	return &ConversationScenario{
		LanguageCode:   languageCode,
		UsageSituation: best.situation,
		TotalPhrases:   len(selected),
		Phrases:        selected,
	}
}