package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/yourusername/bahasa-daerah-platform/internal/auth"
	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
)

const testJWTSecret = "changeme"

// newTestRouter wires up a chi router with the auth routes for testing.
func newTestRouter(repo auth.Repository) *chi.Mux {
	svc := auth.NewService(repo)
	h := auth.NewHandler(svc)
	r := chi.NewRouter()
	r.Route("/api/v1/auth", h.Routes(testJWTSecret))
	return r
}

// postJSON sends a POST request with a JSON body and returns the recorder.
func postJSON(t *testing.T, router http.Handler, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

// postJSONWithToken sends a POST request with a JWT Bearer token.
func postJSONWithToken(t *testing.T, router http.Handler, path, token string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

// makeJWT creates a signed JWT for testing purposes.
func makeJWT(t *testing.T, userID uuid.UUID, role domain.Role, expiry time.Duration) string {
	t.Helper()
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role":    string(role),
		"exp":     time.Now().Add(expiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("sign JWT: %v", err)
	}
	return signed
}

// ── Register handler tests ────────────────────────────────────────────────────

// TestHandlerRegister_Success verifies that a valid registration returns 201
// with a token and user in the body.
// Requirements: 1.1, 1.2
func TestHandlerRegister_Success(t *testing.T) {
	router := newTestRouter(newMockRepo())

	rr := postJSON(t, router, "/api/v1/auth/register", map[string]string{
		"name":     "Budi Santoso",
		"email":    "budi@example.com",
		"password": "password123",
		"role":     "learner",
	})

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d — body: %s", rr.Code, rr.Body.String())
	}

	var res map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if res["token"] == "" || res["token"] == nil {
		t.Error("expected non-empty token in response")
	}
	user, ok := res["user"].(map[string]interface{})
	if !ok {
		t.Fatal("expected user object in response")
	}
	if user["email"] != "budi@example.com" {
		t.Errorf("email mismatch: got %v", user["email"])
	}
}

// TestHandlerRegister_DuplicateEmail verifies that registering with an existing
// email returns HTTP 409.
// Requirements: 1.3
func TestHandlerRegister_DuplicateEmail(t *testing.T) {
	router := newTestRouter(newMockRepo())

	body := map[string]string{
		"name":     "Budi Santoso",
		"email":    "budi@example.com",
		"password": "password123",
		"role":     "learner",
	}

	postJSON(t, router, "/api/v1/auth/register", body) // first registration

	rr := postJSON(t, router, "/api/v1/auth/register", body) // duplicate
	if rr.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d — body: %s", rr.Code, rr.Body.String())
	}
}

// TestHandlerRegister_InvalidRole verifies that an invalid role returns 400.
// Requirements: 1.1
func TestHandlerRegister_InvalidRole(t *testing.T) {
	router := newTestRouter(newMockRepo())

	rr := postJSON(t, router, "/api/v1/auth/register", map[string]string{
		"name":     "Budi Santoso",
		"email":    "budi@example.com",
		"password": "password123",
		"role":     "admin", // not allowed on register
	})

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d — body: %s", rr.Code, rr.Body.String())
	}
}

// TestHandlerRegister_MissingFields verifies that missing required fields
// return 400.
// Requirements: 1.1
func TestHandlerRegister_MissingFields(t *testing.T) {
	router := newTestRouter(newMockRepo())

	cases := []struct {
		name string
		body map[string]string
	}{
		{"missing name", map[string]string{"email": "a@b.com", "password": "password123", "role": "learner"}},
		{"missing email", map[string]string{"name": "Budi", "password": "password123", "role": "learner"}},
		{"missing password", map[string]string{"name": "Budi", "email": "a@b.com", "role": "learner"}},
		{"short password", map[string]string{"name": "Budi", "email": "a@b.com", "password": "short", "role": "learner"}},
		{"invalid email", map[string]string{"name": "Budi", "email": "notanemail", "password": "password123", "role": "learner"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := postJSON(t, router, "/api/v1/auth/register", tc.body)
			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d — body: %s", rr.Code, rr.Body.String())
			}
		})
	}
}

// ── Login handler tests ───────────────────────────────────────────────────────

// TestHandlerLogin_Success verifies that valid credentials return 200 with a
// token.
// Requirements: 1.4
func TestHandlerLogin_Success(t *testing.T) {
	repo := newMockRepo()
	router := newTestRouter(repo)

	// Register first
	postJSON(t, router, "/api/v1/auth/register", map[string]string{
		"name":     "Budi Santoso",
		"email":    "budi@example.com",
		"password": "password123",
		"role":     "learner",
	})

	rr := postJSON(t, router, "/api/v1/auth/login", map[string]string{
		"email":    "budi@example.com",
		"password": "password123",
	})

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", rr.Code, rr.Body.String())
	}

	var res map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if res["token"] == "" || res["token"] == nil {
		t.Error("expected non-empty token in response")
	}
}

// TestHandlerLogin_WrongPassword verifies that wrong credentials return 401.
// Requirements: 1.5
func TestHandlerLogin_WrongPassword(t *testing.T) {
	repo := newMockRepo()
	router := newTestRouter(repo)

	postJSON(t, router, "/api/v1/auth/register", map[string]string{
		"name":     "Budi Santoso",
		"email":    "budi@example.com",
		"password": "password123",
		"role":     "learner",
	})

	rr := postJSON(t, router, "/api/v1/auth/login", map[string]string{
		"email":    "budi@example.com",
		"password": "wrongpassword",
	})

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d — body: %s", rr.Code, rr.Body.String())
	}
}

