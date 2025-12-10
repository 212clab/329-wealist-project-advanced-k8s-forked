package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"storage-service/internal/domain"
)

// FileRepository handles file database operations
type FileRepository struct {
	db *gorm.DB
}

// NewFileRepository creates a new FileRepository
func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

// Create creates a new file record
func (r *FileRepository) Create(ctx context.Context, file *domain.File) error {
	return r.db.WithContext(ctx).Create(file).Error
}

// FindByID finds a file by ID
func (r *FileRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.File, error) {
	var file domain.File
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// FindByIDWithDeleted finds a file by ID including deleted ones
func (r *FileRepository) FindByIDWithDeleted(ctx context.Context, id uuid.UUID) (*domain.File, error) {
	var file domain.File
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// FindByFileKey finds a file by S3 key
func (r *FileRepository) FindByFileKey(ctx context.Context, fileKey string) (*domain.File, error) {
	var file domain.File
	err := r.db.WithContext(ctx).
		Where("file_key = ? AND deleted_at IS NULL", fileKey).
		First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// FindByWorkspaceID finds all files in a workspace with pagination
func (r *FileRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, page, pageSize int) ([]domain.File, int64, error) {
	var files []domain.File
	var total int64

	// Count total
	err := r.db.WithContext(ctx).
		Model(&domain.File{}).
		Where("workspace_id = ? AND deleted_at IS NULL", workspaceID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get files with pagination
	offset := (page - 1) * pageSize
	err = r.db.WithContext(ctx).
		Where("workspace_id = ? AND deleted_at IS NULL", workspaceID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&files).Error

	return files, total, err
}

// FindByFolderID finds all files in a folder
func (r *FileRepository) FindByFolderID(ctx context.Context, workspaceID uuid.UUID, folderID *uuid.UUID) ([]domain.File, error) {
	var files []domain.File
	query := r.db.WithContext(ctx).Where("workspace_id = ? AND deleted_at IS NULL", workspaceID)

	if folderID == nil {
		query = query.Where("folder_id IS NULL")
	} else {
		query = query.Where("folder_id = ?", folderID)
	}

	err := query.Order("name ASC").Find(&files).Error
	return files, err
}

// FindRootFiles finds all files in the root folder
func (r *FileRepository) FindRootFiles(ctx context.Context, workspaceID uuid.UUID) ([]domain.File, error) {
	return r.FindByFolderID(ctx, workspaceID, nil)
}

// FindByUploadedBy finds all files uploaded by a user
func (r *FileRepository) FindByUploadedBy(ctx context.Context, workspaceID, userID uuid.UUID) ([]domain.File, error) {
	var files []domain.File
	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND uploaded_by = ? AND deleted_at IS NULL", workspaceID, userID).
		Order("created_at DESC").
		Find(&files).Error
	return files, err
}

// FindByStatus finds files by status
func (r *FileRepository) FindByStatus(ctx context.Context, status domain.FileStatus) ([]domain.File, error) {
	var files []domain.File
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Find(&files).Error
	return files, err
}

// FindUploadingFiles finds all files that are still in uploading status
func (r *FileRepository) FindUploadingFiles(ctx context.Context, olderThan time.Duration) ([]domain.File, error) {
	var files []domain.File
	cutoff := time.Now().Add(-olderThan)
	err := r.db.WithContext(ctx).
		Where("status = ? AND created_at < ?", domain.FileStatusUploading, cutoff).
		Find(&files).Error
	return files, err
}

// Update updates a file record
func (r *FileRepository) Update(ctx context.Context, file *domain.File) error {
	return r.db.WithContext(ctx).Save(file).Error
}

// UpdateStatus updates the status of a file
func (r *FileRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.FileStatus) error {
	return r.db.WithContext(ctx).
		Model(&domain.File{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// SoftDelete soft deletes a file (move to trash)
func (r *FileRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.File{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_at": now,
			"status":     domain.FileStatusDeleted,
		}).Error
}

// SoftDeleteByFolderID soft deletes all files in a folder
func (r *FileRepository) SoftDeleteByFolderID(ctx context.Context, folderID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.File{}).
		Where("folder_id = ?", folderID).
		Updates(map[string]interface{}{
			"deleted_at": now,
			"status":     domain.FileStatusDeleted,
		}).Error
}

// Restore restores a soft-deleted file
func (r *FileRepository) Restore(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.File{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_at": nil,
			"status":     domain.FileStatusActive,
		}).Error
}

// PermanentDelete permanently deletes a file record
func (r *FileRepository) PermanentDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&domain.File{}, id).Error
}

// FindDeleted finds all deleted files in a workspace (trash)
func (r *FileRepository) FindDeleted(ctx context.Context, workspaceID uuid.UUID) ([]domain.File, error) {
	var files []domain.File
	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND deleted_at IS NOT NULL", workspaceID).
		Order("deleted_at DESC").
		Find(&files).Error
	return files, err
}

// CountByWorkspaceID counts files in a workspace
func (r *FileRepository) CountByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.File{}).
		Where("workspace_id = ? AND deleted_at IS NULL", workspaceID).
		Count(&count).Error
	return count, err
}

