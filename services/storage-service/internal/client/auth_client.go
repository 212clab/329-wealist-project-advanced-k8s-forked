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

// AuthClient handles auth service communication
type AuthClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// TokenValidationRequest represents the request to auth service
type TokenValidationRequest struct {
	Token string `json:"token"`
}

// TokenValidationResponse represents the response from auth service
type TokenValidationResponse struct {
	UserID  string `json:"userId"`
	Valid   bool   `json:"valid"`
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

// ValidateToken validates a JWT token with auth service
func (c *AuthClient) ValidateToken(ctx context.Context, tokenStr string) (uuid.UUID, error) {
	url := fmt.Sprintf("%s/api/auth/validate", c.baseURL)

	// Prepare request body
	reqBody := TokenValidationRequest{Token: tokenStr}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to validate token", zap.Error(err))
		return uuid.Nil, fmt.Errorf("failed to validate token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return uuid.Nil, fmt.Errorf("token validation failed with status: %d", resp.StatusCode)
	}

	var result TokenValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return uuid.Nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Valid {
		return uuid.Nil, fmt.Errorf("token is not valid: %s", result.Message)
	}

	userID, err := uuid.Parse(result.UserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse user ID: %w", err)
	}

	return userID, nil
}
