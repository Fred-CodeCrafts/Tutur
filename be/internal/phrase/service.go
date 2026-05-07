package phrase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
)

// ErrInactiveLanguage is returned when the language_code is not active or doesn't exist.
var ErrInactiveLanguage = errors.New("language not active")

// SubmitPhraseRequest holds the fields for submitting a new phrase.
type SubmitPhraseRequest struct {
	TextLatin         string             `json:"text_latin"`
	TextNativeScript  *string            `json:"text_native_script,omitempty"`
	ScriptType        *domain.ScriptType `json:"script_type,omitempty"`
	Translation       string             `json:"translation"`
	LanguageCode      string             `json:"language_code"`
	CulturalContextID *uuid.UUID         `json:"cultural_context_id,omitempty"`
}

// SubmitPhraseResponse is returned after a successful phrase submission.
type SubmitPhraseResponse struct {
	ID           uuid.UUID            `json:"id"`
	Status       domain.PhraseStatus  `json:"status"`
	ScriptStatus domain.ScriptStatus  `json:"script_status"`
}

// Enqueuer is a dependency for enqueuing AI jobs (injected from main).
type Enqueuer interface {
	EnqueuePhrase(phraseID uuid.UUID, textLatin, translation string)
}

// Service defines the business logic interface for phrase operations.
type Service interface {
	SubmitPhrase(ctx context.Context, contributorID uuid.UUID, req SubmitPhraseRequest) (*SubmitPhraseResponse, error)
	GetPhraseByID(ctx context.Context, id uuid.UUID) (*domain.Phrase, error)
	ListPendingPhrases(ctx context.Context) ([]domain.Phrase, error)
	ListMyPhrases(ctx context.Context, contributorID uuid.UUID) ([]domain.Phrase, error)
}

type service struct {
	repo     Repository
	enqueuer Enqueuer
}

// NewService creates a new phrase service.
func NewService(repo Repository, enqueuer Enqueuer) Service {
	return &service{repo: repo, enqueuer: enqueuer}
}

// SubmitPhrase validates and persists a new phrase with status 'pending',
// then enqueues it for AI processing.
func (s *service) SubmitPhrase(ctx context.Context, contributorID uuid.UUID, req SubmitPhraseRequest) (*SubmitPhraseResponse, error) {
	// Verify language is active
	active, err := s.repo.IsLanguageActive(ctx, req.LanguageCode)
	if err != nil {
		return nil, fmt.Errorf("check language: %w", err)
	}
	if !active {
		return nil, ErrInactiveLanguage
	}

	// Determine script_status: if native script text is provided, set to pending
	scriptStatus := domain.ScriptStatusNone
	if req.TextNativeScript != nil && strings.TrimSpace(*req.TextNativeScript) != "" {
		scriptStatus = domain.ScriptStatusPending
	}

	p := &domain.Phrase{
		ID:                uuid.New(),
		TextLatin:         strings.TrimSpace(req.TextLatin),
		TextNativeScript:  req.TextNativeScript,
		ScriptType:        req.ScriptType,
		Translation:       strings.TrimSpace(req.Translation),
		LanguageCode:      req.LanguageCode,
		Status:            domain.StatusPending,
		ScriptStatus:      scriptStatus,
		AudioStatus:       domain.AudioStatusNone,
		ContributorID:     contributorID,
		CulturalContextID: req.CulturalContextID,
	}

	if err := s.repo.CreatePhrase(ctx, p); err != nil {
		return nil, fmt.Errorf("create phrase: %w", err)
	}

	// Enqueue async AI processing
	if s.enqueuer != nil {
		s.enqueuer.EnqueuePhrase(p.ID, p.TextLatin, p.Translation)
	}

	return &SubmitPhraseResponse{ID: p.ID, Status: p.Status, ScriptStatus: p.ScriptStatus}, nil
}

// GetPhraseByID retrieves a phrase by its UUID.
func (s *service) GetPhraseByID(ctx context.Context, id uuid.UUID) (*domain.Phrase, error) {
	p, err := s.repo.GetPhraseByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get phrase: %w", err)
	}
	return p, nil
}
// ListPendingPhrases returns all pending phrases for voting.
func (s *service) ListPendingPhrases(ctx context.Context) ([]domain.Phrase, error) {
	phrases, err := s.repo.ListPendingPhrases(ctx)
	if err != nil {
		return nil, fmt.Errorf("list pending phrases: %w", err)
	}
	return phrases, nil
}

// ListMyPhrases returns all phrases submitted by the given contributor.
func (s *service) ListMyPhrases(ctx context.Context, contributorID uuid.UUID) ([]domain.Phrase, error) {
	phrases, err := s.repo.ListPhrasesByContributor(ctx, contributorID)
	if err != nil {
		return nil, fmt.Errorf("list my phrases: %w", err)
	}
	return phrases, nil
}