// TestHandlerLogin_UnknownEmail verifies that an unknown email returns 401
// (not 404, to avoid leaking user existence).
// Requirements: 1.5
func TestHandlerLogin_UnknownEmail(t *testing.T) {
	router := newTestRouter(newMockRepo())

	rr := postJSON(t, router, "/api/v1/auth/login", map[string]string{
		"email":    "nobody@example.com",
		"password": "password123",
	})

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d — body: %s", rr.Code, rr.Body.String())
	}
}

// TestHandlerLogin_MissingFields verifies that missing email or password
// returns 400.
// Requirements: 1.4
func TestHandlerLogin_MissingFields(t *testing.T) {
	router := newTestRouter(newMockRepo())

	cases := []struct {
		name string
		body map[string]string
	}{
		{"missing email", map[string]string{"password": "password123"}},
		{"missing password", map[string]string{"email": "a@b.com"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := postJSON(t, router, "/api/v1/auth/login", tc.body)
			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d — body: %s", rr.Code, rr.Body.String())
			}
		})
	}
}

// ── JWT Middleware tests ──────────────────────────────────────────────────────

// TestMiddleware_NoToken verifies that a protected endpoint returns 401 when
// no token is provided.
// Requirements: 1.4, 16.5
func TestMiddleware_NoToken(t *testing.T) {
	repo := newMockRepo()
	// Register a learner so the upgrade endpoint has a user to work with
	svc := auth.NewService(repo)
	regRes, err := svc.Register(context.Background(), auth.RegisterRequest{
		Name:     "Budi",
		Email:    "budi@example.com",
		Password: "password123",
		Role:     domain.RoleLearner,
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	_ = regRes

	router := newTestRouter(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/upgrade-role", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

// TestMiddleware_InvalidToken verifies that a malformed token returns 401.
// Requirements: 1.4, 16.5
func TestMiddleware_InvalidToken(t *testing.T) {
	router := newTestRouter(newMockRepo())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/upgrade-role", nil)
	req.Header.Set("Authorization", "Bearer this.is.not.a.valid.jwt")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

// TestMiddleware_ExpiredToken verifies that an expired token returns 401.
// Requirements: 1.4, 16.5
func TestMiddleware_ExpiredToken(t *testing.T) {
	repo := newMockRepo()
	router := newTestRouter(repo)

	userID := uuid.New()
	expiredToken := makeJWT(t, userID, domain.RoleLearner, -1*time.Hour) // expired 1h ago

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/upgrade-role", bytes.NewReader([]byte("{}")))
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for expired token, got %d", rr.Code)
	}
}

// ── UpgradeRole handler tests ─────────────────────────────────────────────────

// TestHandlerUpgradeRole_Success verifies that a learner can upgrade to
// contributor and receives a new JWT.
// Requirements: 23.3, 23.4, 23.6
func TestHandlerUpgradeRole_Success(t *testing.T) {
	repo := newMockRepo()
	router := newTestRouter(repo)

	// Register a learner
	rr := postJSON(t, router, "/api/v1/auth/register", map[string]string{
		"name":     "Budi Santoso",
		"email":    "budi@example.com",
		"password": "password123",
		"role":     "learner",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("register failed: %d — %s", rr.Code, rr.Body.String())
	}

	var regRes map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&regRes)
	token := regRes["token"].(string)

	// Upgrade role
	upgradeRR := postJSONWithToken(t, router, "/api/v1/auth/upgrade-role", token, nil)
	if upgradeRR.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", upgradeRR.Code, upgradeRR.Body.String())
	}

	var upgradeRes map[string]interface{}
	json.NewDecoder(upgradeRR.Body).Decode(&upgradeRes)

	newToken, ok := upgradeRes["token"].(string)
	if !ok || newToken == "" {
		t.Fatal("expected non-empty token in upgrade response")
	}

	// New JWT must have contributor role
	claims := parseJWTClaims(t, newToken)
	if claims["role"] != string(domain.RoleContributor) {
		t.Errorf("expected contributor role in new JWT, got %v", claims["role"])
	}

	// User in response must have contributor role
	user, ok := upgradeRes["user"].(map[string]interface{})
	if !ok {
		t.Fatal("expected user in upgrade response")
	}
	if user["role"] != string(domain.RoleContributor) {
		t.Errorf("expected contributor role in user, got %v", user["role"])
	}
}

// TestHandlerUpgradeRole_AlreadyContributor verifies that upgrading a
// contributor returns 409.
// Requirements: 23.6
func TestHandlerUpgradeRole_AlreadyContributor(t *testing.T) {
	repo := newMockRepo()
	router := newTestRouter(repo)

	// Register as contributor
	rr := postJSON(t, router, "/api/v1/auth/register", map[string]string{
		"name":     "Sari Dewi",
		"email":    "sari@example.com",
		"password": "password123",
		"role":     "contributor",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("register failed: %d", rr.Code)
	}

	var regRes map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&regRes)
	token := regRes["token"].(string)

	upgradeRR := postJSONWithToken(t, router, "/api/v1/auth/upgrade-role", token, nil)
	if upgradeRR.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d — body: %s", upgradeRR.Code, upgradeRR.Body.String())
	}
}

// TestHandlerUpgradeRole_RequiresAuth verifies that the upgrade endpoint
// requires authentication.
// Requirements: 23.3
func TestHandlerUpgradeRole_RequiresAuth(t *testing.T) {
	router := newTestRouter(newMockRepo())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/upgrade-role", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}
