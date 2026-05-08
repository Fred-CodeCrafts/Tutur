package admin

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

// Handler holds HTTP handlers for admin routes.
type Handler struct {
	svc Service
}

// NewHandler creates a new admin handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Routes mounts all admin routes — all protected by admin role.
func (h *Handler) Routes(jwtSecret string) func(r chi.Router) {
	return func(r chi.Router) {
		r.Use(middleware.Authenticate(jwtSecret))
		r.Use(middleware.RequireRole(domain.RoleAdmin))

		// Phrase moderation
		r.Get("/phrases/flagged", h.ListFlaggedPhrases)
		r.Patch("/phrases/{id}/status", h.ModeratePhrase)
		r.Delete("/phrases/{id}", h.DeletePhrase)

		// User management
		r.Get("/users", h.ListUsers)
		r.Patch("/users/{id}/ban", h.BanUser)
		r.Patch("/users/{id}/role", h.AssignRole)
	}
}

// ListFlaggedPhrases handles GET /api/v1/admin/phrases/flagged
func (h *Handler) ListFlaggedPhrases(w http.ResponseWriter, r *http.Request) {
	phrases, err := h.svc.ListFlaggedPhrases(r.Context())
	if err != nil {
		response.InternalServerError(w)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"phrases": phrases})
}

// ModeratePhrase handles PATCH /api/v1/admin/phrases/:id/status
func (h *Handler) ModeratePhrase(w http.ResponseWriter, r *http.Request) {
	phraseID, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	adminID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	var req struct {
		Action string `json:"action"` // "approve" or "reject"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_JSON", "Invalid request body.")
		return
	}

	if err := h.svc.ModeratePhrase(r.Context(), phraseID, adminID, req.Action); err != nil {
		response.BadRequest(w, "INVALID_ACTION", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Phrase status updated."})
}

// DeletePhrase handles DELETE /api/v1/admin/phrases/:id
func (h *Handler) DeletePhrase(w http.ResponseWriter, r *http.Request) {
	phraseID, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}

	if err := h.svc.DeletePhrase(r.Context(), phraseID); err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(w, "Phrase not found.")
			return
		}
		response.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListUsers handles GET /api/v1/admin/users
// Query param: search (optional)
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	users, err := h.svc.ListUsers(r.Context(), search)
	if err != nil {
		response.InternalServerError(w)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"users": users})
}

// BanUser handles PATCH /api/v1/admin/users/:id/ban
func (h *Handler) BanUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}

	if err := h.svc.BanUser(r.Context(), userID); err != nil {
		response.InternalServerError(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "User banned and pending phrases rejected."})
}

// AssignRole handles PATCH /api/v1/admin/users/:id/role
func (h *Handler) AssignRole(w http.ResponseWriter, r *http.Request) {
	targetID, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	adminID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_JSON", "Invalid request body.")
		return
	}

	if err := h.svc.AssignRole(r.Context(), targetID, adminID, domain.Role(req.Role)); err != nil {
		response.BadRequest(w, "INVALID_ROLE", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Role assigned."})
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func parseUUID(w http.ResponseWriter, r *http.Request, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, param))
	if err != nil {
		response.BadRequest(w, "INVALID_ID", param+" must be a valid UUID.")
		return uuid.Nil, false
	}
	return id, true
}