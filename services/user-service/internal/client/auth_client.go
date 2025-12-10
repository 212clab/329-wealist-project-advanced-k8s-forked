package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AuthClient handles authentication with auth-service
type AuthClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// TokenValidationRequest represents the request to auth-service
type TokenValidationRequest struct {
	Token string `json:"token"`
}

// TokenValidationResponse represents the response from auth-service
type TokenValidationResponse struct {
	Valid   bool   `json:"valid"`
	UserID  string `json:"userId"`
	Message string `json:"message,omitempty"`
}

// NewAuthClient creates a new AuthClient
func NewAuthClient(baseURL string, timeout time.Duration, logger *zap.Logger) *AuthClient {
	return &AuthClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

// ValidateToken validates a token via auth-service
func (c *AuthClient) ValidateToken(ctx context.Context, tokenStr string) (uuid.UUID, error) {
	// Build request URL
	url := fmt.Sprintf("%s/api/auth/validate", c.baseURL)

	// Create request body
	reqBody := TokenValidationRequest{Token: tokenStr}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create POST request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to validate token", zap.Error(err))
		return uuid.Nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return uuid.Nil, fmt.Errorf("token validation failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var tokenResp TokenValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return uuid.Nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !tokenResp.Valid {
		return uuid.Nil, fmt.Errorf("invalid token: %s", tokenResp.Message)
	}

	// Parse user ID
	userID, err := uuid.Parse(tokenResp.UserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	return userID, nil
}
