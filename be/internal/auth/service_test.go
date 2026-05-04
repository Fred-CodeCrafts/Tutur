package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/yourusername/bahasa-daerah-platform/internal/auth"
	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
)

// ── Mock Repository ───────────────────────────────────────────────────────────

type mockRepo struct {
	users map[string]*domain.User // keyed by email
	byID  map[uuid.UUID]*domain.User
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		users: make(map[string]*domain.User),
		byID:  make(map[uuid.UUID]*domain.User),
	}
}

func (m *mockRepo) CreateUser(_ context.Context, user *domain.User) error {
	if _, exists := m.users[user.Email]; exists {
		return auth.ErrDuplicateEmail
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.Email] = user
	m.byID[user.ID] = user
	return nil
}

func (m *mockRepo) GetUserByEmail(_ context.Context, email string) (*domain.User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, auth.ErrNotFound
	}
	return u, nil
}

func (m *mockRepo) GetUserByID(_ context.Context, userID uuid.UUID) (*domain.User, error) {
	u, ok := m.byID[userID]
	if !ok {
		return nil, auth.ErrNotFound
	}
	return u, nil
}

func (m *mockRepo) UpdateUserRole(_ context.Context, userID uuid.UUID, role domain.Role) error {
	u, ok := m.byID[userID]
	if !ok {
		return auth.ErrNotFound
	}
	u.Role = role
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func parseJWTClaims(t *testing.T, tokenStr string) jwt.MapClaims {
	t.Helper()
	token, err := jwt.Parse(tokenStr, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte("changeme"), nil
	})
	if err != nil {
		t.Fatalf("parse JWT: %v", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("claims not MapClaims")
	}
	return claims
}

// ── Register tests ────────────────────────────────────────────────────────────

// TestRegister_Success verifies that a valid registration creates a user and
// returns a signed JWT containing the correct user_id and role.
// Requirements: 1.1, 1.2, 1.6
func TestRegister_Success(t *testing.T) {
	repo := newMockRepo()
	svc := auth.NewService(repo)

	req := auth.RegisterRequest{
		Name:     "Budi Santoso",
		Email:    "budi@example.com",
		Password: "password123",
		Role:     domain.RoleLearner,
	}

	res, err := svc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if res.User == nil {
		t.Fatal("expected non-nil user")
	}
	if res.User.Email != req.Email {
		t.Errorf("email mismatch: got %q, want %q", res.User.Email, req.Email)
	}
	if res.User.Role != domain.RoleLearner {
		t.Errorf("role mismatch: got %q, want %q", res.User.Role, domain.RoleLearner)
	}
	// Password must NOT be stored in plain text
	if res.User.PasswordHash == req.Password {
		t.Error("password hash must not equal plain-text password")
	}
	if res.User.PasswordHash == "" {
		t.Error("password hash must not be empty")
	}

	// JWT must contain correct claims
	claims := parseJWTClaims(t, res.Token)
	if claims["user_id"] != res.User.ID.String() {
		t.Errorf("JWT user_id mismatch: got %v, want %v", claims["user_id"], res.User.ID.String())
	}
	if claims["role"] != string(domain.RoleLearner) {
		t.Errorf("JWT role mismatch: got %v, want %v", claims["role"], domain.RoleLearner)
	}
}

// TestRegister_ContributorRole verifies that a contributor role is accepted.
// Requirements: 1.1
func TestRegister_ContributorRole(t *testing.T) {
	repo := newMockRepo()
	svc := auth.NewService(repo)

	req := auth.RegisterRequest{
		Name:     "Sari Dewi",
		Email:    "sari@example.com",
		Password: "securepass",
		Role:     domain.RoleContributor,
	}

	res, err := svc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.User.Role != domain.RoleContributor {
		t.Errorf("expected contributor role, got %q", res.User.Role)
	}
}

// TestRegister_DuplicateEmail verifies that registering with an existing email
// returns ErrDuplicateEmail.
// Requirements: 1.3
func TestRegister_DuplicateEmail(t *testing.T) {
	repo := newMockRepo()
	svc := auth.NewService(repo)

	req := auth.RegisterRequest{
		Name:     "Budi Santoso",
		Email:    "budi@example.com",
		Password: "password123",
		Role:     domain.RoleLearner,
	}

	if _, err := svc.Register(context.Background(), req); err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	_, err := svc.Register(context.Background(), req)
	if !errors.Is(err, auth.ErrDuplicateEmail) {
		t.Errorf("expected ErrDuplicateEmail, got %v", err)
	}
}

// TestRegister_JWTExpiry verifies that the JWT has an expiry approximately 24h
// in the future.
// Requirements: 1.6
func TestRegister_JWTExpiry(t *testing.T) {
	repo := newMockRepo()
	svc := auth.NewService(repo)

	req := auth.RegisterRequest{
		Name:     "Andi",
		Email:    "andi@example.com",
		Password: "password123",
		Role:     domain.RoleLearner,
	}

	before := time.Now()
	res, err := svc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	claims := parseJWTClaims(t, res.Token)
	expFloat, ok := claims["exp"].(float64)
	if !ok {
		t.Fatal("exp claim missing or wrong type")
	}
	expTime := time.Unix(int64(expFloat), 0)
	expectedExp := before.Add(24 * time.Hour)

	// Allow ±5 seconds tolerance
	diff := expTime.Sub(expectedExp)
	if diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("JWT expiry %v is not ~24h from now (expected ~%v)", expTime, expectedExp)
	}
}

