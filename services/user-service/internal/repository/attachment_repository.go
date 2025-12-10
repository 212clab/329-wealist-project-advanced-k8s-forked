package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"user-service/internal/domain"
)

// AttachmentRepository handles attachment data access
type AttachmentRepository struct {
	db *gorm.DB
}

// NewAttachmentRepository creates a new AttachmentRepository
func NewAttachmentRepository(db *gorm.DB) *AttachmentRepository {
	return &AttachmentRepository{db: db}
}

// Create creates a new attachment
func (r *AttachmentRepository) Create(attachment *domain.Attachment) error {
	return r.db.Create(attachment).Error
}

// FindByID finds an attachment by ID
func (r *AttachmentRepository) FindByID(id uuid.UUID) (*domain.Attachment, error) {
	var attachment domain.Attachment
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&attachment).Error
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

// FindByEntity finds attachments by entity
func (r *AttachmentRepository) FindByEntity(entityType domain.EntityType, entityID uuid.UUID) ([]domain.Attachment, error) {
	var attachments []domain.Attachment
	err := r.db.Where("entity_type = ? AND entity_id = ? AND deleted_at IS NULL", entityType, entityID).Find(&attachments).Error
	return attachments, err
}

// FindTempByUser finds temporary attachments by user
func (r *AttachmentRepository) FindTempByUser(userID uuid.UUID) ([]domain.Attachment, error) {
	var attachments []domain.Attachment
	err := r.db.Where("uploaded_by = ? AND status = ? AND deleted_at IS NULL", userID, domain.AttachmentStatusTemp).Find(&attachments).Error
	return attachments, err
}

// Update updates an attachment
func (r *AttachmentRepository) Update(attachment *domain.Attachment) error {
	return r.db.Save(attachment).Error
}

// ConfirmAttachment confirms a temporary attachment
func (r *AttachmentRepository) ConfirmAttachment(id uuid.UUID, entityID uuid.UUID) error {
	return r.db.Model(&domain.Attachment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     domain.AttachmentStatusConfirmed,
			"entity_id":  entityID,
			"expires_at": nil,
		}).Error
}

// SoftDelete soft deletes an attachment
func (r *AttachmentRepository) SoftDelete(id uuid.UUID) error {
	return r.db.Model(&domain.Attachment{}).
		Where("id = ?", id).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}

// DeleteExpired deletes expired temporary attachments
func (r *AttachmentRepository) DeleteExpired() error {
	return r.db.Where("status = ? AND expires_at < ? AND deleted_at IS NULL", domain.AttachmentStatusTemp, time.Now()).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}

// FindExpired finds expired temporary attachments
func (r *AttachmentRepository) FindExpired() ([]domain.Attachment, error) {
	var attachments []domain.Attachment
	err := r.db.Where("status = ? AND expires_at < ? AND deleted_at IS NULL", domain.AttachmentStatusTemp, time.Now()).Find(&attachments).Error
	return attachments, err
}
