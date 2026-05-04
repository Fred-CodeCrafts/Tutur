package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
	"github.com/yourusername/bahasa-daerah-platform/pkg/middleware"
)

const testSecret = "test-secret"

func makeToken(t *testing.T, userID uuid.UUID, role domain.Role, expiry time.Duration) string {
	t.Helper()
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role":    string(role),
		"exp":     time.Now().Add(expiry).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

// nextHandler is a simple handler that records whether it was called and
// captures the context values injected by the middleware.
type nextHandler struct {
	called bool
	userID uuid.UUID
	role   domain.Role
}

func (h *nextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.called = true
	h.userID, _ = middleware.UserIDFromContext(r.Context())
	h.role, _ = middleware.RoleFromContext(r.Context())
	w.WriteHeader(http.StatusOK)
}

// ── Authenticate middleware tests ─────────────────────────────────────────────

// TestAuthenticate_ValidToken verifies that a valid JWT passes through and
// injects user_id and role into the context.
// Requirements: 1.4, 16.5
func TestAuthenticate_ValidToken(t *testing.T) {
	userID := uuid.New()
	token := makeToken(t, userID, domain.RoleLearner, time.Hour)

	next := &nextHandler{}
	handler := middleware.Authenticate(testSecret)(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if !next.called {
		t.Error("expected next handler to be called")
	}
	if next.userID != userID {
		t.Errorf("userID mismatch: got %v, want %v", next.userID, userID)
	}
	if next.role != domain.RoleLearner {
		t.Errorf("role mismatch: got %v, want %v", next.role, domain.RoleLearner)
	}
}

// TestAuthenticate_MissingHeader verifies that a missing Authorization header
// returns 401.
// Requirements: 16.5
func TestAuthenticate_MissingHeader(t *testing.T) {
	next := &nextHandler{}
	handler := middleware.Authenticate(testSecret)(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
	if next.called {
		t.Error("next handler should not be called")
	}
}

// TestAuthenticate_InvalidToken verifies that a malformed token returns 401.
// Requirements: 16.5
func TestAuthenticate_InvalidToken(t *testing.T) {
	next := &nextHandler{}
	handler := middleware.Authenticate(testSecret)(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

// TestAuthenticate_ExpiredToken verifies that an expired token returns 401.
// Requirements: 16.5
func TestAuthenticate_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	token := makeToken(t, userID, domain.RoleLearner, -time.Hour) // expired

	next := &nextHandler{}
	handler := middleware.Authenticate(testSecret)(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for expired token, got %d", rr.Code)
	}
}

// TestAuthenticate_WrongSecret verifies that a token signed with a different
// secret returns 401.
// Requirements: 16.5
func TestAuthenticate_WrongSecret(t *testing.T) {
	userID := uuid.New()
	// Sign with a different secret
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role":    string(domain.RoleLearner),
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, _ := tok.SignedString([]byte("wrong-secret"))

	next := &nextHandler{}
	handler := middleware.Authenticate(testSecret)(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong secret, got %d", rr.Code)
	}
}

// ── RequireRole middleware tests ──────────────────────────────────────────────

// TestRequireRole_AllowedRole verifies that a user with the correct role passes
// through.
// Requirements: 16.5
func TestRequireRole_AllowedRole(t *testing.T) {
	userID := uuid.New()
	token := makeToken(t, userID, domain.RoleContributor, time.Hour)

	next := &nextHandler{}
	handler := middleware.Authenticate(testSecret)(
		middleware.RequireRole(domain.RoleContributor)(next),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if !next.called {
		t.Error("expected next handler to be called")
	}
}

// TestRequireRole_ForbiddenRole verifies that a user with the wrong role gets
// 403.
// Requirements: 16.5
func TestRequireRole_ForbiddenRole(t *testing.T) {
	userID := uuid.New()
	token := makeToken(t, userID, domain.RoleLearner, time.Hour)

	next := &nextHandler{}
	handler := middleware.Authenticate(testSecret)(
		middleware.RequireRole(domain.RoleContributor)(next),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
	if next.called {
		t.Error("next handler should not be called")
	}
}

// TestRequireRole_MultipleAllowedRoles verifies that RequireRole accepts
// multiple roles.
// Requirements: 16.5
func TestRequireRole_MultipleAllowedRoles(t *testing.T) {
	cases := []struct {
		role    domain.Role
		allowed bool
	}{
		{domain.RoleLearner, true},
		{domain.RoleContributor, true},
		{domain.RoleAdmin, false},
	}

	for _, tc := range cases {
		t.Run(string(tc.role), func(t *testing.T) {
			userID := uuid.New()
			token := makeToken(t, userID, tc.role, time.Hour)

			next := &nextHandler{}
			handler := middleware.Authenticate(testSecret)(
				middleware.RequireRole(domain.RoleLearner, domain.RoleContributor)(next),
			)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if tc.allowed && rr.Code != http.StatusOK {
				t.Errorf("expected 200 for role %v, got %d", tc.role, rr.Code)
			}
			if !tc.allowed && rr.Code != http.StatusForbidden {
				t.Errorf("expected 403 for role %v, got %d", tc.role, rr.Code)
			}
		})
	}
}
