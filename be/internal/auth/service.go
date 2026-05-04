package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
)

const (
	bcryptCost = 12
	jwtExpiry  = 24 * time.Hour
)

// ErrInvalidCredentials is returned when email/password don't match.
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrRoleAlreadyUpgraded is returned when the user is already contributor or admin.
var ErrRoleAlreadyUpgraded = errors.New("role already upgraded")

// RegisterRequest holds the fields required to register a new user.
type RegisterRequest struct {
	Name     string      `json:"name"`
	Email    string      `json:"email"`
	Password string      `json:"password"`
	Role     domain.Role `json:"role"`
}

// LoginRequest holds the fields required to log in.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is returned after a successful register, login, or role upgrade.
type AuthResponse struct {
	Token string       `json:"token"`
	User  *domain.User `json:"user"`
}

// Service defines the business logic interface for auth operations.
type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	UpgradeRole(ctx context.Context, userID uuid.UUID) (*AuthResponse, error)
}

type service struct {
	repo Repository
}

// NewService creates a new auth service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// Register hashes the password, creates the user, and returns a JWT.
func (s *service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         req.Role,
		IsActive:     true,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err // ErrDuplicateEmail propagates as-is
	}

	token, err := generateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &AuthResponse{Token: token, User: user}, nil
}

// Login verifies credentials and returns a JWT.
func (s *service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := generateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &AuthResponse{Token: token, User: user}, nil
}

// UpgradeRole promotes a learner to contributor and returns a new JWT.
func (s *service) UpgradeRole(ctx context.Context, userID uuid.UUID) (*AuthResponse, error) {
	// We need the current user to check their role. Fetch by ID via a small
	// workaround: we'll rely on the caller having validated the JWT, so we
	// trust the userID. We still need to load the user to return it.
	// Use a temporary approach: get user by ID via a direct query.
	// Since Repository only has GetUserByEmail, we'll add a helper or
	// use the update + re-fetch pattern.
	//
	// The repository interface has UpdateUserRole; we need to know the current
	// role. We'll use a separate internal method to get user by ID.
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	if user.Role == domain.RoleContributor || user.Role == domain.RoleAdmin {
		return nil, ErrRoleAlreadyUpgraded
	}

	if err := s.repo.UpdateUserRole(ctx, userID, domain.RoleContributor); err != nil {
		return nil, fmt.Errorf("upgrade role: %w", err)
	}

	user.Role = domain.RoleContributor

	token, err := generateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &AuthResponse{Token: token, User: user}, nil
}

// generateJWT creates a signed JWT for the given user with 24h expiry.
func generateJWT(user *domain.User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "changeme"
	}

	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"role":    string(user.Role),
		"exp":     time.Now().Add(jwtExpiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
