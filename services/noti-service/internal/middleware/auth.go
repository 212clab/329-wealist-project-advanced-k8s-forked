// Package middleware provides HTTP middleware for noti-service.
//
// This package includes authentication middleware that validates JWT tokens
// either through the auth-service or locally using the configured secret key.
// It also provides workspace extraction and internal API key validation.
package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"noti-service/internal/response"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TokenValidator defines the interface for JWT token validation.
type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (uuid.UUID, error)
}

// AuthServiceValidator implements TokenValidator using auth-service with
// local JWT fallback.
type AuthServiceValidator struct {
	authServiceURL string
	secretKey      string
	httpClient     *http.Client
	logger         *zap.Logger
}

// NewAuthServiceValidator creates a new AuthServiceValidator.
func NewAuthServiceValidator(authServiceURL, secretKey string, logger *zap.Logger) *AuthServiceValidator {
	return &AuthServiceValidator{
		authServiceURL: authServiceURL,
		secretKey:      secretKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		logger: logger,
	}
}

func (v *AuthServiceValidator) ValidateToken(ctx context.Context, tokenString string) (uuid.UUID, error) {
	// Try auth service first
	if v.authServiceURL != "" {
		userID, err := v.validateWithAuthService(ctx, tokenString)
		if err == nil {
			return userID, nil
		}
		v.logger.Debug("Auth service validation failed, falling back to local", zap.Error(err))
	}

	// Fallback to local JWT validation
	return v.validateLocally(tokenString)
}

func (v *AuthServiceValidator) validateWithAuthService(ctx context.Context, token string) (uuid.UUID, error) {
	url := v.authServiceURL + "/api/auth/validate"

	reqBody, _ := json.Marshal(map[string]string{"token": token})
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return uuid.Nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return uuid.Nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return uuid.Nil, jwt.ErrTokenInvalidClaims
	}

	var result struct {
		UserID string `json:"userId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(result.UserID)
}

func (v *AuthServiceValidator) validateLocally(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(v.secretKey), nil
	})

	if err != nil || !token.Valid {
		return uuid.Nil, jwt.ErrTokenInvalidClaims
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, jwt.ErrTokenInvalidClaims
	}

	// Try different claim keys
	var userIDStr string
	for _, key := range []string{"sub", "userId", "user_id"} {
		if val, exists := claims[key]; exists {
			userIDStr = val.(string)
			break
		}
	}

	if userIDStr == "" {
		return uuid.Nil, jwt.ErrTokenInvalidClaims
	}

	return uuid.Parse(userIDStr)
}

// AuthMiddleware validates JWT token from Authorization header.
// It extracts the token from the Bearer scheme and validates it.
func AuthMiddleware(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "No authorization header")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]
		userID, err := validator.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			response.Unauthorized(c, "Invalid token")
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

// InternalAuthMiddleware validates internal API key for service-to-service calls.
func InternalAuthMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		providedKey := c.GetHeader("x-internal-api-key")
		if providedKey == "" {
			providedKey = c.GetHeader("X-Internal-Api-Key")
		}

		if providedKey == "" || providedKey != apiKey {
			response.Unauthorized(c, "Invalid internal API key")
			c.Abort()
			return
		}

		c.Next()
	}
}

// WorkspaceMiddleware extracts workspace ID from header
func WorkspaceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceIDStr := c.GetHeader("x-workspace-id")
		if workspaceIDStr == "" {
			workspaceIDStr = c.GetHeader("X-Workspace-Id")
		}

		if workspaceIDStr != "" {
			workspaceID, err := uuid.Parse(workspaceIDStr)
			if err == nil {
				c.Set("workspace_id", workspaceID)
			}
		}

		c.Next()
	}
}

// RequireWorkspace ensures workspace ID is present in the context.
func RequireWorkspace() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("workspace_id")
		if !exists {
			response.BadRequest(c, "x-workspace-id header is required")
			c.Abort()
			return
		}
		c.Next()
	}
}

// SSEAuthMiddleware validates JWT token from query parameter for SSE connections.
// EventSource API doesn't support custom headers, so token must be passed as query param.
func SSEAuthMiddleware(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try query parameter first (for SSE)
		tokenString := c.Query("token")

		// Fallback to Authorization header
		if tokenString == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					tokenString = parts[1]
				}
			}
		}

		if tokenString == "" {
			response.Unauthorized(c, "No token provided")
			c.Abort()
			return
		}

		userID, err := validator.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			response.Unauthorized(c, "Invalid token")
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
