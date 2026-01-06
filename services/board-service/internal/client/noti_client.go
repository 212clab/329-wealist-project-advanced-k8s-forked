// Package client provides HTTP clients for external service communication.
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

	"project-board-api/internal/config"
)

// NotificationType defines the type of notification
type NotificationType string

const (
	// Board (Kanban) notification types
	NotificationTypeBoardAssigned         NotificationType = "BOARD_ASSIGNED"
	NotificationTypeBoardUnassigned       NotificationType = "BOARD_UNASSIGNED"
	NotificationTypeBoardParticipantAdded NotificationType = "BOARD_PARTICIPANT_ADDED"
	NotificationTypeBoardUpdated          NotificationType = "BOARD_UPDATED"
	NotificationTypeBoardStatusChanged    NotificationType = "BOARD_STATUS_CHANGED"
	NotificationTypeBoardCommentAdded     NotificationType = "BOARD_COMMENT_ADDED"
	NotificationTypeBoardDueSoon          NotificationType = "BOARD_DUE_SOON"
	NotificationTypeBoardOverdue          NotificationType = "BOARD_OVERDUE"
)

// ResourceType defines the type of resource
type ResourceType string

const (
	ResourceTypeBoard ResourceType = "board"
)

// NotificationEvent represents an incoming notification event
type NotificationEvent struct {
	Type         NotificationType       `json:"type"`
	ActorID      uuid.UUID              `json:"actorId"`
	TargetUserID uuid.UUID              `json:"targetUserId"`
	WorkspaceID  uuid.UUID              `json:"workspaceId"`
	ResourceType ResourceType           `json:"resourceType"`
	ResourceID   uuid.UUID              `json:"resourceId"`
	ResourceName *string                `json:"resourceName,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// BulkNotificationRequest represents a bulk notification creation request
type BulkNotificationRequest struct {
	Notifications []NotificationEvent `json:"notifications"`
}

// NotiClient defines the interface for notification service client
type NotiClient interface {
	// SendNotification sends a single notification
	SendNotification(ctx context.Context, event *NotificationEvent) error
	// SendBulkNotifications sends multiple notifications at once
	SendBulkNotifications(ctx context.Context, events []NotificationEvent) error
	// IsEnabled returns whether notification sending is enabled
	IsEnabled() bool
}

// notiClientImpl is the implementation of NotiClient
type notiClientImpl struct {
	httpClient     *http.Client
	baseURL        string
	internalAPIKey string
	enabled        bool
	logger         *zap.Logger
}

// NewNotiClient creates a new notification service client
func NewNotiClient(cfg *config.NotiAPIConfig, logger *zap.Logger) NotiClient {
	return &notiClientImpl{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		baseURL:        cfg.BaseURL,
		internalAPIKey: cfg.InternalAPIKey,
		enabled:        cfg.Enabled,
		logger:         logger,
	}
}

// IsEnabled returns whether notification sending is enabled
func (c *notiClientImpl) IsEnabled() bool {
	return c.enabled && c.internalAPIKey != ""
}

// SendNotification sends a single notification to noti-service
func (c *notiClientImpl) SendNotification(ctx context.Context, event *NotificationEvent) error {
	if !c.IsEnabled() {
		c.logger.Debug("Notification sending is disabled, skipping",
			zap.String("type", string(event.Type)),
			zap.String("targetUserId", event.TargetUserID.String()))
		return nil
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal notification event: %w", err)
	}

	url := fmt.Sprintf("%s/api/internal/notifications", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Api-Key", c.internalAPIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to send notification",
			zap.String("type", string(event.Type)),
			zap.Error(err))
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		c.logger.Error("Notification service returned error",
			zap.String("type", string(event.Type)),
			zap.Int("status", resp.StatusCode))
		return fmt.Errorf("notification service returned status %d", resp.StatusCode)
	}

	c.logger.Debug("Notification sent successfully",
		zap.String("type", string(event.Type)),
		zap.String("targetUserId", event.TargetUserID.String()))

	return nil
}

// SendBulkNotifications sends multiple notifications at once
func (c *notiClientImpl) SendBulkNotifications(ctx context.Context, events []NotificationEvent) error {
	if !c.IsEnabled() {
		c.logger.Debug("Notification sending is disabled, skipping bulk notifications",
			zap.Int("count", len(events)))
		return nil
	}

	if len(events) == 0 {
		return nil
	}

	// Maximum 100 notifications per request
	if len(events) > 100 {
		events = events[:100]
	}

	bulkReq := BulkNotificationRequest{
		Notifications: events,
	}

	body, err := json.Marshal(bulkReq)
	if err != nil {
		return fmt.Errorf("failed to marshal bulk notification request: %w", err)
	}

	url := fmt.Sprintf("%s/api/internal/notifications/bulk", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Api-Key", c.internalAPIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to send bulk notifications",
			zap.Int("count", len(events)),
			zap.Error(err))
		return fmt.Errorf("failed to send bulk notifications: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		c.logger.Error("Notification service returned error for bulk request",
			zap.Int("count", len(events)),
			zap.Int("status", resp.StatusCode))
		return fmt.Errorf("notification service returned status %d", resp.StatusCode)
	}

	c.logger.Debug("Bulk notifications sent successfully",
		zap.Int("count", len(events)))

	return nil
}

// NoopNotiClient is a no-op implementation of NotiClient for testing
type NoopNotiClient struct{}

func NewNoopNotiClient() NotiClient {
	return &NoopNotiClient{}
}

func (c *NoopNotiClient) SendNotification(ctx context.Context, event *NotificationEvent) error {
	return nil
}

func (c *NoopNotiClient) SendBulkNotifications(ctx context.Context, events []NotificationEvent) error {
	return nil
}

func (c *NoopNotiClient) IsEnabled() bool {
	return false
}

// Helper function to create a notification event for board assignment
func NewBoardAssignedEvent(actorID, targetUserID, workspaceID, boardID uuid.UUID, boardTitle, projectName string) *NotificationEvent {
	return &NotificationEvent{
		Type:         NotificationTypeBoardAssigned,
		ActorID:      actorID,
		TargetUserID: targetUserID,
		WorkspaceID:  workspaceID,
		ResourceType: ResourceTypeBoard,
		ResourceID:   boardID,
		ResourceName: &boardTitle,
		Metadata: map[string]interface{}{
			"projectName": projectName,
			"timestamp":   time.Now().Unix(),
		},
	}
}

// Helper function to create a notification event for board participant addition
func NewBoardParticipantAddedEvent(actorID, targetUserID, workspaceID, boardID uuid.UUID, boardTitle, projectName string) *NotificationEvent {
	return &NotificationEvent{
		Type:         NotificationTypeBoardParticipantAdded,
		ActorID:      actorID,
		TargetUserID: targetUserID,
		WorkspaceID:  workspaceID,
		ResourceType: ResourceTypeBoard,
		ResourceID:   boardID,
		ResourceName: &boardTitle,
		Metadata: map[string]interface{}{
			"projectName": projectName,
			"timestamp":   time.Now().Unix(),
		},
	}
}

// Helper function to create a notification event for board update
func NewBoardUpdatedEvent(actorID, targetUserID, workspaceID, boardID uuid.UUID, boardTitle, projectName string, changes map[string]interface{}) *NotificationEvent {
	metadata := map[string]interface{}{
		"projectName": projectName,
		"timestamp":   time.Now().Unix(),
	}
	// Merge changes into metadata
	for k, v := range changes {
		metadata[k] = v
	}

	return &NotificationEvent{
		Type:         NotificationTypeBoardUpdated,
		ActorID:      actorID,
		TargetUserID: targetUserID,
		WorkspaceID:  workspaceID,
		ResourceType: ResourceTypeBoard,
		ResourceID:   boardID,
		ResourceName: &boardTitle,
		Metadata:     metadata,
	}
}

// Helper function to create a notification event for board status change
func NewBoardStatusChangedEvent(actorID, targetUserID, workspaceID, boardID uuid.UUID, boardTitle, projectName, oldStatus, newStatus string) *NotificationEvent {
	return &NotificationEvent{
		Type:         NotificationTypeBoardStatusChanged,
		ActorID:      actorID,
		TargetUserID: targetUserID,
		WorkspaceID:  workspaceID,
		ResourceType: ResourceTypeBoard,
		ResourceID:   boardID,
		ResourceName: &boardTitle,
		Metadata: map[string]interface{}{
			"projectName": projectName,
			"oldStatus":   oldStatus,
			"newStatus":   newStatus,
			"timestamp":   time.Now().Unix(),
		},
	}
}
