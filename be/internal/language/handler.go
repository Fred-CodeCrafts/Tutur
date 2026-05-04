package language

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
	"github.com/yourusername/bahasa-daerah-platform/pkg/middleware"
	"github.com/yourusername/bahasa-daerah-platform/pkg/response"
	"github.com/yourusername/bahasa-daerah-platform/pkg/validator"
)

// Handler holds the HTTP handlers for language routes.
type Handler struct {
	svc Service
}

// NewHandler creates a new language handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// PublicRoutes mounts public language routes (no auth required).
func (h *Handler) PublicRoutes() func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", h.ListLanguages)
	}
}

// AdminRoutes mounts admin-only language routes.
func (h *Handler) AdminRoutes(jwtSecret string) func(r chi.Router) {
	return func(r chi.Router) {
		r.Use(middleware.Authenticate(jwtSecret))
		r.Use(middleware.RequireRole(domain.RoleAdmin))
		r.Post("/", h.CreateLanguage)
		r.Patch("/{code}", h.ToggleActive)
	}
}

// ListLanguages handles GET /api/v1/languages
func (h *Handler) ListLanguages(w http.ResponseWriter, r *http.Request) {
	langs, err := h.svc.ListLanguages(r.Context())
	if err != nil {
		response.InternalServerError(w)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"languages": langs})
}

// CreateLanguage handles POST /api/v1/admin/languages
func (h *Handler) CreateLanguage(w http.ResponseWriter, r *http.Request) {
	var req CreateLanguageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_JSON", "Invalid request body.")
		return
	}

	v := validator.New()
	v.Check(strings.TrimSpace(req.Code) != "", "code", "Language code is required.")
	v.Check(len(strings.TrimSpace(req.Code)) <= 20, "code", "Language code must be at most 20 characters.")
	v.Check(strings.TrimSpace(req.Name) != "", "name", "Language name is required.")

	if !v.Valid() {
		response.BadRequest(w, "VALIDATION_ERROR", v.Err().Error())
		return
	}

	lang, err := h.svc.CreateLanguage(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrDuplicateCode) {
			response.Conflict(w, "DUPLICATE_CODE", "A language with this code already exists.")
			return
		}
		response.InternalServerError(w)
		return
	}

	response.JSON(w, http.StatusCreated, lang)
}

// ToggleActive handles PATCH /api/v1/admin/languages/:code
func (h *Handler) ToggleActive(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		response.BadRequest(w, "MISSING_CODE", "Language code is required.")
		return
	}

	var req ToggleActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_JSON", "Invalid request body.")
		return
	}

	lang, err := h.svc.ToggleActive(r.Context(), code, req.IsActive)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "Language not found.")
			return
		}
		response.InternalServerError(w)
		return
	}

	response.JSON(w, http.StatusOK, lang)
}
