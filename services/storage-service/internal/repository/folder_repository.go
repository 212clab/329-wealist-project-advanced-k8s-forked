package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"storage-service/internal/domain"
)

// FolderRepository handles folder database operations
type FolderRepository struct {
	db *gorm.DB
}

// NewFolderRepository creates a new FolderRepository
func NewFolderRepository(db *gorm.DB) *FolderRepository {
	return &FolderRepository{db: db}
}

// Create creates a new folder
func (r *FolderRepository) Create(ctx context.Context, folder *domain.Folder) error {
	return r.db.WithContext(ctx).Create(folder).Error
}

// FindByID finds a folder by ID
func (r *FolderRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Folder, error) {
	var folder domain.Folder
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&folder).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

// FindByIDWithDeleted finds a folder by ID including deleted ones (for trash)
func (r *FolderRepository) FindByIDWithDeleted(ctx context.Context, id uuid.UUID) (*domain.Folder, error) {
	var folder domain.Folder
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&folder).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

// FindByWorkspaceID finds all folders in a workspace
func (r *FolderRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]domain.Folder, error) {
	var folders []domain.Folder
	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND deleted_at IS NULL", workspaceID).
		Order("path ASC, name ASC").
		Find(&folders).Error
	return folders, err
}

// FindByParentID finds all child folders of a parent
func (r *FolderRepository) FindByParentID(ctx context.Context, workspaceID uuid.UUID, parentID *uuid.UUID) ([]domain.Folder, error) {
	var folders []domain.Folder
	query := r.db.WithContext(ctx).Where("workspace_id = ? AND deleted_at IS NULL", workspaceID)

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", parentID)
	}

	err := query.Order("name ASC").Find(&folders).Error
	return folders, err
}

// FindRootFolders finds all root folders in a workspace
func (r *FolderRepository) FindRootFolders(ctx context.Context, workspaceID uuid.UUID) ([]domain.Folder, error) {
	return r.FindByParentID(ctx, workspaceID, nil)
}

// FindByPath finds a folder by its path
func (r *FolderRepository) FindByPath(ctx context.Context, workspaceID uuid.UUID, path string) (*domain.Folder, error) {
	var folder domain.Folder
	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND path = ? AND deleted_at IS NULL", workspaceID, path).
		First(&folder).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

// FindChildrenRecursive finds all descendant folders of a folder
func (r *FolderRepository) FindChildrenRecursive(ctx context.Context, workspaceID uuid.UUID, parentPath string) ([]domain.Folder, error) {
	var folders []domain.Folder
	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND path LIKE ? AND deleted_at IS NULL", workspaceID, parentPath+"/%").
		Order("path ASC").
		Find(&folders).Error
	return folders, err
}

// Update updates a folder
func (r *FolderRepository) Update(ctx context.Context, folder *domain.Folder) error {
	return r.db.WithContext(ctx).Save(folder).Error
}

// UpdatePath updates the path of a folder and all its descendants
func (r *FolderRepository) UpdatePath(ctx context.Context, workspaceID uuid.UUID, oldPath, newPath string) error {
	// Update the folder itself
	err := r.db.WithContext(ctx).
		Model(&domain.Folder{}).
		Where("workspace_id = ? AND path = ?", workspaceID, oldPath).
		Update("path", newPath).Error
	if err != nil {
		return err
	}

	// Update all descendants by replacing prefix
	return r.db.WithContext(ctx).
		Model(&domain.Folder{}).
		Where("workspace_id = ? AND path LIKE ?", workspaceID, oldPath+"/%").
		Update("path", gorm.Expr("REPLACE(path, ?, ?)", oldPath, newPath)).Error
}

// SoftDelete soft deletes a folder (move to trash)
func (r *FolderRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.Folder{}).
		Where("id = ?", id).
		Update("deleted_at", now).Error
}

// SoftDeleteByPath soft deletes a folder and all its descendants
func (r *FolderRepository) SoftDeleteByPath(ctx context.Context, workspaceID uuid.UUID, path string) error {
	now := time.Now()
	// Delete the folder itself
	err := r.db.WithContext(ctx).
		Model(&domain.Folder{}).
		Where("workspace_id = ? AND path = ?", workspaceID, path).
		Update("deleted_at", now).Error
	if err != nil {
		return err
	}

	// Delete all descendants
	return r.db.WithContext(ctx).
		Model(&domain.Folder{}).
		Where("workspace_id = ? AND path LIKE ?", workspaceID, path+"/%").
		Update("deleted_at", now).Error
}

// Restore restores a soft-deleted folder
func (r *FolderRepository) Restore(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.Folder{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

// PermanentDelete permanently deletes a folder
func (r *FolderRepository) PermanentDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&domain.Folder{}, id).Error
}

// FindDeleted finds all deleted folders in a workspace (trash)
func (r *FolderRepository) FindDeleted(ctx context.Context, workspaceID uuid.UUID) ([]domain.Folder, error) {
	var folders []domain.Folder
	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND deleted_at IS NOT NULL", workspaceID).
		Order("deleted_at DESC").
		Find(&folders).Error
	return folders, err
}

// CountByWorkspaceID counts folders in a workspace
func (r *FolderRepository) CountByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Folder{}).
		Where("workspace_id = ? AND deleted_at IS NULL", workspaceID).
		Count(&count).Error
	return count, err
}

// CountByParentID counts child folders of a parent
func (r *FolderRepository) CountByParentID(ctx context.Context, parentID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Folder{}).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error
	return count, err
}

// ExistsByNameInParent checks if a folder with the given name exists in the parent
func (r *FolderRepository) ExistsByNameInParent(ctx context.Context, workspaceID uuid.UUID, parentID *uuid.UUID, name string) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.Folder{}).
		Where("workspace_id = ? AND name = ? AND deleted_at IS NULL", workspaceID, name)

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", parentID)
	}

	err := query.Count(&count).Error
	return count > 0, err
}

// GenerateUniqueName generates a unique folder name if conflict exists
func (r *FolderRepository) GenerateUniqueName(ctx context.Context, workspaceID uuid.UUID, parentID *uuid.UUID, name string) (string, error) {
	exists, err := r.ExistsByNameInParent(ctx, workspaceID, parentID, name)
	if err != nil {
		return "", err
	}
	if !exists {
		return name, nil
	}

	// Try with suffix (1), (2), etc.
	for i := 1; i <= 100; i++ {
		newName := fmt.Sprintf("%s (%d)", name, i)
		exists, err = r.ExistsByNameInParent(ctx, workspaceID, parentID, newName)
		if err != nil {
			return "", err
		}
		if !exists {
			return newName, nil
		}
	}

	return "", fmt.Errorf("could not generate unique name for folder: %s", name)
}
