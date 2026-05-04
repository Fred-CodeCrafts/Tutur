package validation

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
)

// Threshold constants for status transitions.
const (
	upvoteApproveThreshold  = 3
	downvoteRejectThreshold = 5
	flagThreshold           = 3
)

// Engine handles threshold-based status transitions for phrases, audio, and scripts.
type Engine struct {
	repo Repository
}

// NewEngine creates a new validation engine.
func NewEngine(repo Repository) *Engine {
	return &Engine{repo: repo}
}

// CheckPhraseThresholds evaluates vote counts and updates phrase status if a
// threshold is crossed. It is called after every text vote operation.
//
// Transitions:
//   - upvote_count >= 3  → status = approved
//   - downvote_count >= 5 → status = rejected
func (e *Engine) CheckPhraseThresholds(ctx context.Context, phraseID uuid.UUID, upvotes, downvotes int) error {
	var newStatus domain.PhraseStatus

	switch {
	case upvotes >= upvoteApproveThreshold:
		newStatus = domain.StatusApproved
	case downvotes >= downvoteRejectThreshold:
		newStatus = domain.StatusRejected
	default:
		return nil // no threshold crossed
	}

	if err := e.repo.UpdatePhraseStatus(ctx, phraseID, newStatus); err != nil {
		return fmt.Errorf("check phrase thresholds: %w", err)
	}
	return nil
}

// CheckFlagThreshold evaluates flag count and updates phrase status to flagged
// if the threshold is crossed. Called after every flag operation.
//
// Transition:
//   - flag_count >= 3 → status = flagged
func (e *Engine) CheckFlagThreshold(ctx context.Context, phraseID uuid.UUID, flagCount int) error {
	if flagCount < flagThreshold {
		return nil
	}

	if err := e.repo.UpdatePhraseStatus(ctx, phraseID, domain.StatusFlagged); err != nil {
		return fmt.Errorf("check flag threshold: %w", err)
	}
	return nil
}

// CheckAudioThresholds evaluates audio vote counts and updates audio_status.
// Called after every audio vote operation.
//
// Transitions:
//   - audio_upvote_count >= 3  → audio_status = audio_approved
//   - audio_downvote_count >= 5 → audio_status = audio_rejected
func (e *Engine) CheckAudioThresholds(ctx context.Context, phraseID uuid.UUID, upvotes, downvotes int) error {
	var newStatus domain.AudioStatus

	switch {
	case upvotes >= upvoteApproveThreshold:
		newStatus = domain.AudioStatusApproved
	case downvotes >= downvoteRejectThreshold:
		newStatus = domain.AudioStatusRejected
	default:
		return nil
	}

	if err := e.repo.UpdateAudioStatus(ctx, phraseID, newStatus); err != nil {
		return fmt.Errorf("check audio thresholds: %w", err)
	}
	return nil
}

// CheckScriptThresholds evaluates script vote counts and updates script_status.
// Called after every script vote operation.
//
// Transitions:
//   - script_upvote_count >= 3  → script_status = approved
//   - script_downvote_count >= 5 → script_status = rejected
func (e *Engine) CheckScriptThresholds(ctx context.Context, phraseID uuid.UUID, upvotes, downvotes int) error {
	var newStatus domain.ScriptStatus

	switch {
	case upvotes >= upvoteApproveThreshold:
		newStatus = domain.ScriptStatusApproved
	case downvotes >= downvoteRejectThreshold:
		newStatus = domain.ScriptStatusRejected
	default:
		return nil
	}

	if err := e.repo.UpdateScriptStatus(ctx, phraseID, newStatus); err != nil {
		return fmt.Errorf("check script thresholds: %w", err)
	}
	return nil
}
