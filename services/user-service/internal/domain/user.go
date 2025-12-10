package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"userId"`
	Email     string     `gorm:"uniqueIndex;not null" json:"email"`
	Name      string     `gorm:"not null;default:''" json:"name"`
	GoogleID  *string    `gorm:"uniqueIndex;column:google_id" json:"googleId,omitempty"`
	Provider  string     `gorm:"default:'google'" json:"provider"`
	IsActive  bool       `gorm:"default:true" json:"isActive"`
	CreatedAt time.Time  `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"not null" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"index" json:"deletedAt,omitempty"`

	// Relations
	Workspaces        []Workspace        `gorm:"foreignKey:OwnerID" json:"workspaces,omitempty"`
	WorkspaceMembers  []WorkspaceMember  `gorm:"foreignKey:UserID" json:"workspaceMembers,omitempty"`
	Profiles          []UserProfile      `gorm:"foreignKey:UserID" json:"profiles,omitempty"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Email    string  `json:"email" binding:"required,email"`
	GoogleID *string `json:"googleId,omitempty"`
	Provider string  `json:"provider,omitempty"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Email    *string `json:"email,omitempty" binding:"omitempty,email"`
	IsActive *bool   `json:"isActive,omitempty"`
}

// OAuthLoginRequest represents the request for OAuth login (internal API)
type OAuthLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required"`
	Provider string `json:"provider" binding:"required"`
}

// UserResponse represents the user response
type UserResponse struct {
	UserID    uuid.UUID  `json:"userId"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	GoogleID  *string    `json:"googleId,omitempty"`
	Provider  string     `json:"provider"`
	IsActive  bool       `json:"isActive"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		UserID:    u.ID,
		Email:     u.Email,
		Name:      u.Name,
		GoogleID:  u.GoogleID,
		Provider:  u.Provider,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		DeletedAt: u.DeletedAt,
	}
}
