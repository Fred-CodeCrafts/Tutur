package flashcard

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
	"github.com/yourusername/bahasa-daerah-platform/pkg/middleware"
	"github.com/yourusername/bahasa-daerah-platform/pkg/response"
)

// Handler holds HTTP handlers for flashcard and practice routes.
type Handler struct {
	svc Service
}

// NewHandler creates a new flashcard handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Routes mounts all flashcard-related routes.
func (h *Handler) Routes(jwtSecret string) func(r chi.Router) {
	return func(r chi.Router) {
		r.Use(middleware.Authenticate(jwtSecret))
		r.Use(middleware.RequireRole(domain.RoleLearner, domain.RoleContributor, domain.RoleAdmin))

		r.Get("/flashcards", h.ListFlashcards)
		r.Get("/conversation-scenarios", h.GetConversationScenario)
		r.Post("/phrase-practice", h.SavePracticeResult)
	}
}

// ListFlashcards handles GET /api/v1/flashcards
// Query params: language_code (required), tone (optional), cursor (optional)
func (h *Handler) ListFlashcards(w http.ResponseWriter, r *http.Request) {
	languageCode := r.URL.Query().Get("language_code")
	if languageCode == "" {
		response.BadRequest(w, "MISSING_LANGUAGE", "language_code query parameter is required.")
		return
	}

	filter := FlashcardFilter{
		LanguageCode: languageCode,
		Tone:         domain.Tone(r.URL.Query().Get("tone")),
		Cursor:       r.URL.Query().Get("cursor"),
	}

	page, err := h.svc.ListFlashcards(r.Context(), filter)
	if err != nil {
		response.InternalServerError(w)
		return
	}

	response.JSON(w, http.StatusOK, page)
}

// GetConversationScenario handles GET /api/v1/conversation-scenarios
// Query params: language_code (required)
func (h *Handler) GetConversationScenario(w http.ResponseWriter, r *http.Request) {
	languageCode := r.URL.Query().Get("language_code")
	if languageCode == "" {
		response.BadRequest(w, "MISSING_LANGUAGE", "language_code query parameter is required.")
		return
	}

	scenario, err := h.svc.GetConversationScenario(r.Context(), languageCode)
	if err != nil {
		if errors.Is(err, ErrInsufficientContent) {
			response.NotFound(w, "Belum ada konten yang cukup untuk bahasa ini. Minimal 3 frasa disetujui diperlukan.")
			return
		}
		response.InternalServerError(w)
		return
	}

	response.JSON(w, http.StatusOK, scenario)
}

// SavePracticeResult handles POST /api/v1/phrase-practice
func (h *Handler) SavePracticeResult(w http.ResponseWriter, r *http.Request) {
	learnerID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	var req PracticeResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_JSON", "Invalid request body.")
		return
	}

	if err := h.svc.SavePracticeResult(r.Context(), learnerID, req); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, map[string]string{"message": "Practice result recorded."})
}