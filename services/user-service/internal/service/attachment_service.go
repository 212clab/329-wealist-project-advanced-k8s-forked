package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"user-service/internal/client"
	"user-service/internal/domain"
	"user-service/internal/repository"
)

// AllowedImageTypes defines allowed image content types
var AllowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// MaxFileSize is the maximum allowed file size (20MB)
const MaxFileSize int64 = 20 * 1024 * 1024

// AttachmentService handles attachment business logic
type AttachmentService struct {
	attachmentRepo *repository.AttachmentRepository
	s3Client       *client.S3Client
	logger         *zap.Logger
}

// NewAttachmentService creates a new AttachmentService
func NewAttachmentService(
	attachmentRepo *repository.AttachmentRepository,
	s3Client *client.S3Client,
	logger *zap.Logger,
) *AttachmentService {
	return &AttachmentService{
		attachmentRepo: attachmentRepo,
		s3Client:       s3Client,
		logger:         logger,
	}
}

// GeneratePresignedURL generates a presigned URL for file upload
func (s *AttachmentService) GeneratePresignedURL(ctx context.Context, userID uuid.UUID, req domain.PresignedURLRequest) (*domain.PresignedURLResponse, error) {
	// Validate file type
	if !AllowedImageTypes[strings.ToLower(req.ContentType)] {
		return nil, errors.New("invalid file type, allowed: jpg, jpeg, png, gif, webp")
	}

	// Validate file size
	if req.FileSize > MaxFileSize {
		return nil, errors.New("file size exceeds maximum of 20MB")
	}

	// Generate presigned URL
	uploadURL, fileKey, err := s.s3Client.GeneratePresignedURL(ctx, req.FileName, req.ContentType)
	if err != nil {
		s.logger.Error("Failed to generate presigned URL", zap.Error(err))
		return nil, err
	}

	expiresAt := time.Now().Add(5 * time.Minute).Unix()

	return &domain.PresignedURLResponse{
		UploadURL: uploadURL,
		FileKey:   fileKey,
		ExpiresAt: expiresAt,
	}, nil
}

// SaveAttachment saves attachment metadata after S3 upload
func (s *AttachmentService) SaveAttachment(ctx context.Context, userID uuid.UUID, req domain.SaveAttachmentRequest) (*domain.Attachment, error) {
	// Validate file type
	if !AllowedImageTypes[strings.ToLower(req.ContentType)] {
		return nil, errors.New("invalid file type")
	}

	expiresAt := time.Now().Add(1 * time.Hour)
	fileURL := s.s3Client.GetFileURL(req.FileKey)

	attachment := &domain.Attachment{
		ID:          uuid.New(),
		EntityType:  domain.EntityTypeUserProfile,
		Status:      domain.AttachmentStatusTemp,
		FileName:    req.FileName,
		FileURL:     fileURL,
		FileSize:    req.FileSize,
		ContentType: req.ContentType,
		UploadedBy:  userID,
		ExpiresAt:   &expiresAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.attachmentRepo.Create(attachment); err != nil {
		s.logger.Error("Failed to save attachment", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Attachment saved", zap.String("attachmentId", attachment.ID.String()))
	return attachment, nil
}

// ConfirmAttachment confirms a temporary attachment and links it to an entity
func (s *AttachmentService) ConfirmAttachment(ctx context.Context, userID uuid.UUID, attachmentID, entityID uuid.UUID) (*domain.Attachment, error) {
	attachment, err := s.attachmentRepo.FindByID(attachmentID)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if attachment.UploadedBy != userID {
		return nil, errors.New("attachment not owned by user")
	}

	// Check status
	if attachment.Status != domain.AttachmentStatusTemp {
		return nil, errors.New("attachment is not in temporary status")
	}

	// Confirm attachment
	if err := s.attachmentRepo.ConfirmAttachment(attachmentID, entityID); err != nil {
		s.logger.Error("Failed to confirm attachment", zap.Error(err))
		return nil, err
	}

	attachment.Status = domain.AttachmentStatusConfirmed
	attachment.EntityID = &entityID
	attachment.ExpiresAt = nil

	s.logger.Info("Attachment confirmed", zap.String("attachmentId", attachmentID.String()))
	return attachment, nil
}

// GetAttachment gets an attachment by ID
func (s *AttachmentService) GetAttachment(attachmentID uuid.UUID) (*domain.Attachment, error) {
	return s.attachmentRepo.FindByID(attachmentID)
}

// DeleteAttachment deletes an attachment
func (s *AttachmentService) DeleteAttachment(ctx context.Context, userID uuid.UUID, attachmentID uuid.UUID) error {
	attachment, err := s.attachmentRepo.FindByID(attachmentID)
	if err != nil {
		return err
	}

	// Check ownership
	if attachment.UploadedBy != userID {
		return errors.New("attachment not owned by user")
	}

	// Delete from S3
	fileKey := strings.TrimPrefix(attachment.FileURL, s.s3Client.GetFileURL(""))
	if err := s.s3Client.DeleteFile(ctx, fileKey); err != nil {
		s.logger.Error("Failed to delete file from S3", zap.Error(err))
		// Continue to soft delete even if S3 deletion fails
	}

	// Soft delete from database
	if err := s.attachmentRepo.SoftDelete(attachmentID); err != nil {
		s.logger.Error("Failed to delete attachment", zap.Error(err))
		return err
	}

	s.logger.Info("Attachment deleted", zap.String("attachmentId", attachmentID.String()))
	return nil
}

// CleanupExpiredAttachments cleans up expired temporary attachments
func (s *AttachmentService) CleanupExpiredAttachments(ctx context.Context) error {
	// Find expired attachments
	expired, err := s.attachmentRepo.FindExpired()
	if err != nil {
		s.logger.Error("Failed to find expired attachments", zap.Error(err))
		return err
	}

	for _, attachment := range expired {
		// Delete from S3
		fileKey := strings.TrimPrefix(attachment.FileURL, s.s3Client.GetFileURL(""))
		if err := s.s3Client.DeleteFile(ctx, fileKey); err != nil {
			s.logger.Error("Failed to delete expired file from S3", zap.Error(err), zap.String("fileKey", fileKey))
		}
	}

	// Soft delete from database
	if err := s.attachmentRepo.DeleteExpired(); err != nil {
		s.logger.Error("Failed to delete expired attachments", zap.Error(err))
		return err
	}

	s.logger.Info("Cleaned up expired attachments", zap.Int("count", len(expired)))
	return nil
}
