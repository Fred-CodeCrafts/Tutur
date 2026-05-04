package language

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
)

// CreateLanguageRequest holds the fields required to create a new language.
type CreateLanguageRequest struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Region string `json:"region"`
}

// ToggleActiveRequest holds the desired active state for a language.
type ToggleActiveRequest struct {
	IsActive bool `json:"is_active"`
}

// Service defines the business logic interface for language operations.
type Service interface {
	ListLanguages(ctx context.Context) ([]domain.Language, error)
	CreateLanguage(ctx context.Context, req CreateLanguageRequest) (*domain.Language, error)
	ToggleActive(ctx context.Context, code string, isActive bool) (*domain.Language, error)
}

type service struct {
	repo Repository
}

// NewService creates a new language service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ListLanguages returns all languages.
func (s *service) ListLanguages(ctx context.Context) ([]domain.Language, error) {
	langs, err := s.repo.ListLanguages(ctx)
	if err != nil {
		return nil, fmt.Errorf("list languages: %w", err)
	}
	return langs, nil
}

// CreateLanguage validates and creates a new language entry.
func (s *service) CreateLanguage(ctx context.Context, req CreateLanguageRequest) (*domain.Language, error) {
	lang := &domain.Language{
		Code:     strings.ToLower(strings.TrimSpace(req.Code)),
		Name:     strings.TrimSpace(req.Name),
		Region:   strings.TrimSpace(req.Region),
		IsActive: true,
	}

	if err := s.repo.CreateLanguage(ctx, lang); err != nil {
		if errors.Is(err, ErrDuplicateCode) {
			return nil, ErrDuplicateCode
		}
		return nil, fmt.Errorf("create language: %w", err)
	}
	return lang, nil
}

// ToggleActive sets the is_active flag for a language and returns the updated language.
func (s *service) ToggleActive(ctx context.Context, code string, isActive bool) (*domain.Language, error) {
	if err := s.repo.SetLanguageActive(ctx, code, isActive); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("toggle active: %w", err)
	}

	lang, err := s.repo.GetLanguageByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("fetch updated language: %w", err)
	}
	return lang, nil
}
