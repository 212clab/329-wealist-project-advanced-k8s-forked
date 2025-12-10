package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"user-service/internal/domain"
)

// UserRepository handles user data access
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("id = ? AND is_active = true AND deleted_at IS NULL", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByIDIncludeDeleted finds a user by ID including soft deleted
func (r *UserRepository) FindByIDIncludeDeleted(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("email = ? AND is_active = true AND deleted_at IS NULL", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByGoogleID finds a user by Google ID
func (r *UserRepository) FindByGoogleID(googleID string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("google_id = ? AND is_active = true AND deleted_at IS NULL", googleID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

// SoftDelete soft deletes a user
func (r *UserRepository) SoftDelete(id uuid.UUID) error {
	return r.db.Model(&domain.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_active":  false,
			"deleted_at": gorm.Expr("NOW()"),
		}).Error
}

// Restore restores a soft deleted user
func (r *UserRepository) Restore(id uuid.UUID) error {
	return r.db.Model(&domain.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_active":  true,
			"deleted_at": nil,
		}).Error
}

// Exists checks if a user exists
func (r *UserRepository) Exists(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&domain.User{}).Where("id = ? AND is_active = true AND deleted_at IS NULL", id).Count(&count).Error
	return count > 0, err
}
