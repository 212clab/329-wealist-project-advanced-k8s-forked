package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"user-service/internal/domain"
)

// JoinRequestRepository handles workspace join request data access
type JoinRequestRepository struct {
	db *gorm.DB
}

// NewJoinRequestRepository creates a new JoinRequestRepository
func NewJoinRequestRepository(db *gorm.DB) *JoinRequestRepository {
	return &JoinRequestRepository{db: db}
}

// Create creates a new join request
func (r *JoinRequestRepository) Create(request *domain.WorkspaceJoinRequest) error {
	return r.db.Create(request).Error
}

// FindByID finds a join request by ID
func (r *JoinRequestRepository) FindByID(id uuid.UUID) (*domain.WorkspaceJoinRequest, error) {
	var request domain.WorkspaceJoinRequest
	err := r.db.Preload("User").Preload("Workspace").Where("id = ?", id).First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// FindByWorkspaceAndUser finds a join request by workspace and user
func (r *JoinRequestRepository) FindByWorkspaceAndUser(workspaceID, userID uuid.UUID) (*domain.WorkspaceJoinRequest, error) {
	var request domain.WorkspaceJoinRequest
	err := r.db.Where("workspace_id = ? AND user_id = ?", workspaceID, userID).First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// FindPendingByWorkspace finds pending join requests for a workspace
func (r *JoinRequestRepository) FindPendingByWorkspace(workspaceID uuid.UUID) ([]domain.WorkspaceJoinRequest, error) {
	var requests []domain.WorkspaceJoinRequest
	err := r.db.Preload("User").
		Where("workspace_id = ? AND status = ?", workspaceID, domain.JoinStatusPending).
		Find(&requests).Error
	return requests, err
}

// FindByWorkspace finds all join requests for a workspace
func (r *JoinRequestRepository) FindByWorkspace(workspaceID uuid.UUID) ([]domain.WorkspaceJoinRequest, error) {
	var requests []domain.WorkspaceJoinRequest
	err := r.db.Preload("User").Where("workspace_id = ?", workspaceID).Find(&requests).Error
	return requests, err
}

// FindPendingByUser finds pending join requests for a user
func (r *JoinRequestRepository) FindPendingByUser(userID uuid.UUID) ([]domain.WorkspaceJoinRequest, error) {
	var requests []domain.WorkspaceJoinRequest
	err := r.db.Preload("Workspace").
		Where("user_id = ? AND status = ?", userID, domain.JoinStatusPending).
		Find(&requests).Error
	return requests, err
}

// Update updates a join request
func (r *JoinRequestRepository) Update(request *domain.WorkspaceJoinRequest) error {
	return r.db.Save(request).Error
}

// Delete deletes a join request
func (r *JoinRequestRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.WorkspaceJoinRequest{}, "id = ?", id).Error
}

// HasPendingRequest checks if user has a pending request for workspace
func (r *JoinRequestRepository) HasPendingRequest(workspaceID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&domain.WorkspaceJoinRequest{}).
		Where("workspace_id = ? AND user_id = ? AND status = ?", workspaceID, userID, domain.JoinStatusPending).
		Count(&count).Error
	return count > 0, err
}