// ── Login tests ───────────────────────────────────────────────────────────────

// TestLogin_Success verifies that valid credentials return a JWT with correct
// claims.
// Requirements: 1.4, 1.5
func TestLogin_Success(t *testing.T) {
	repo := newMockRepo()
	svc := auth.NewService(repo)

	// Register first
	regReq := auth.RegisterRequest{
		Name:     "Budi Santoso",
		Email:    "budi@example.com",
		Password: "password123",
		Role:     domain.RoleLearner,
	}
	regRes, err := svc.Register(context.Background(), regReq)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Now login
	loginReq := auth.LoginRequest{
		Email:    "budi@example.com",
		Password: "password123",
	}
	res, err := svc.Login(context.Background(), loginReq)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Token == "" {
		t.Fatal("expected non-empty token")
	}

	claims := parseJWTClaims(t, res.Token)
	if claims["user_id"] != regRes.User.ID.String() {
		t.Errorf("JWT user_id mismatch: got %v, want %v", claims["user_id"], regRes.User.ID.String())
	}
	if claims["role"] != string(domain.RoleLearner) {
		t.Errorf("JWT role mismatch: got %v, want %v", claims["role"], domain.RoleLearner)
	}
}

// TestLogin_WrongPassword verifies that an incorrect password returns
// ErrInvalidCredentials.
// Requirements: 1.5
func TestLogin_WrongPassword(t *testing.T) {
	repo := newMockRepo()
	svc := auth.NewService(repo)

	regReq := auth.RegisterRequest{
		Name:     "Budi Santoso",
		Email:    "budi@example.com",
		Password: "password123",
		Role:     domain.RoleLearner,
	}
	if _, err := svc.Register(context.Background(), regReq); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	_, err := svc.Login(context.Background(), auth.LoginRequest{
		Email:    "budi@example.com",
		Password: "wrongpassword",
	})
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

// TestLogin_UnknownEmail verifies that a non-existent email returns
// ErrInvalidCredentials (not ErrNotFound, to avoid leaking info).
// Requirements: 1.5
func TestLogin_UnknownEmail(t *testing.T) {
	repo := newMockRepo()
	svc := auth.NewService(repo)

	_, err := svc.Login(context.Background(), auth.LoginRequest{
		Email:    "nobody@example.com",
		Password: "password123",
	})
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

// TestLogin_JWTExpiry verifies that the login JWT has ~24h expiry.
// Requirements: 1.4
func TestLogin_JWTExpiry(t *testing.T) {
	repo := newMockRepo()
	svc := auth.NewService(repo)

	regReq := auth.RegisterRequest{
		Name:     "Budi",
		Email:    "budi@example.com",
		Password: "password123",
		Role:     domain.RoleLearner,
	}
	if _, err := svc.Register(context.Background(), regReq); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	before := time.Now()
	res, err := svc.Login(context.Background(), auth.LoginRequest{
		Email:    "budi@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	claims := parseJWTClaims(t, res.Token)
	expFloat, ok := claims["exp"].(float64)
	if !ok {
		t.Fatal("exp claim missing or wrong type")
	}
	expTime := time.Unix(int64(expFloat), 0)
	expectedExp := before.Add(24 * time.Hour)

	diff := expTime.Sub(expectedExp)
	if diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("JWT expiry %v is not ~24h from now (expected ~%v)", expTime, expectedExp)
	}
}

// ── UpgradeRole tests ─────────────────────────────────────────────────────────

// TestUpgradeRole_LearnerToContributor verifies that a learner can upgrade to
// contributor and receives a new JWT with the updated role.
// Requirements: 23.3, 23.4, 23.6
func TestUpgradeRole_LearnerToContributor(t *testing.T) {
	repo := newMockRepo()
	svc := auth.NewService(repo)

	regReq := auth.RegisterRequest{
		Name:     "Budi Santoso",
		Email:    "budi@example.com",
		Password: "password123",
		Role:     domain.RoleLearner,
	}
	regRes, err := svc.Register(context.Background(), regReq)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	upgradeRes, err := svc.UpgradeRole(context.Background(), regRes.User.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if upgradeRes.User.Role != domain.RoleContributor {
		t.Errorf("expected contributor role, got %q", upgradeRes.User.Role)
	}
	if upgradeRes.Token == "" {
		t.Fatal("expected non-empty token after upgrade")
	}

	// New JWT must reflect the new role
	claims := parseJWTClaims(t, upgradeRes.Token)
	if claims["role"] != string(domain.RoleContributor) {
		t.Errorf("JWT role mismatch: got %v, want %v", claims["role"], domain.RoleContributor)
	}
}

// TestUpgradeRole_AlreadyContributor verifies that upgrading a contributor
// returns ErrRoleAlreadyUpgraded.
// Requirements: 23.6
func TestUpgradeRole_AlreadyContributor(t *testing.T) {
	repo := newMockRepo()
	svc := auth.NewService(repo)

	regReq := auth.RegisterRequest{
		Name:     "Sari Dewi",
		Email:    "sari@example.com",
		Password: "password123",
		Role:     domain.RoleContributor,
	}
	regRes, err := svc.Register(context.Background(), regReq)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	_, err = svc.UpgradeRole(context.Background(), regRes.User.ID)
	if !errors.Is(err, auth.ErrRoleAlreadyUpgraded) {
		t.Errorf("expected ErrRoleAlreadyUpgraded, got %v", err)
	}
}

// TestUpgradeRole_AdminNotUpgradeable verifies that an admin cannot be upgraded.
// Requirements: 23.6
func TestUpgradeRole_AdminNotUpgradeable(t *testing.T) {
	repo := newMockRepo()
	svc := auth.NewService(repo)

	// Manually insert an admin user into the mock repo
	adminID := uuid.New()
	adminUser := &domain.User{
		ID:           adminID,
		Name:         "Admin User",
		Email:        "admin@example.com",
		PasswordHash: "somehash",
		Role:         domain.RoleAdmin,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	repo.users[adminUser.Email] = adminUser
	repo.byID[adminID] = adminUser

	_, err := svc.UpgradeRole(context.Background(), adminID)
	if !errors.Is(err, auth.ErrRoleAlreadyUpgraded) {
		t.Errorf("expected ErrRoleAlreadyUpgraded for admin, got %v", err)
	}
}

// TestUpgradeRole_UserNotFound verifies that upgrading a non-existent user
// returns ErrNotFound.
func TestUpgradeRole_UserNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := auth.NewService(repo)

	_, err := svc.UpgradeRole(context.Background(), uuid.New())
	if !errors.Is(err, auth.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
