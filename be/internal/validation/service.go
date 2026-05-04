package validation

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
)

// Sentinel errors returned by the service layer.
var (
	// ErrPhraseNotFound is returned when the target phrase does not exist.
	ErrPhraseNotFound = errors.New("phrase not found")
	// ErrDuplicateVote is returned when a contributor tries to vote twice.
	ErrDuplicateVote = errors.New("duplicate vote")
	// ErrSelfVote is returned when a contributor tries to vote on their own phrase.
	ErrSelfVote = errors.New("self vote not allowed")
)

// VoteRequest holds the input for a vote operation.
type VoteRequest struct {
	VoteType domain.VoteType `json:"vote_type"`
}

// FlagRequest holds the input for a flag operation.
type FlagRequest struct {
	Reason domain.FlagReason `json:"reason"`
}

// Service defines the business logic interface for validation operations.
type Service interface {
	// VotePhrase records a text vote on a phrase.
	VotePhrase(ctx context.Context, phraseID, contributorID uuid.UUID, req VoteRequest) error
	// FlagPhrase records a flag on a phrase.
	FlagPhrase(ctx context.Context, phraseID, userID uuid.UUID, req FlagRequest) error
	// VoteAudio records an audio vote on a phrase.
	VoteAudio(ctx context.Context, phraseID, contributorID uuid.UUID, req VoteRequest) error
	// VoteScript records a script vote on a phrase.
	VoteScript(ctx context.Context, phraseID, contributorID uuid.UUID, req VoteRequest) error
}

type service struct {
	repo   Repository
	engine *Engine
}

// NewService creates a new validation service.
func NewService(repo Repository) Service {
	return &service{
		repo:   repo,
		engine: NewEngine(repo),
	}
}

// VotePhrase validates and records a text vote, then checks thresholds.
// Returns ErrPhraseNotFound, ErrSelfVote, or ErrDuplicateVote on failure.
func (s *service) VotePhrase(ctx context.Context, phraseID, contributorID uuid.UUID, req VoteRequest) error {
	if err := validateVoteType(req.VoteType); err != nil {
		return err
	}

	// Self-vote check
	ownerID, err := s.repo.GetPhraseContributorID(ctx, phraseID)
	if err != nil {
		return fmt.Errorf("vote phrase: %w", err)
	}
	if ownerID == contributorID {
		return ErrSelfVote
	}

	upvotes, downvotes, err := s.repo.InsertVoteAndUpdateCount(ctx, phraseID, contributorID, req.VoteType)
	if err != nil {
		return fmt.Errorf("vote phrase: %w", err)
	}

	if err := s.engine.CheckPhraseThresholds(ctx, phraseID, upvotes, downvotes); err != nil {
		return fmt.Errorf("vote phrase threshold check: %w", err)
	}
	return nil
}

// FlagPhrase validates and records a flag, then checks the flag threshold.
// Returns ErrPhraseNotFound or ErrDuplicateVote on failure.
func (s *service) FlagPhrase(ctx context.Context, phraseID, userID uuid.UUID, req FlagRequest) error {
	if err := validateFlagReason(req.Reason); err != nil {
		return err
	}

	// Verify phrase exists
	if _, err := s.repo.GetPhraseContributorID(ctx, phraseID); err != nil {
		return fmt.Errorf("flag phrase: %w", err)
	}

	flagCount, err := s.repo.InsertFlagAndUpdateCount(ctx, phraseID, userID, req.Reason)
	if err != nil {
		return fmt.Errorf("flag phrase: %w", err)
	}

	if err := s.engine.CheckFlagThreshold(ctx, phraseID, flagCount); err != nil {
		return fmt.Errorf("flag phrase threshold check: %w", err)
	}
	return nil
}

// VoteAudio validates and records an audio vote, then checks audio thresholds.
// Returns ErrPhraseNotFound, ErrSelfVote, or ErrDuplicateVote on failure.
func (s *service) VoteAudio(ctx context.Context, phraseID, contributorID uuid.UUID, req VoteRequest) error {
	if err := validateVoteType(req.VoteType); err != nil {
		return err
	}

	// Self-vote check
	ownerID, err := s.repo.GetPhraseContributorID(ctx, phraseID)
	if err != nil {
		return fmt.Errorf("vote audio: %w", err)
	}
	if ownerID == contributorID {
		return ErrSelfVote
	}

	upvotes, downvotes, err := s.repo.InsertAudioVoteAndUpdateCount(ctx, phraseID, contributorID, req.VoteType)
	if err != nil {
		return fmt.Errorf("vote audio: %w", err)
	}

	if err := s.engine.CheckAudioThresholds(ctx, phraseID, upvotes, downvotes); err != nil {
		return fmt.Errorf("vote audio threshold check: %w", err)
	}
	return nil
}

// VoteScript validates and records a script vote, then checks script thresholds.
// Returns ErrPhraseNotFound, ErrSelfVote, or ErrDuplicateVote on failure.
func (s *service) VoteScript(ctx context.Context, phraseID, contributorID uuid.UUID, req VoteRequest) error {
	if err := validateVoteType(req.VoteType); err != nil {
		return err
	}

	// Self-vote check
	ownerID, err := s.repo.GetPhraseContributorID(ctx, phraseID)
	if err != nil {
		return fmt.Errorf("vote script: %w", err)
	}
	if ownerID == contributorID {
		return ErrSelfVote
	}

	upvotes, downvotes, err := s.repo.InsertScriptVoteAndUpdateCount(ctx, phraseID, contributorID, req.VoteType)
	if err != nil {
		return fmt.Errorf("vote script: %w", err)
	}

	if err := s.engine.CheckScriptThresholds(ctx, phraseID, upvotes, downvotes); err != nil {
		return fmt.Errorf("vote script threshold check: %w", err)
	}
	return nil
}

// ── Input validation helpers ──────────────────────────────────────────────────

var ErrInvalidVoteType = errors.New("vote_type must be 'upvote' or 'downvote'")
var ErrInvalidFlagReason = errors.New("reason must be one of: inaccurate_translation, inappropriate_content, duplicate")

func validateVoteType(vt domain.VoteType) error {
	if vt != domain.VoteUpvote && vt != domain.VoteDownvote {
		return ErrInvalidVoteType
	}
	return nil
}

func validateFlagReason(r domain.FlagReason) error {
	switch r {
	case domain.FlagInaccurateTranslation, domain.FlagInappropriateContent, domain.FlagDuplicate:
		return nil
	}
	return ErrInvalidFlagReason
}