// CountByFolderID counts files in a folder
func (r *FileRepository) CountByFolderID(ctx context.Context, folderID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.File{}).
		Where("folder_id = ? AND deleted_at IS NULL", folderID).
		Count(&count).Error
	return count, err
}

// SumSizeByWorkspaceID calculates total size of files in a workspace
func (r *FileRepository) SumSizeByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	var result struct {
		TotalSize int64
	}
	err := r.db.WithContext(ctx).
		Model(&domain.File{}).
		Select("COALESCE(SUM(file_size), 0) as total_size").
		Where("workspace_id = ? AND deleted_at IS NULL", workspaceID).
		Scan(&result).Error
	return result.TotalSize, err
}

// SumSizeByFolderID calculates total size of files in a folder
func (r *FileRepository) SumSizeByFolderID(ctx context.Context, folderID uuid.UUID) (int64, error) {
	var result struct {
		TotalSize int64
	}
	err := r.db.WithContext(ctx).
		Model(&domain.File{}).
		Select("COALESCE(SUM(file_size), 0) as total_size").
		Where("folder_id = ? AND deleted_at IS NULL", folderID).
		Scan(&result).Error
	return result.TotalSize, err
}

// ExistsByNameInFolder checks if a file with the given name exists in the folder
func (r *FileRepository) ExistsByNameInFolder(ctx context.Context, workspaceID uuid.UUID, folderID *uuid.UUID, name string) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.File{}).
		Where("workspace_id = ? AND name = ? AND deleted_at IS NULL", workspaceID, name)

	if folderID == nil {
		query = query.Where("folder_id IS NULL")
	} else {
		query = query.Where("folder_id = ?", folderID)
	}

	err := query.Count(&count).Error
	return count > 0, err
}

// GenerateUniqueName generates a unique file name if conflict exists
func (r *FileRepository) GenerateUniqueName(ctx context.Context, workspaceID uuid.UUID, folderID *uuid.UUID, name string) (string, error) {
	exists, err := r.ExistsByNameInFolder(ctx, workspaceID, folderID, name)
	if err != nil {
		return "", err
	}
	if !exists {
		return name, nil
	}

	// Try with suffix (1), (2), etc.
	for i := 1; i <= 100; i++ {
		newName := fmt.Sprintf("%s (%d)", name, i)
		exists, err = r.ExistsByNameInFolder(ctx, workspaceID, folderID, newName)
		if err != nil {
			return "", err
		}
		if !exists {
			return newName, nil
		}
	}

	return "", fmt.Errorf("could not generate unique name for file: %s", name)
}

// Search searches files by name
func (r *FileRepository) Search(ctx context.Context, workspaceID uuid.UUID, query string, page, pageSize int) ([]domain.File, int64, error) {
	var files []domain.File
	var total int64

	searchQuery := "%" + query + "%"

	// Count total
	err := r.db.WithContext(ctx).
		Model(&domain.File{}).
		Where("workspace_id = ? AND (name ILIKE ? OR original_name ILIKE ?) AND deleted_at IS NULL",
			workspaceID, searchQuery, searchQuery).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get files with pagination
	offset := (page - 1) * pageSize
	err = r.db.WithContext(ctx).
		Where("workspace_id = ? AND (name ILIKE ? OR original_name ILIKE ?) AND deleted_at IS NULL",
			workspaceID, searchQuery, searchQuery).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&files).Error

	return files, total, err
}
