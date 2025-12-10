package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (uuid.UUID, error)
}

type AuthServiceValidator struct {
	authServiceURL string
	secretKey      string
	httpClient     *http.Client
	logger         *zap.Logger
}

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
	defer resp.Body.Close()

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

// AuthMiddleware validates JWT token from Authorization header
func AuthMiddleware(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   gin.H{"code": "UNAUTHORIZED", "message": "No authorization header"},
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   gin.H{"code": "UNAUTHORIZED", "message": "Invalid authorization header format"},
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		userID, err := validator.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   gin.H{"code": "UNAUTHORIZED", "message": "Invalid token"},
			})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("token", tokenString)
		c.Next()
	}
}
