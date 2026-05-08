package admin

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
	"github.com/yourusername/bahasa-daerah-platform/internal/storage"
)

// Service defines admin business logic.
type Service interface {
	ListFlaggedPhrases(ctx context.Context) ([]FlaggedPhrase, error)
	ModeratePhrase(ctx context.Context, phraseID, adminID uuid.UUID, action string) error
	ListUsers(ctx context.Context, search string) ([]UserSummary, error)
	BanUser(ctx context.Context, userID uuid.UUID) error
	AssignRole(ctx context.Context, targetUserID, adminID uuid.UUID, role domain.Role) error
	DeletePhrase(ctx context.Context, phraseID uuid.UUID) error
}

type service struct {
	repo        Repository
	storageSvc  *storage.Service
}

// NewService creates a new admin service.
func NewService(repo Repository, storageSvc *storage.Service) Service {
	return &service{repo: repo, storageSvc: storageSvc}
}

func (s *service) ListFlaggedPhrases(ctx context.Context) ([]FlaggedPhrase, error) {
	return s.repo.ListFlaggedPhrases(ctx)
}

// ModeratePhrase approves or rejects a flagged phrase.
func (s *service) ModeratePhrase(ctx context.Context, phraseID, adminID uuid.UUID, action string) error {
	var status domain.PhraseStatus
	switch action {
	case "approve":
		status = domain.StatusApproved
	case "reject":
		status = domain.StatusRejected
	default:
		return fmt.Errorf("action must be 'approve' or 'reject'")
	}
	return s.repo.UpdatePhraseStatus(ctx, phraseID, adminID, status)
}

func (s *service) ListUsers(ctx context.Context, search string) ([]UserSummary, error) {
	return s.repo.ListUsers(ctx, search)
}

func (s *service) BanUser(ctx context.Context, userID uuid.UUID) error {
	return s.repo.BanUser(ctx, userID)
}

func (s *service) AssignRole(ctx context.Context, targetUserID, adminID uuid.UUID, role domain.Role) error {
	if role != domain.RoleAdmin && role != domain.RoleContributor && role != domain.RoleLearner {
		return fmt.Errorf("invalid role: %s", role)
	}
	return s.repo.AssignRole(ctx, targetUserID, adminID, role)
}

// DeletePhrase hard-deletes a phrase and its associated files from storage.
func (s *service) DeletePhrase(ctx context.Context, phraseID uuid.UUID) error {
	phrase, err := s.repo.GetPhraseByID(ctx, phraseID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("get phrase for deletion: %w", err)
	}

	// Clean up storage files (best-effort — don't fail deletion if storage fails)
	if phrase.AudioURL != nil && *phrase.AudioURL != "" {
		_ = s.storageSvc.DeleteAudio(ctx, phraseID)
	}
	if phrase.NativeScriptImageURL != nil && *phrase.NativeScriptImageURL != "" {
		_ = s.storageSvc.DeleteImage(ctx, phraseID)
	}

	return s.repo.DeletePhrase(ctx, phraseID)
}