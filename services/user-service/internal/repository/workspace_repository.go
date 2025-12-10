package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"user-service/internal/domain"
)

// WorkspaceRepository handles workspace data access
type WorkspaceRepository struct {
	db *gorm.DB
}

// NewWorkspaceRepository creates a new WorkspaceRepository
func NewWorkspaceRepository(db *gorm.DB) *WorkspaceRepository {
	return &WorkspaceRepository{db: db}
}

// Create creates a new workspace
func (r *WorkspaceRepository) Create(workspace *domain.Workspace) error {
	return r.db.Create(workspace).Error
}

// FindByID finds a workspace by ID
func (r *WorkspaceRepository) FindByID(id uuid.UUID) (*domain.Workspace, error) {
	var workspace domain.Workspace
	err := r.db.Where("id = ? AND is_active = true AND deleted_at IS NULL", id).First(&workspace).Error
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

// FindByIDWithOwner finds a workspace by ID with owner
func (r *WorkspaceRepository) FindByIDWithOwner(id uuid.UUID) (*domain.Workspace, error) {
	var workspace domain.Workspace
	err := r.db.Preload("Owner").Where("id = ? AND is_active = true AND deleted_at IS NULL", id).First(&workspace).Error
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

// FindByOwnerID finds workspaces by owner ID
func (r *WorkspaceRepository) FindByOwnerID(ownerID uuid.UUID) ([]domain.Workspace, error) {
	var workspaces []domain.Workspace
	err := r.db.Where("owner_id = ? AND is_active = true AND deleted_at IS NULL", ownerID).Find(&workspaces).Error
	return workspaces, err
}

// FindPublicByName finds public workspaces by name (search)
func (r *WorkspaceRepository) FindPublicByName(name string) ([]domain.Workspace, error) {
	var workspaces []domain.Workspace
	err := r.db.Where("workspace_name ILIKE ? AND is_public = true AND is_active = true AND deleted_at IS NULL", "%"+name+"%").
		Limit(20).
		Find(&workspaces).Error
	return workspaces, err
}

// Update updates a workspace
func (r *WorkspaceRepository) Update(workspace *domain.Workspace) error {
	return r.db.Save(workspace).Error
}

// SoftDelete soft deletes a workspace
func (r *WorkspaceRepository) SoftDelete(id uuid.UUID) error {
	return r.db.Model(&domain.Workspace{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_active":  false,
			"deleted_at": gorm.Expr("NOW()"),
		}).Error
}

// Exists checks if a workspace exists
func (r *WorkspaceRepository) Exists(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Workspace{}).Where("id = ? AND is_active = true AND deleted_at IS NULL", id).Count(&count).Error
	return count > 0, err
}
