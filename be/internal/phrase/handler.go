package phrase

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
	"github.com/yourusername/bahasa-daerah-platform/pkg/middleware"
	"github.com/yourusername/bahasa-daerah-platform/pkg/response"
	"github.com/yourusername/bahasa-daerah-platform/pkg/validator"
)

// validScriptTypes is the set of accepted script_type values.
var validScriptTypes = map[domain.ScriptType]struct{}{
	domain.ScriptLatin:     {},
	domain.ScriptJavanese:  {},
	domain.ScriptSundanese: {},
	domain.ScriptBalinese:  {},
	domain.ScriptLontara:   {},
	domain.ScriptBatak:     {},
	domain.ScriptOther:     {},
}

// Handler holds the HTTP handlers for phrase routes.
type Handler struct {
	svc Service
}

// NewHandler creates a new phrase handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Routes mounts phrase routes. All routes require authentication.
// Contributor-only routes are further protected by role middleware.
func (h *Handler) Routes(jwtSecret string) func(r chi.Router) {
	return func(r chi.Router) {
		r.Use(middleware.Authenticate(jwtSecret))

		// GET /api/v1/phrases — list pending phrases (Contributor)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole(domain.RoleContributor, domain.RoleAdmin))
			r.Get("/", h.ListPendingPhrases)
			r.Post("/", h.SubmitPhrase)
		})

		// GET /api/v1/phrases/my — phrases by the logged-in contributor
		// Must be registered before /:id to avoid chi routing conflict
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole(domain.RoleContributor, domain.RoleAdmin))
			r.Get("/my", h.ListMyPhrases)
		})

		// GET /api/v1/phrases/:id — phrase detail (any authenticated user)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole(domain.RoleLearner, domain.RoleContributor, domain.RoleAdmin))
			r.Get("/{id}", h.GetPhraseByID)
		})
	}
}

// SubmitPhrase handles POST /api/v1/phrases
func (h *Handler) SubmitPhrase(w http.ResponseWriter, r *http.Request) {
	contributorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	var req SubmitPhraseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_JSON", "Invalid request body.")
		return
	}

	// ── Validation ────────────────────────────────────────────────────────────
	v := validator.New()

	// text_latin: required, max 500 chars
	trimmedLatin := strings.TrimSpace(req.TextLatin)
	v.Check(trimmedLatin != "", "text_latin", "text_latin is required.")
	if trimmedLatin != "" {
		v.Check(len(trimmedLatin) <= 500, "text_latin", "text_latin must not exceed 500 characters.")
	}

	// translation: required
	v.Check(strings.TrimSpace(req.Translation) != "", "translation", "translation is required.")

	// language_code: required
	v.Check(strings.TrimSpace(req.LanguageCode) != "", "language_code", "language_code is required.")

	// If text_native_script is provided, script_type is required and must be valid
	if req.TextNativeScript != nil && strings.TrimSpace(*req.TextNativeScript) != "" {
		v.Check(req.ScriptType != nil, "script_type", "script_type is required when text_native_script is provided.")
		if req.ScriptType != nil {
			_, validScript := validScriptTypes[*req.ScriptType]
			v.Check(validScript, "script_type", "script_type must be one of: latin, javanese, sundanese, balinese, lontara, batak, other.")
		}
	}

	if !v.Valid() {
		// Distinguish between missing required fields (400) and constraint violations (422)
		errs := v.Errors()
		if _, hasLatin := errs["text_latin"]; hasLatin && trimmedLatin == "" {
			response.BadRequest(w, "VALIDATION_ERROR", v.Err().Error())
			return
		}
		if _, hasTranslation := errs["translation"]; hasTranslation {
			response.BadRequest(w, "VALIDATION_ERROR", v.Err().Error())
			return
		}
		// Length / enum violations → 422
		response.UnprocessableEntity(w, "VALIDATION_ERROR", v.Err().Error())
		return
	}

	res, err := h.svc.SubmitPhrase(r.Context(), contributorID, req)
	if err != nil {
		if errors.Is(err, ErrInactiveLanguage) {
			response.BadRequest(w, "INACTIVE_LANGUAGE", "The specified language_code is not active or does not exist.")
			return
		}
		response.InternalServerError(w)
		return
	}

	response.JSON(w, http.StatusCreated, res)
}

// ListPendingPhrases handles GET /api/v1/phrases
func (h *Handler) ListPendingPhrases(w http.ResponseWriter, r *http.Request) {
	phrases, err := h.svc.ListPendingPhrases(r.Context())
	if err != nil {
		response.InternalServerError(w)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"phrases": phrases})
}

// GetPhraseByID handles GET /api/v1/phrases/:id
func (h *Handler) GetPhraseByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "INVALID_ID", "Phrase ID must be a valid UUID.")
		return
	}

	p, err := h.svc.GetPhraseByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "Phrase not found.")
			return
		}
		response.InternalServerError(w)
		return
	}

	response.JSON(w, http.StatusOK, p)
}

// ListMyPhrases handles GET /api/v1/phrases/my
func (h *Handler) ListMyPhrases(w http.ResponseWriter, r *http.Request) {
	contributorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	phrases, err := h.svc.ListMyPhrases(r.Context(), contributorID)
	if err != nil {
		response.InternalServerError(w)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"phrases": phrases})
}
