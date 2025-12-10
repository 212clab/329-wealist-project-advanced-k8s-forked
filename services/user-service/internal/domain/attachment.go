package domain

import (
	"time"

	"github.com/google/uuid"
)

// EntityType represents the type of entity the attachment belongs to
type EntityType string

const (
	EntityTypeUserProfile EntityType = "USER_PROFILE"
)

// AttachmentStatus represents the status of an attachment
type AttachmentStatus string

const (
	AttachmentStatusTemp      AttachmentStatus = "TEMP"
	AttachmentStatusConfirmed AttachmentStatus = "CONFIRMED"
)

// Attachment represents a file attachment
type Attachment struct {
	ID          uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EntityType  EntityType       `gorm:"type:varchar(50);not null;index" json:"entityType"`
	EntityID    *uuid.UUID       `gorm:"type:uuid;index" json:"entityId,omitempty"`
	Status      AttachmentStatus `gorm:"type:varchar(20);not null;default:'TEMP';index" json:"status"`
	FileName    string           `gorm:"not null" json:"fileName"`
	FileURL     string           `gorm:"not null" json:"fileUrl"`
	FileSize    int64            `gorm:"not null" json:"fileSize"`
	ContentType string           `gorm:"not null" json:"contentType"`
	UploadedBy  uuid.UUID        `gorm:"type:uuid;not null;index" json:"uploadedBy"`
	ExpiresAt   *time.Time       `gorm:"index" json:"expiresAt,omitempty"`
	CreatedAt   time.Time        `gorm:"not null" json:"createdAt"`
	UpdatedAt   time.Time        `gorm:"not null" json:"updatedAt"`
	DeletedAt   *time.Time       `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName specifies the table name for Attachment
func (Attachment) TableName() string {
	return "attachments"
}

// PresignedURLRequest represents the request for a presigned URL
type PresignedURLRequest struct {
	FileName    string `json:"fileName" binding:"required"`
	ContentType string `json:"contentType" binding:"required"`
	FileSize    int64  `json:"fileSize" binding:"required"`
}

// PresignedURLResponse represents the response for a presigned URL
type PresignedURLResponse struct {
	UploadURL string `json:"uploadUrl"`
	FileKey   string `json:"fileKey"`
	ExpiresAt int64  `json:"expiresAt"`
}

// SaveAttachmentRequest represents the request to save attachment metadata
type SaveAttachmentRequest struct {
	FileKey     string `json:"fileKey" binding:"required"`
	FileName    string `json:"fileName" binding:"required"`
	ContentType string `json:"contentType" binding:"required"`
	FileSize    int64  `json:"fileSize" binding:"required"`
}

// ConfirmAttachmentRequest represents the request to confirm an attachment
type ConfirmAttachmentRequest struct {
	AttachmentID uuid.UUID `json:"attachmentId" binding:"required"`
}

// AttachmentResponse represents the attachment response
type AttachmentResponse struct {
	AttachmentID uuid.UUID        `json:"attachmentId"`
	EntityType   EntityType       `json:"entityType"`
	EntityID     *uuid.UUID       `json:"entityId,omitempty"`
	Status       AttachmentStatus `json:"status"`
	FileName     string           `json:"fileName"`
	FileURL      string           `json:"fileUrl"`
	FileSize     int64            `json:"fileSize"`
	ContentType  string           `json:"contentType"`
	UploadedBy   uuid.UUID        `json:"uploadedBy"`
	ExpiresAt    *time.Time       `json:"expiresAt,omitempty"`
	CreatedAt    time.Time        `json:"createdAt"`
}

// ToResponse converts Attachment to AttachmentResponse
func (a *Attachment) ToResponse() AttachmentResponse {
	return AttachmentResponse{
		AttachmentID: a.ID,
		EntityType:   a.EntityType,
		EntityID:     a.EntityID,
		Status:       a.Status,
		FileName:     a.FileName,
		FileURL:      a.FileURL,
		FileSize:     a.FileSize,
		ContentType:  a.ContentType,
		UploadedBy:   a.UploadedBy,
		ExpiresAt:    a.ExpiresAt,
		CreatedAt:    a.CreatedAt,
	}
}
