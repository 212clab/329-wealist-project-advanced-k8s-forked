package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"storage-service/internal/domain"
	"storage-service/internal/repository"
)

// ShareService handles share business logic
type ShareService struct {
	shareRepo  *repository.ShareRepository
	fileRepo   *repository.FileRepository
	folderRepo *repository.FolderRepository
	logger     *zap.Logger
}

// NewShareService creates a new ShareService
func NewShareService(
	shareRepo *repository.ShareRepository,
	fileRepo *repository.FileRepository,
	folderRepo *repository.FolderRepository,
	logger *zap.Logger,
) *ShareService {
	return &ShareService{
		shareRepo:  shareRepo,
		fileRepo:   fileRepo,
		folderRepo: folderRepo,
		logger:     logger,
	}
}

// generateShareLink generates a random share link token
func generateShareLink() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:32], nil
}

// CreateShare creates a new share for a file or folder
func (s *ShareService) CreateShare(ctx context.Context, req domain.CreateShareRequest, userID uuid.UUID) (*domain.ShareResponse, error) {
	// Validate entity exists
	var entityName string
	if req.EntityType == domain.ShareTypeFile {
		file, err := s.fileRepo.FindByID(ctx, req.EntityID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("file not found")
			}
			return nil, err
		}
		entityName = file.Name
	} else if req.EntityType == domain.ShareTypeFolder {
		folder, err := s.folderRepo.FindByID(ctx, req.EntityID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("folder not found")
			}
			return nil, err
		}
		entityName = folder.Name
	} else {
		return nil, errors.New("invalid entity type")
	}

	// Generate share link if public
	var shareLink *string
	var linkExpiresAt *time.Time

	if req.IsPublic {
		link, err := generateShareLink()
		if err != nil {
			return nil, errors.New("failed to generate share link")
		}
		shareLink = &link

		if req.ExpiresInDays != nil && *req.ExpiresInDays > 0 {
			expires := time.Now().AddDate(0, 0, *req.ExpiresInDays)
			linkExpiresAt = &expires
		}
	}

	now := time.Now()

	if req.EntityType == domain.ShareTypeFile {
		// Check if share already exists
		if req.SharedWithID != nil {
			existing, err := s.shareRepo.FindFileShareByFileAndUser(ctx, req.EntityID, *req.SharedWithID)
			if err == nil && existing != nil {
				return nil, errors.New("file is already shared with this user")
			}
		}

		share := &domain.FileShare{
			ID:            uuid.New(),
			FileID:        req.EntityID,
			SharedWithID:  req.SharedWithID,
			SharedByID:    userID,
			Permission:    req.Permission,
			ShareLink:     shareLink,
			LinkExpiresAt: linkExpiresAt,
			IsPublic:      req.IsPublic,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if err := s.shareRepo.CreateFileShare(ctx, share); err != nil {
			return nil, err
		}

		s.logger.Info("File share created",
			zap.String("shareId", share.ID.String()),
			zap.String("fileId", req.EntityID.String()),
			zap.String("userId", userID.String()),
		)

		return &domain.ShareResponse{
			ID:            share.ID,
			EntityType:    domain.ShareTypeFile,
			EntityID:      req.EntityID,
			EntityName:    entityName,
			SharedWithID:  req.SharedWithID,
			SharedByID:    userID,
			Permission:    req.Permission,
			ShareLink:     shareLink,
			LinkExpiresAt: linkExpiresAt,
			IsPublic:      req.IsPublic,
			IsExpired:     share.IsExpired(),
			CreatedAt:     now,
			UpdatedAt:     now,
		}, nil
	}

	// Folder share
	includeChildren := true
	if req.IncludeChildren != nil {
		includeChildren = *req.IncludeChildren
	}

	// Check if share already exists
	if req.SharedWithID != nil {
		existing, err := s.shareRepo.FindFolderShareByFolderAndUser(ctx, req.EntityID, *req.SharedWithID)
		if err == nil && existing != nil {
			return nil, errors.New("folder is already shared with this user")
		}
	}

	share := &domain.FolderShare{
		ID:              uuid.New(),
		FolderID:        req.EntityID,
		SharedWithID:    req.SharedWithID,
		SharedByID:      userID,
		Permission:      req.Permission,
		ShareLink:       shareLink,
		LinkExpiresAt:   linkExpiresAt,
		IsPublic:        req.IsPublic,
		IncludeChildren: includeChildren,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.shareRepo.CreateFolderShare(ctx, share); err != nil {
		return nil, err
	}

	s.logger.Info("Folder share created",
		zap.String("shareId", share.ID.String()),
		zap.String("folderId", req.EntityID.String()),
		zap.String("userId", userID.String()),
	)

	return &domain.ShareResponse{
		ID:              share.ID,
		EntityType:      domain.ShareTypeFolder,
		EntityID:        req.EntityID,
		EntityName:      entityName,
		SharedWithID:    req.SharedWithID,
		SharedByID:      userID,
		Permission:      req.Permission,
		ShareLink:       shareLink,
		LinkExpiresAt:   linkExpiresAt,
		IsPublic:        req.IsPublic,
		IsExpired:       share.IsExpired(),
		IncludeChildren: includeChildren,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// GetFileShares gets all shares for a file
func (s *ShareService) GetFileShares(ctx context.Context, fileID uuid.UUID) ([]domain.ShareResponse, error) {
	file, err := s.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("file not found")
		}
		return nil, err
	}

	shares, err := s.shareRepo.FindFileSharesByFileID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	var responses []domain.ShareResponse
	for _, share := range shares {
		responses = append(responses, domain.ShareResponse{
			ID:            share.ID,
			EntityType:    domain.ShareTypeFile,
			EntityID:      share.FileID,
			EntityName:    file.Name,
			SharedWithID:  share.SharedWithID,
			SharedByID:    share.SharedByID,
			Permission:    share.Permission,
			ShareLink:     share.ShareLink,
			LinkExpiresAt: share.LinkExpiresAt,
			IsPublic:      share.IsPublic,
			IsExpired:     share.IsExpired(),
			CreatedAt:     share.CreatedAt,
			UpdatedAt:     share.UpdatedAt,
		})
	}

	return responses, nil
}

// GetFolderShares gets all shares for a folder
func (s *ShareService) GetFolderShares(ctx context.Context, folderID uuid.UUID) ([]domain.ShareResponse, error) {
	folder, err := s.folderRepo.FindByID(ctx, folderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("folder not found")
		}
		return nil, err
	}

	shares, err := s.shareRepo.FindFolderSharesByFolderID(ctx, folderID)
	if err != nil {
		return nil, err
	}

	var responses []domain.ShareResponse
	for _, share := range shares {
		responses = append(responses, domain.ShareResponse{
			ID:              share.ID,
			EntityType:      domain.ShareTypeFolder,
			EntityID:        share.FolderID,
			EntityName:      folder.Name,
			SharedWithID:    share.SharedWithID,
			SharedByID:      share.SharedByID,
			Permission:      share.Permission,
			ShareLink:       share.ShareLink,
			LinkExpiresAt:   share.LinkExpiresAt,
			IsPublic:        share.IsPublic,
			IsExpired:       share.IsExpired(),
			IncludeChildren: share.IncludeChildren,
			CreatedAt:       share.CreatedAt,
			UpdatedAt:       share.UpdatedAt,
		})
	}

	return responses, nil
}

