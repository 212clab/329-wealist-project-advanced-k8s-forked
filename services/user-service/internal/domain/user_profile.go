package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserProfile represents a user's profile in a specific workspace
type UserProfile struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"profileId"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"userId"`
	WorkspaceID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"workspaceId"`
	NickName        string     `gorm:"not null" json:"nickName"`
	Email           string     `gorm:"not null" json:"email"`
	ProfileImageURL *string    `gorm:"column:profile_image_url" json:"profileImageUrl,omitempty"`
	CreatedAt       time.Time  `gorm:"not null" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"not null" json:"updatedAt"`

	// Relations
	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Workspace *Workspace `gorm:"foreignKey:WorkspaceID" json:"workspace,omitempty"`
}

// TableName specifies the table name for UserProfile
func (UserProfile) TableName() string {
	return "user_profiles"
}

// CreateProfileRequest represents the request to create a profile
type CreateProfileRequest struct {
	WorkspaceID     uuid.UUID `json:"workspaceId" binding:"required"`
	NickName        string    `json:"nickName" binding:"required"`
	Email           string    `json:"email" binding:"required,email"`
	ProfileImageURL *string   `json:"profileImageUrl,omitempty"`
}

// UpdateProfileRequest represents the request to update a profile
type UpdateProfileRequest struct {
	NickName        *string `json:"nickName,omitempty"`
	ProfileImageURL *string `json:"profileImageUrl,omitempty"`
}

// UserProfileResponse represents the user profile response
type UserProfileResponse struct {
	ProfileID       uuid.UUID `json:"profileId"`
	UserID          uuid.UUID `json:"userId"`
	WorkspaceID     uuid.UUID `json:"workspaceId"`
	NickName        string    `json:"nickName"`
	Email           string    `json:"email"`
	ProfileImageURL *string   `json:"profileImageUrl,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// ToResponse converts UserProfile to UserProfileResponse
func (p *UserProfile) ToResponse() UserProfileResponse {
	return UserProfileResponse{
		ProfileID:       p.ID,
		UserID:          p.UserID,
		WorkspaceID:     p.WorkspaceID,
		NickName:        p.NickName,
		Email:           p.Email,
		ProfileImageURL: p.ProfileImageURL,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}
