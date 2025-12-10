package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"user-service/internal/domain"
)

// UserProfileRepository handles user profile data access
type UserProfileRepository struct {
	db *gorm.DB
}

// NewUserProfileRepository creates a new UserProfileRepository
func NewUserProfileRepository(db *gorm.DB) *UserProfileRepository {
	return &UserProfileRepository{db: db}
}

// Create creates a new user profile
func (r *UserProfileRepository) Create(profile *domain.UserProfile) error {
	return r.db.Create(profile).Error
}

// FindByID finds a user profile by ID
func (r *UserProfileRepository) FindByID(id uuid.UUID) (*domain.UserProfile, error) {
	var profile domain.UserProfile
	err := r.db.Where("id = ?", id).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// FindByUserAndWorkspace finds a profile by user and workspace
func (r *UserProfileRepository) FindByUserAndWorkspace(userID, workspaceID uuid.UUID) (*domain.UserProfile, error) {
	var profile domain.UserProfile
	err := r.db.Where("user_id = ? AND workspace_id = ?", userID, workspaceID).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// FindByUser finds all profiles for a user
func (r *UserProfileRepository) FindByUser(userID uuid.UUID) ([]domain.UserProfile, error) {
	var profiles []domain.UserProfile
	err := r.db.Preload("Workspace").Where("user_id = ?", userID).Find(&profiles).Error
	return profiles, err
}

// FindByWorkspace finds all profiles in a workspace
func (r *UserProfileRepository) FindByWorkspace(workspaceID uuid.UUID) ([]domain.UserProfile, error) {
	var profiles []domain.UserProfile
	err := r.db.Where("workspace_id = ?", workspaceID).Find(&profiles).Error
	return profiles, err
}

// Update updates a user profile
func (r *UserProfileRepository) Update(profile *domain.UserProfile) error {
	return r.db.Save(profile).Error
}

// Delete deletes a user profile
func (r *UserProfileRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.UserProfile{}, "id = ?", id).Error
}

// DeleteByUserAndWorkspace deletes a profile by user and workspace
func (r *UserProfileRepository) DeleteByUserAndWorkspace(userID, workspaceID uuid.UUID) error {
	return r.db.Delete(&domain.UserProfile{}, "user_id = ? AND workspace_id = ?", userID, workspaceID).Error
}