// GetSharedWithMe gets all items shared with the current user
func (s *ShareService) GetSharedWithMe(ctx context.Context, userID uuid.UUID) ([]domain.SharedItem, error) {
	var items []domain.SharedItem

	// Get file shares
	fileShares, err := s.shareRepo.FindFileSharesBySharedWith(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, share := range fileShares {
		if share.File != nil {
			items = append(items, domain.SharedItem{
				EntityType: domain.ShareTypeFile,
				EntityID:   share.FileID,
				EntityName: share.File.Name,
				Permission: share.Permission,
				SharedByID: share.SharedByID,
				SharedAt:   share.CreatedAt,
			})
		}
	}

	// Get folder shares
	folderShares, err := s.shareRepo.FindFolderSharesBySharedWith(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, share := range folderShares {
		if share.Folder != nil {
			items = append(items, domain.SharedItem{
				EntityType: domain.ShareTypeFolder,
				EntityID:   share.FolderID,
				EntityName: share.Folder.Name,
				Permission: share.Permission,
				SharedByID: share.SharedByID,
				SharedAt:   share.CreatedAt,
			})
		}
	}

	return items, nil
}

// GetShareByLink gets a share by its public link
func (s *ShareService) GetShareByLink(ctx context.Context, shareLink string) (*domain.ShareResponse, error) {
	// Try file share first
	fileShare, err := s.shareRepo.FindFileShareByLink(ctx, shareLink)
	if err == nil {
		if fileShare.IsExpired() {
			return nil, errors.New("share link has expired")
		}
		return &domain.ShareResponse{
			ID:            fileShare.ID,
			EntityType:    domain.ShareTypeFile,
			EntityID:      fileShare.FileID,
			EntityName:    fileShare.File.Name,
			Permission:    fileShare.Permission,
			ShareLink:     fileShare.ShareLink,
			LinkExpiresAt: fileShare.LinkExpiresAt,
			IsPublic:      fileShare.IsPublic,
			CreatedAt:     fileShare.CreatedAt,
		}, nil
	}

	// Try folder share
	folderShare, err := s.shareRepo.FindFolderShareByLink(ctx, shareLink)
	if err == nil {
		if folderShare.IsExpired() {
			return nil, errors.New("share link has expired")
		}
		return &domain.ShareResponse{
			ID:              folderShare.ID,
			EntityType:      domain.ShareTypeFolder,
			EntityID:        folderShare.FolderID,
			EntityName:      folderShare.Folder.Name,
			Permission:      folderShare.Permission,
			ShareLink:       folderShare.ShareLink,
			LinkExpiresAt:   folderShare.LinkExpiresAt,
			IsPublic:        folderShare.IsPublic,
			IncludeChildren: folderShare.IncludeChildren,
			CreatedAt:       folderShare.CreatedAt,
		}, nil
	}

	return nil, errors.New("share link not found")
}

// UpdateShare updates a share's permission
func (s *ShareService) UpdateShare(ctx context.Context, shareID uuid.UUID, entityType domain.ShareType, req domain.UpdateShareRequest, userID uuid.UUID) error {
	if entityType == domain.ShareTypeFile {
		share, err := s.shareRepo.FindFileShareByID(ctx, shareID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("share not found")
			}
			return err
		}

		if share.SharedByID != userID {
			return errors.New("not authorized to update this share")
		}

		if req.Permission != nil {
			share.Permission = *req.Permission
		}

		if req.ExpiresInDays != nil {
			if *req.ExpiresInDays <= 0 {
				share.LinkExpiresAt = nil
			} else {
				expires := time.Now().AddDate(0, 0, *req.ExpiresInDays)
				share.LinkExpiresAt = &expires
			}
		}

		return s.shareRepo.UpdateFileShare(ctx, share)
	}

	share, err := s.shareRepo.FindFolderShareByID(ctx, shareID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("share not found")
		}
		return err
	}

	if share.SharedByID != userID {
		return errors.New("not authorized to update this share")
	}

	if req.Permission != nil {
		share.Permission = *req.Permission
	}

	if req.ExpiresInDays != nil {
		if *req.ExpiresInDays <= 0 {
			share.LinkExpiresAt = nil
		} else {
			expires := time.Now().AddDate(0, 0, *req.ExpiresInDays)
			share.LinkExpiresAt = &expires
		}
	}

	return s.shareRepo.UpdateFolderShare(ctx, share)
}

