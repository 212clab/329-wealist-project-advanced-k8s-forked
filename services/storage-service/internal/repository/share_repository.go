package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"storage-service/internal/domain"
)

// ShareRepository handles share database operations
type ShareRepository struct {
	db *gorm.DB
}

// NewShareRepository creates a new ShareRepository
func NewShareRepository(db *gorm.DB) *ShareRepository {
	return &ShareRepository{db: db}
}

// ============================================================
// File Share Operations
// ============================================================

// CreateFileShare creates a new file share
func (r *ShareRepository) CreateFileShare(ctx context.Context, share *domain.FileShare) error {
	return r.db.WithContext(ctx).Create(share).Error
}

// FindFileShareByID finds a file share by ID
func (r *ShareRepository) FindFileShareByID(ctx context.Context, id uuid.UUID) (*domain.FileShare, error) {
	var share domain.FileShare
	err := r.db.WithContext(ctx).
		Preload("File").
		Where("id = ?", id).
		First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

// FindFileShareByLink finds a file share by share link
func (r *ShareRepository) FindFileShareByLink(ctx context.Context, shareLink string) (*domain.FileShare, error) {
	var share domain.FileShare
	err := r.db.WithContext(ctx).
		Preload("File").
		Where("share_link = ?", shareLink).
		First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

// FindFileSharesByFileID finds all shares for a file
func (r *ShareRepository) FindFileSharesByFileID(ctx context.Context, fileID uuid.UUID) ([]domain.FileShare, error) {
	var shares []domain.FileShare
	err := r.db.WithContext(ctx).
		Where("file_id = ?", fileID).
		Order("created_at DESC").
		Find(&shares).Error
	return shares, err
}

// FindFileSharesBySharedWith finds all file shares with a specific user
func (r *ShareRepository) FindFileSharesBySharedWith(ctx context.Context, userID uuid.UUID) ([]domain.FileShare, error) {
	var shares []domain.FileShare
	err := r.db.WithContext(ctx).
		Preload("File").
		Where("shared_with_id = ?", userID).
		Order("created_at DESC").
		Find(&shares).Error
	return shares, err
}

// FindFileShareByFileAndUser finds a share for a specific file and user
func (r *ShareRepository) FindFileShareByFileAndUser(ctx context.Context, fileID, userID uuid.UUID) (*domain.FileShare, error) {
	var share domain.FileShare
	err := r.db.WithContext(ctx).
		Where("file_id = ? AND shared_with_id = ?", fileID, userID).
		First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

// UpdateFileShare updates a file share
func (r *ShareRepository) UpdateFileShare(ctx context.Context, share *domain.FileShare) error {
	share.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(share).Error
}

// DeleteFileShare deletes a file share
func (r *ShareRepository) DeleteFileShare(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.FileShare{}, id).Error
}

// DeleteFileSharesByFileID deletes all shares for a file
func (r *ShareRepository) DeleteFileSharesByFileID(ctx context.Context, fileID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("file_id = ?", fileID).Delete(&domain.FileShare{}).Error
}

// ============================================================
// Folder Share Operations
// ============================================================

// CreateFolderShare creates a new folder share
func (r *ShareRepository) CreateFolderShare(ctx context.Context, share *domain.FolderShare) error {
	return r.db.WithContext(ctx).Create(share).Error
}

// FindFolderShareByID finds a folder share by ID
func (r *ShareRepository) FindFolderShareByID(ctx context.Context, id uuid.UUID) (*domain.FolderShare, error) {
	var share domain.FolderShare
	err := r.db.WithContext(ctx).
		Preload("Folder").
		Where("id = ?", id).
		First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

// FindFolderShareByLink finds a folder share by share link
func (r *ShareRepository) FindFolderShareByLink(ctx context.Context, shareLink string) (*domain.FolderShare, error) {
	var share domain.FolderShare
	err := r.db.WithContext(ctx).
		Preload("Folder").
		Where("share_link = ?", shareLink).
		First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

// FindFolderSharesByFolderID finds all shares for a folder
func (r *ShareRepository) FindFolderSharesByFolderID(ctx context.Context, folderID uuid.UUID) ([]domain.FolderShare, error) {
	var shares []domain.FolderShare
	err := r.db.WithContext(ctx).
		Where("folder_id = ?", folderID).
		Order("created_at DESC").
		Find(&shares).Error
	return shares, err
}

// FindFolderSharesBySharedWith finds all folder shares with a specific user
func (r *ShareRepository) FindFolderSharesBySharedWith(ctx context.Context, userID uuid.UUID) ([]domain.FolderShare, error) {
	var shares []domain.FolderShare
	err := r.db.WithContext(ctx).
		Preload("Folder").
		Where("shared_with_id = ?", userID).
		Order("created_at DESC").
		Find(&shares).Error
	return shares, err
}

// FindFolderShareByFolderAndUser finds a share for a specific folder and user
func (r *ShareRepository) FindFolderShareByFolderAndUser(ctx context.Context, folderID, userID uuid.UUID) (*domain.FolderShare, error) {
	var share domain.FolderShare
	err := r.db.WithContext(ctx).
		Where("folder_id = ? AND shared_with_id = ?", folderID, userID).
		First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

// UpdateFolderShare updates a folder share
func (r *ShareRepository) UpdateFolderShare(ctx context.Context, share *domain.FolderShare) error {
	share.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(share).Error
}

// DeleteFolderShare deletes a folder share
func (r *ShareRepository) DeleteFolderShare(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.FolderShare{}, id).Error
}

// DeleteFolderSharesByFolderID deletes all shares for a folder
func (r *ShareRepository) DeleteFolderSharesByFolderID(ctx context.Context, folderID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("folder_id = ?", folderID).Delete(&domain.FolderShare{}).Error
}

// ============================================================
// Common Operations
// ============================================================

// CheckFileAccess checks if a user has access to a file (via direct share or folder share)
func (r *ShareRepository) CheckFileAccess(ctx context.Context, fileID, userID uuid.UUID) (*domain.PermissionLevel, error) {
	// Check direct file share
	var fileShare domain.FileShare
	err := r.db.WithContext(ctx).
		Where("file_id = ? AND shared_with_id = ?", fileID, userID).
		First(&fileShare).Error
	if err == nil {
		return &fileShare.Permission, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Check if file's folder is shared
	var file domain.File
	err = r.db.WithContext(ctx).Where("id = ?", fileID).First(&file).Error
	if err != nil {
		return nil, err
	}

	if file.FolderID == nil {
		return nil, gorm.ErrRecordNotFound
	}

	var folderShare domain.FolderShare
	err = r.db.WithContext(ctx).
		Where("folder_id = ? AND shared_with_id = ?", file.FolderID, userID).
		First(&folderShare).Error
	if err == nil && folderShare.IncludeChildren {
		return &folderShare.Permission, nil
	}

	return nil, gorm.ErrRecordNotFound
}

// CheckFolderAccess checks if a user has access to a folder
func (r *ShareRepository) CheckFolderAccess(ctx context.Context, folderID, userID uuid.UUID) (*domain.PermissionLevel, error) {
	var share domain.FolderShare
	err := r.db.WithContext(ctx).
		Where("folder_id = ? AND shared_with_id = ?", folderID, userID).
		First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share.Permission, nil
}

// GetSharedWithMeCount gets count of items shared with a user
func (r *ShareRepository) GetSharedWithMeCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var fileCount, folderCount int64

	err := r.db.WithContext(ctx).
		Model(&domain.FileShare{}).
		Where("shared_with_id = ?", userID).
		Count(&fileCount).Error
	if err != nil {
		return 0, err
	}

	err = r.db.WithContext(ctx).
		Model(&domain.FolderShare{}).
		Where("shared_with_id = ?", userID).
		Count(&folderCount).Error
	if err != nil {
		return 0, err
	}

	return fileCount + folderCount, nil
}

// CleanupExpiredShares removes expired share links
func (r *ShareRepository) CleanupExpiredShares(ctx context.Context) error {
	now := time.Now()

	// Clean up expired file shares
	err := r.db.WithContext(ctx).
		Where("link_expires_at IS NOT NULL AND link_expires_at < ?", now).
		Delete(&domain.FileShare{}).Error
	if err != nil {
		return err
	}

	// Clean up expired folder shares
	return r.db.WithContext(ctx).
		Where("link_expires_at IS NOT NULL AND link_expires_at < ?", now).
		Delete(&domain.FolderShare{}).Error
}
