package service

import (
	"context"
	"encoding/json"
	"fmt"
	"noti-service/internal/config"
	"noti-service/internal/domain"
	"noti-service/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type NotificationService struct {
	repo   *repository.NotificationRepository
	redis  *redis.Client
	config *config.Config
	logger *zap.Logger
}

func NewNotificationService(
	repo *repository.NotificationRepository,
	redis *redis.Client,
	config *config.Config,
	logger *zap.Logger,
) *NotificationService {
	return &NotificationService{
		repo:   repo,
		redis:  redis,
		config: config,
		logger: logger,
	}
}

func (s *NotificationService) CreateNotification(ctx context.Context, event *domain.NotificationEvent) (*domain.Notification, error) {
	notification := &domain.Notification{
		ID:           uuid.New(),
		Type:         event.Type,
		ActorID:      event.ActorID,
		TargetUserID: event.TargetUserID,
		WorkspaceID:  event.WorkspaceID,
		ResourceType: event.ResourceType,
		ResourceID:   event.ResourceID,
		ResourceName: event.ResourceName,
		Metadata:     event.Metadata,
		IsRead:       false,
		CreatedAt:    time.Now(),
	}

	if event.OccurredAt != nil {
		notification.CreatedAt = *event.OccurredAt
	}

	if err := s.repo.Create(notification); err != nil {
		return nil, err
	}

	// Publish to Redis for SSE clients
	s.publishNotification(ctx, notification)

	// Invalidate cache
	s.invalidateUnreadCountCache(ctx, notification.TargetUserID, notification.WorkspaceID)

	s.logger.Info("notification created",
		zap.String("id", notification.ID.String()),
		zap.String("type", string(notification.Type)),
		zap.String("targetUserId", notification.TargetUserID.String()),
	)

	return notification, nil
}

func (s *NotificationService) CreateBulkNotifications(ctx context.Context, events []domain.NotificationEvent) ([]domain.Notification, error) {
	notifications := make([]domain.Notification, 0, len(events))

	for _, event := range events {
		notification, err := s.CreateNotification(ctx, &event)
		if err != nil {
			s.logger.Error("failed to create notification in bulk", zap.Error(err))
			continue
		}
		notifications = append(notifications, *notification)
	}

	return notifications, nil
}

func (s *NotificationService) GetNotifications(ctx context.Context, userID, workspaceID uuid.UUID, page, limit int, unreadOnly bool) (*domain.PaginatedNotifications, error) {
	notifications, total, err := s.repo.GetByUserAndWorkspace(userID, workspaceID, page, limit, unreadOnly)
	if err != nil {
		return nil, err
	}

	hasMore := int64(page*limit) < total

	return &domain.PaginatedNotifications{
		Notifications: notifications,
		Total:         total,
		Page:          page,
		Limit:         limit,
		HasMore:       hasMore,
	}, nil
}

func (s *NotificationService) GetNotificationByID(ctx context.Context, id, userID uuid.UUID) (*domain.Notification, error) {
	return s.repo.GetByIDAndUserID(id, userID)
}

func (s *NotificationService) MarkAsRead(ctx context.Context, id, userID uuid.UUID) (*domain.Notification, error) {
	notification, err := s.repo.MarkAsRead(id, userID)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	s.invalidateUnreadCountCache(ctx, notification.TargetUserID, notification.WorkspaceID)

	return notification, nil
}

func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID, workspaceID uuid.UUID) (int64, error) {
	count, err := s.repo.MarkAllAsRead(userID, workspaceID)
	if err != nil {
		return 0, err
	}

	// Invalidate cache
	s.invalidateUnreadCountCache(ctx, userID, workspaceID)

	return count, nil
}

func (s *NotificationService) GetUnreadCount(ctx context.Context, userID, workspaceID uuid.UUID) (*domain.UnreadCount, error) {
	cacheKey := fmt.Sprintf("unread:%s:%s", userID.String(), workspaceID.String())

	// Try cache first
	if s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Int64()
		if err == nil {
			return &domain.UnreadCount{
				Count:       cached,
				WorkspaceID: workspaceID,
			}, nil
		}
	}

	// Get from DB
	count, err := s.repo.GetUnreadCount(userID, workspaceID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if s.redis != nil {
		ttl := time.Duration(s.config.App.CacheUnreadTTL) * time.Second
		s.redis.Set(ctx, cacheKey, count, ttl)
	}

	return &domain.UnreadCount{
		Count:       count,
		WorkspaceID: workspaceID,
	}, nil
}

func (s *NotificationService) DeleteNotification(ctx context.Context, id, userID uuid.UUID) (bool, error) {
	// Get notification to find workspace ID for cache invalidation
	notification, _ := s.repo.GetByIDAndUserID(id, userID)

	deleted, wasUnread, err := s.repo.Delete(id, userID)
	if err != nil {
		return false, err
	}

	// Invalidate cache if was unread
	if deleted && wasUnread && notification != nil {
		s.invalidateUnreadCountCache(ctx, userID, notification.WorkspaceID)
	}

	return deleted, nil
}

func (s *NotificationService) CleanupOldNotifications(ctx context.Context) (int64, error) {
	return s.repo.CleanupOld(s.config.App.CleanupDays)
}

func (s *NotificationService) publishNotification(ctx context.Context, notification *domain.Notification) {
	if s.redis == nil {
		return
	}

	channel := fmt.Sprintf("notifications:user:%s", notification.TargetUserID.String())
	data, err := json.Marshal(notification)
	if err != nil {
		s.logger.Error("failed to marshal notification for publish", zap.Error(err))
		return
	}

	if err := s.redis.Publish(ctx, channel, data).Err(); err != nil {
		s.logger.Error("failed to publish notification", zap.Error(err))
	}
}

func (s *NotificationService) invalidateUnreadCountCache(ctx context.Context, userID, workspaceID uuid.UUID) {
	if s.redis == nil {
		return
	}

	cacheKey := fmt.Sprintf("unread:%s:%s", userID.String(), workspaceID.String())
	if err := s.redis.Del(ctx, cacheKey).Err(); err != nil {
		s.logger.Error("failed to invalidate unread cache", zap.Error(err))
	}
}