// DeleteShare deletes a share
func (s *ShareService) DeleteShare(ctx context.Context, shareID uuid.UUID, entityType domain.ShareType, userID uuid.UUID) error {
	if entityType == domain.ShareTypeFile {
		share, err := s.shareRepo.FindFileShareByID(ctx, shareID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("share not found")
			}
			return err
		}

		if share.SharedByID != userID {
			return errors.New("not authorized to delete this share")
		}

		return s.shareRepo.DeleteFileShare(ctx, shareID)
	}

	share, err := s.shareRepo.FindFolderShareByID(ctx, shareID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("share not found")
		}
		return err
	}

	if share.SharedByID != userID {
		return errors.New("not authorized to delete this share")
	}

	return s.shareRepo.DeleteFolderShare(ctx, shareID)
}

// CheckAccess checks if a user has access to a file or folder
func (s *ShareService) CheckAccess(ctx context.Context, entityType domain.ShareType, entityID, userID uuid.UUID) (*domain.PermissionLevel, error) {
	if entityType == domain.ShareTypeFile {
		return s.shareRepo.CheckFileAccess(ctx, entityID, userID)
	}
	return s.shareRepo.CheckFolderAccess(ctx, entityID, userID)
}

// CleanupExpiredShares removes expired share links
func (s *ShareService) CleanupExpiredShares(ctx context.Context) error {
	return s.shareRepo.CleanupExpiredShares(ctx)
}
