package search

import (
	"context"
	"fmt"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
)

// SearchRequest holds the search parameters.
type SearchRequest struct {
	Query        string
	LanguageCode string
	SearchByRoot bool
	Offset       int
}

// SearchResponse is the paginated search result.
type SearchResponse struct {
	Phrases []domain.Phrase `json:"phrases"`
	Total   int             `json:"total"`
	Offset  int             `json:"offset"`
	Limit   int             `json:"limit"`
}

// Service defines the search business logic interface.
type Service interface {
	Search(ctx context.Context, req SearchRequest) (*SearchResponse, error)
}

type service struct {
	repo Repository
}

// NewService creates a new search service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// Search executes a full-text search over approved phrases.
func (s *service) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	if req.Query == "" {
		return &SearchResponse{Phrases: []domain.Phrase{}, Total: 0, Offset: req.Offset, Limit: 50}, nil
	}
	if len(req.Query) > 100 {
		req.Query = req.Query[:100]
	}

	phrases, total, err := s.repo.Search(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	return &SearchResponse{
		Phrases: phrases,
		Total:   total,
		Offset:  req.Offset,
		Limit:   50,
	}, nil
}