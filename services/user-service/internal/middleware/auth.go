// Package middleware는 HTTP 미들웨어를 제공합니다.
package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	commonauth "github.com/OrangesCloud/wealist-advanced-go-pkg/auth"
	"user-service/internal/response"
)

// TokenValidator는 auth-service 토큰 검증 인터페이스입니다.
type TokenValidator interface {
	ValidateToken(ctx context.Context, tokenStr string) (uuid.UUID, error)
}

// AuthWithValidator는 auth-service를 통해 JWT 토큰을 검증하는 미들웨어입니다.
func AuthWithValidator(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		userID, err := validator.ValidateToken(ctx, tokenString)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("jwtToken", tokenString)
		c.Next()
	}
}

// Auth는 JWT 토큰을 로컬에서 검증하는 미들웨어입니다 (fallback).
// 공통 모듈을 사용합니다.
func Auth(jwtSecret string) gin.HandlerFunc {
	return commonauth.JWTMiddleware(jwtSecret)
}

// GetUserID는 컨텍스트에서 사용자 ID를 추출합니다.
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	return commonauth.GetUserID(c)
}

// GetJWTToken은 컨텍스트에서 JWT 토큰을 추출합니다.
func GetJWTToken(c *gin.Context) (string, bool) {
	return commonauth.GetJWTToken(c)
}
