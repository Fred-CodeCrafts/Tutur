package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/yourusername/bahasa-daerah-platform/internal/domain"
	"github.com/yourusername/bahasa-daerah-platform/pkg/response"
)

// contextKey is an unexported type for context keys in this package.
type contextKey int

const (
	contextKeyUserID contextKey = iota
	contextKeyRole
)

// Authenticate returns middleware that validates a Bearer JWT and injects
// userID (uuid.UUID) and role (domain.Role) into the request context.
// Returns 401 if the token is missing or invalid.
func Authenticate(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				response.Unauthorized(w)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				response.Unauthorized(w)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				response.Unauthorized(w)
				return
			}

			userIDStr, ok := claims["user_id"].(string)
			if !ok {
				response.Unauthorized(w)
				return
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				response.Unauthorized(w)
				return
			}

			roleStr, ok := claims["role"].(string)
			if !ok {
				response.Unauthorized(w)
				return
			}

			role := domain.Role(roleStr)

			ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
			ctx = context.WithValue(ctx, contextKeyRole, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole returns middleware that checks the context role against the
// allowed roles. Returns 403 if the role is not in the allowed list.
func RequireRole(roles ...domain.Role) func(http.Handler) http.Handler {
	allowed := make(map[domain.Role]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := RoleFromContext(r.Context())
			if !ok {
				response.Unauthorized(w)
				return
			}

			if _, permitted := allowed[role]; !permitted {
				response.Forbidden(w, "You do not have permission to access this resource.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// UserIDFromContext extracts the userID from the context.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(contextKeyUserID).(uuid.UUID)
	return id, ok
}

// RoleFromContext extracts the role from the context.
func RoleFromContext(ctx context.Context) (domain.Role, bool) {
	role, ok := ctx.Value(contextKeyRole).(domain.Role)
	return role, ok
}
