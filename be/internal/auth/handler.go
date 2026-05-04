package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
	"github.com/yourusername/bahasa-daerah-platform/pkg/middleware"
	"github.com/yourusername/bahasa-daerah-platform/pkg/response"
	"github.com/yourusername/bahasa-daerah-platform/pkg/validator"
)

var emailRegex = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// Handler holds the HTTP handlers for auth routes.
type Handler struct {
	svc Service
}

// NewHandler creates a new auth handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Routes mounts the auth routes onto the given chi router.
func (h *Handler) Routes(jwtSecret string) func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(jwtSecret))
			// Any authenticated user can call this endpoint; the service returns
			// 409 if the role is already contributor or admin.
			r.Use(middleware.RequireRole(domain.RoleLearner, domain.RoleContributor, domain.RoleAdmin))
			r.Post("/upgrade-role", h.UpgradeRole)
		})
	}
}

// Register handles POST /api/v1/auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_JSON", "Invalid request body.")
		return
	}

	v := validator.New()
	v.Check(strings.TrimSpace(req.Name) != "", "name", "Name is required.")
	v.Check(strings.TrimSpace(req.Email) != "", "email", "Email is required.")
	v.Check(emailRegex.MatchString(req.Email), "email", "Email must be a valid email address.")
	v.Check(len(req.Password) >= 8, "password", "Password must be at least 8 characters.")
	v.Check(req.Role == domain.RoleLearner || req.Role == domain.RoleContributor, "role", "Role must be 'learner' or 'contributor'.")

	if !v.Valid() {
		var ve *validator.ValidationError
		if err := v.Err(); errors.As(err, &ve) {
			response.BadRequest(w, "VALIDATION_ERROR", ve.Error())
		}
		return
	}

	res, err := h.svc.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrDuplicateEmail) {
			response.Conflict(w, "DUPLICATE_EMAIL", "A user with this email already exists.")
			return
		}
		response.InternalServerError(w)
		return
	}

	response.JSON(w, http.StatusCreated, res)
}

// Login handles POST /api/v1/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_JSON", "Invalid request body.")
		return
	}

	v := validator.New()
	v.Check(strings.TrimSpace(req.Email) != "", "email", "Email is required.")
	v.Check(strings.TrimSpace(req.Password) != "", "password", "Password is required.")

	if !v.Valid() {
		var ve *validator.ValidationError
		if err := v.Err(); errors.As(err, &ve) {
			response.BadRequest(w, "VALIDATION_ERROR", ve.Error())
		}
		return
	}

	res, err := h.svc.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Unauthorized(w)
			return
		}
		response.InternalServerError(w)
		return
	}

	response.JSON(w, http.StatusOK, res)
}

// UpgradeRole handles POST /api/v1/auth/upgrade-role
func (h *Handler) UpgradeRole(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	res, err := h.svc.UpgradeRole(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrRoleAlreadyUpgraded) {
			response.Conflict(w, "ROLE_ALREADY_UPGRADED", "Role has already been upgraded.")
			return
		}
		if errors.Is(err, ErrNotFound) {
			response.Unauthorized(w)
			return
		}
		response.InternalServerError(w)
		return
	}

	response.JSON(w, http.StatusOK, res)
}
