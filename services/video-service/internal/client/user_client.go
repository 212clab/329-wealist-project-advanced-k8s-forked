package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UserClient handles communication with user-service
type UserClient interface {
	ValidateWorkspaceMember(ctx context.Context, workspaceID, userID uuid.UUID, token string) (bool, error)
}

// WorkspaceValidationResponse represents the response from workspace validation endpoint
type WorkspaceValidationResponse struct {
	WorkspaceID uuid.UUID `json:"workspaceId"`
	UserID      uuid.UUID `json:"userId"`
	Valid       bool      `json:"valid"`
	IsValid     bool      `json:"isValid"`
	IsMember    bool      `json:"isMember"`
}

type userClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewUserClient creates a new user-service client
func NewUserClient(baseURL string, logger *zap.Logger) UserClient {
	return &userClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		logger: logger,
	}
}

// ValidateWorkspaceMember checks if a user is a member of a workspace
func (c *userClient) ValidateWorkspaceMember(ctx context.Context, workspaceID, userID uuid.UUID, token string) (bool, error) {
	// baseURL is expected to be like "http://user-service:8080/api"
	// The endpoint is at /workspaces/{workspaceId}/validate-member/{userId}
	url := fmt.Sprintf("%s/workspaces/%s/validate-member/%s", c.baseURL, workspaceID.String(), userID.String())

	c.logger.Debug("Validating workspace member",
		zap.String("url", url),
		zap.String("workspace_id", workspaceID.String()),
		zap.String("user_id", userID.String()),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to call user-service",
			zap.Error(err),
			zap.String("url", url),
		)
		return false, fmt.Errorf("failed to call user-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("User-service returned non-200 status",
			zap.Int("status", resp.StatusCode),
			zap.String("url", url),
		)
		// 403 = not a member, 404 = workspace not found
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, fmt.Errorf("user-service returned status %d", resp.StatusCode)
	}

	var response WorkspaceValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		c.logger.Error("Failed to decode response", zap.Error(err))
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check all possible fields
	isValid := response.Valid || response.IsValid || response.IsMember

	c.logger.Debug("Workspace member validation result",
		zap.Bool("is_valid", isValid),
		zap.String("workspace_id", workspaceID.String()),
		zap.String("user_id", userID.String()),
	)

	return isValid, nil
}
