package validation

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
	"github.com/yourusername/bahasa-daerah-platform/pkg/middleware"
	"github.com/yourusername/bahasa-daerah-platform/pkg/response"
)

// Handler holds the HTTP handlers for validation routes (votes and flags).
type Handler struct {
	svc Service
}

// NewHandler creates a new validation handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Routes mounts validation sub-routes under a phrase router.
// All routes require authentication. Votes require contributor role;
// flags are accessible to learners and contributors.
func (h *Handler) Routes(jwtSecret string) func(r chi.Router) {
	return func(r chi.Router) {
		r.Use(middleware.Authenticate(jwtSecret))

		// POST /api/v1/phrases/:id/votes — text vote (Contributor only)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole(domain.RoleContributor, domain.RoleAdmin))
			r.Post("/{id}/votes", h.VotePhrase)
			r.Post("/{id}/audio-votes", h.VoteAudio)
			r.Post("/{id}/script-votes", h.VoteScript)
		})

		// POST /api/v1/phrases/:id/flags — flag (Learner, Contributor, Admin)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole(domain.RoleLearner, domain.RoleContributor, domain.RoleAdmin))
			r.Post("/{id}/flags", h.FlagPhrase)
		})
	}
}

// VotePhrase handles POST /api/v1/phrases/:id/votes
func (h *Handler) VotePhrase(w http.ResponseWriter, r *http.Request) {
	phraseID, ok := parsePhraseID(w, r)
	if !ok {
		return
	}

	contributorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	var req VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_JSON", "Invalid request body.")
		return
	}

	if err := h.svc.VotePhrase(r.Context(), phraseID, contributorID, req); err != nil {
		handleVoteError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, map[string]string{"message": "Vote recorded."})
}

// FlagPhrase handles POST /api/v1/phrases/:id/flags
func (h *Handler) FlagPhrase(w http.ResponseWriter, r *http.Request) {
	phraseID, ok := parsePhraseID(w, r)
	if !ok {
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	var req FlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_JSON", "Invalid request body.")
		return
	}

	if err := h.svc.FlagPhrase(r.Context(), phraseID, userID, req); err != nil {
		handleFlagError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, map[string]string{"message": "Flag recorded."})
}

// VoteAudio handles POST /api/v1/phrases/:id/audio-votes
func (h *Handler) VoteAudio(w http.ResponseWriter, r *http.Request) {
	phraseID, ok := parsePhraseID(w, r)
	if !ok {
		return
	}

	contributorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	var req VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_JSON", "Invalid request body.")
		return
	}

	if err := h.svc.VoteAudio(r.Context(), phraseID, contributorID, req); err != nil {
		handleVoteError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, map[string]string{"message": "Audio vote recorded."})
}

// VoteScript handles POST /api/v1/phrases/:id/script-votes
func (h *Handler) VoteScript(w http.ResponseWriter, r *http.Request) {
	phraseID, ok := parsePhraseID(w, r)
	if !ok {
		return
	}

	contributorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	var req VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_JSON", "Invalid request body.")
		return
	}

	if err := h.svc.VoteScript(r.Context(), phraseID, contributorID, req); err != nil {
		handleVoteError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, map[string]string{"message": "Script vote recorded."})
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func parsePhraseID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "INVALID_ID", "Phrase ID must be a valid UUID.")
		return uuid.Nil, false
	}
	return id, true
}

func handleVoteError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrPhraseNotFound):
		response.NotFound(w, "Phrase not found.")
	case errors.Is(err, ErrSelfVote):
		response.Forbidden(w, "You cannot vote on your own phrase.")
	case errors.Is(err, ErrDuplicateVote):
		response.Conflict(w, "DUPLICATE_VOTE", "You have already voted on this phrase.")
	case errors.Is(err, ErrInvalidVoteType):
		response.BadRequest(w, "INVALID_VOTE_TYPE", err.Error())
	default:
		response.InternalServerError(w)
	}
}

func handleFlagError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrPhraseNotFound):
		response.NotFound(w, "Phrase not found.")
	case errors.Is(err, ErrDuplicateVote):
		response.Conflict(w, "DUPLICATE_FLAG", "You have already flagged this phrase.")
	case errors.Is(err, ErrInvalidFlagReason):
		response.BadRequest(w, "INVALID_FLAG_REASON", err.Error())
	default:
		response.InternalServerError(w)
	}
}
