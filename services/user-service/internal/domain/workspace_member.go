package domain

import (
	"time"

	"github.com/google/uuid"
)

// RoleName represents the role of a workspace member
type RoleName string

const (
	RoleOwner  RoleName = "OWNER"
	RoleAdmin  RoleName = "ADMIN"
	RoleMember RoleName = "MEMBER"
)

// WorkspaceMember represents a member of a workspace
type WorkspaceMember struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"workspaceMemberId"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;not null;index" json:"workspaceId"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	RoleName    RoleName  `gorm:"type:varchar(20);not null;default:'MEMBER'" json:"roleName"`
	IsDefault   bool      `gorm:"default:false" json:"isDefault"`
	IsActive    bool      `gorm:"default:true" json:"isActive"`
	JoinedAt    time.Time `gorm:"not null" json:"joinedAt"`
	UpdatedAt   time.Time `gorm:"not null" json:"updatedAt"`

	// Relations
	Workspace *Workspace `gorm:"foreignKey:WorkspaceID" json:"workspace,omitempty"`
	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name for WorkspaceMember
func (WorkspaceMember) TableName() string {
	return "workspace_members"
}

// InviteMemberRequest represents the request to invite a member
type InviteMemberRequest struct {
	Email    string   `json:"email" binding:"required,email"`
	RoleName RoleName `json:"roleName,omitempty"`
}

// UpdateMemberRoleRequest represents the request to update member role
type UpdateMemberRoleRequest struct {
	RoleName RoleName `json:"roleName" binding:"required"`
}

// WorkspaceMemberResponse represents the workspace member response
type WorkspaceMemberResponse struct {
	WorkspaceMemberID uuid.UUID `json:"workspaceMemberId"`
	WorkspaceID       uuid.UUID `json:"workspaceId"`
	UserID            uuid.UUID `json:"userId"`
	RoleName          RoleName  `json:"roleName"`
	IsDefault         bool      `json:"isDefault"`
	IsActive          bool      `json:"isActive"`
	JoinedAt          time.Time `json:"joinedAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	UserEmail         string    `json:"userEmail,omitempty"`
	NickName          string    `json:"nickName,omitempty"`
	ProfileImageUrl   string    `json:"profileImageUrl,omitempty"`
}

// ToResponse converts WorkspaceMember to WorkspaceMemberResponse
func (m *WorkspaceMember) ToResponse() WorkspaceMemberResponse {
	resp := WorkspaceMemberResponse{
		WorkspaceMemberID: m.ID,
		WorkspaceID:       m.WorkspaceID,
		UserID:            m.UserID,
		RoleName:          m.RoleName,
		IsDefault:         m.IsDefault,
		IsActive:          m.IsActive,
		JoinedAt:          m.JoinedAt,
		UpdatedAt:         m.UpdatedAt,
	}
	if m.User != nil {
		resp.UserEmail = m.User.Email
	}
	return resp
}
