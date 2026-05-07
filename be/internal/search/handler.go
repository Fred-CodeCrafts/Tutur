package search

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
	"github.com/yourusername/bahasa-daerah-platform/pkg/middleware"
	"github.com/yourusername/bahasa-daerah-platform/pkg/response"
)

// Handler holds HTTP handlers for search routes.
type Handler struct {
	svc Service
}

// NewHandler creates a new search handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Routes mounts the search route.
func (h *Handler) Routes(jwtSecret string) func(r chi.Router) {
	return func(r chi.Router) {
		r.Use(middleware.Authenticate(jwtSecret))
		r.Use(middleware.RequireRole(domain.RoleLearner, domain.RoleContributor, domain.RoleAdmin))
		r.Get("/search", h.Search)
	}
}

// Search handles GET /api/v1/search
// Query params: q (required), language_code (required), root (optional bool), offset (optional)
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		response.BadRequest(w, "MISSING_QUERY", "q query parameter is required.")
		return
	}
	languageCode := r.URL.Query().Get("language_code")
	if languageCode == "" {
		response.BadRequest(w, "MISSING_LANGUAGE", "language_code query parameter is required.")
		return
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	searchByRoot := r.URL.Query().Get("root") == "true"

	req := SearchRequest{
		Query:        q,
		LanguageCode: languageCode,
		SearchByRoot: searchByRoot,
		Offset:       offset,
	}

	res, err := h.svc.Search(r.Context(), req)
	if err != nil {
		response.InternalServerError(w)
		return
	}

	response.JSON(w, http.StatusOK, res)
}