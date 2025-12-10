package repository

import (
	"noti-service/internal/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(notification *domain.Notification) error {
	return r.db.Create(notification).Error
}

func (r *NotificationRepository) CreateBatch(notifications []domain.Notification) error {
	return r.db.Create(&notifications).Error
}

func (r *NotificationRepository) GetByID(id uuid.UUID) (*domain.Notification, error) {
	var notification domain.Notification
	err := r.db.First(&notification, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *NotificationRepository) GetByIDAndUserID(id, userID uuid.UUID) (*domain.Notification, error) {
	var notification domain.Notification
	err := r.db.First(&notification, "id = ? AND target_user_id = ?", id, userID).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *NotificationRepository) GetByUserAndWorkspace(userID, workspaceID uuid.UUID, page, limit int, unreadOnly bool) ([]domain.Notification, int64, error) {
	var notifications []domain.Notification
	var total int64

	query := r.db.Model(&domain.Notification{}).
		Where("target_user_id = ? AND workspace_id = ?", userID, workspaceID)

	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

func (r *NotificationRepository) GetUnreadCount(userID, workspaceID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&domain.Notification{}).
		Where("target_user_id = ? AND workspace_id = ? AND is_read = ?", userID, workspaceID, false).
		Count(&count).Error
	return count, err
}

func (r *NotificationRepository) MarkAsRead(id, userID uuid.UUID) (*domain.Notification, error) {
	now := time.Now()
	result := r.db.Model(&domain.Notification{}).
		Where("id = ? AND target_user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		})

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return r.GetByID(id)
}

func (r *NotificationRepository) MarkAllAsRead(userID, workspaceID uuid.UUID) (int64, error) {
	now := time.Now()
	result := r.db.Model(&domain.Notification{}).
		Where("target_user_id = ? AND workspace_id = ? AND is_read = ?", userID, workspaceID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		})

	return result.RowsAffected, result.Error
}

func (r *NotificationRepository) Delete(id, userID uuid.UUID) (bool, bool, error) {
	// Get notification first to check if it was unread
	var notification domain.Notification
	if err := r.db.First(&notification, "id = ? AND target_user_id = ?", id, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, false, nil
		}
		return false, false, err
	}

	wasUnread := !notification.IsRead

	// Delete
	if err := r.db.Delete(&notification).Error; err != nil {
		return false, false, err
	}

	return true, wasUnread, nil
}

func (r *NotificationRepository) CleanupOld(daysOld int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -daysOld)
	result := r.db.Where("is_read = ? AND created_at < ?", true, cutoff).
		Delete(&domain.Notification{})
	return result.RowsAffected, result.Error
}
