package domain

import (
	"time"

	"github.com/google/uuid"
)

// Workspace represents a workspace
type Workspace struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"workspaceId"`
	OwnerID              uuid.UUID  `gorm:"type:uuid;not null;index" json:"ownerId"`
	WorkspaceName        string     `gorm:"not null" json:"workspaceName"`
	WorkspaceDescription *string    `json:"workspaceDescription,omitempty"`
	IsPublic             bool       `gorm:"default:true" json:"isPublic"`
	NeedApproved         bool       `gorm:"default:true" json:"needApproved"`
	OnlyOwnerCanInvite   bool       `gorm:"default:true" json:"onlyOwnerCanInvite"`
	IsActive             bool       `gorm:"default:true" json:"isActive"`
	CreatedAt            time.Time  `gorm:"not null" json:"createdAt"`
	DeletedAt            *time.Time `gorm:"index" json:"deletedAt,omitempty"`

	// Relations
	Owner        *User                  `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Members      []WorkspaceMember      `gorm:"foreignKey:WorkspaceID" json:"members,omitempty"`
	Profiles     []UserProfile          `gorm:"foreignKey:WorkspaceID" json:"profiles,omitempty"`
	JoinRequests []WorkspaceJoinRequest `gorm:"foreignKey:WorkspaceID" json:"joinRequests,omitempty"`
}

// TableName specifies the table name for Workspace
func (Workspace) TableName() string {
	return "workspaces"
}

// CreateWorkspaceRequest represents the request to create a workspace
type CreateWorkspaceRequest struct {
	WorkspaceName        string  `json:"workspaceName" binding:"required"`
	WorkspaceDescription *string `json:"workspaceDescription,omitempty"`
	IsPublic             *bool   `json:"isPublic,omitempty"`
	NeedApproved         *bool   `json:"needApproved,omitempty"`
}

// UpdateWorkspaceRequest represents the request to update a workspace
type UpdateWorkspaceRequest struct {
	WorkspaceName        *string `json:"workspaceName,omitempty"`
	WorkspaceDescription *string `json:"workspaceDescription,omitempty"`
	IsPublic             *bool   `json:"isPublic,omitempty"`
	NeedApproved         *bool   `json:"needApproved,omitempty"`
}

// WorkspaceResponse represents the workspace response
type WorkspaceResponse struct {
	WorkspaceID          uuid.UUID  `json:"workspaceId"`
	OwnerID              uuid.UUID  `json:"ownerId"`
	WorkspaceName        string     `json:"workspaceName"`
	WorkspaceDescription *string    `json:"workspaceDescription,omitempty"`
	IsPublic             bool       `json:"isPublic"`
	NeedApproved         bool       `json:"needApproved"`
	IsActive             bool       `json:"isActive"`
	CreatedAt            time.Time  `json:"createdAt"`
	DeletedAt            *time.Time `json:"deletedAt,omitempty"`
	OwnerNickName        string     `json:"ownerNickName,omitempty"`
	OwnerEmail           string     `json:"ownerEmail,omitempty"`
}

// ToResponse converts Workspace to WorkspaceResponse
func (w *Workspace) ToResponse() WorkspaceResponse {
	resp := WorkspaceResponse{
		WorkspaceID:          w.ID,
		OwnerID:              w.OwnerID,
		WorkspaceName:        w.WorkspaceName,
		WorkspaceDescription: w.WorkspaceDescription,
		IsPublic:             w.IsPublic,
		NeedApproved:         w.NeedApproved,
		IsActive:             w.IsActive,
		CreatedAt:            w.CreatedAt,
		DeletedAt:            w.DeletedAt,
	}
	if w.Owner != nil {
		resp.OwnerEmail = w.Owner.Email
	}
	return resp
}

// ToResponseWithOwnerProfile converts Workspace to WorkspaceResponse with owner profile info
func (w *Workspace) ToResponseWithOwnerProfile(ownerNickName string) WorkspaceResponse {
	resp := w.ToResponse()
	resp.OwnerNickName = ownerNickName
	return resp
}

// UserWorkspaceResponse represents workspace info for user's workspace list
// This is used in GET /api/workspaces/all endpoint
type UserWorkspaceResponse struct {
	WorkspaceID          uuid.UUID `json:"workspaceId"`
	WorkspaceName        string    `json:"workspaceName"`
	WorkspaceDescription string    `json:"workspaceDescription"`
	Owner                bool      `json:"owner"`
	Role                 string    `json:"role"`
	CreatedAt            time.Time `json:"createdAt"`
}

// WorkspaceSettingsResponse represents workspace settings
// This is used in GET /api/workspaces/{workspaceId}/settings endpoint
type WorkspaceSettingsResponse struct {
	WorkspaceID          uuid.UUID `json:"workspaceId"`
	WorkspaceName        string    `json:"workspaceName"`
	WorkspaceDescription string    `json:"workspaceDescription"`
	IsPublic             bool      `json:"isPublic"`
	RequiresApproval     bool      `json:"requiresApproval"`
	OnlyOwnerCanInvite   bool      `json:"onlyOwnerCanInvite"`
}

// ToSettingsResponse converts Workspace to WorkspaceSettingsResponse
func (w *Workspace) ToSettingsResponse() WorkspaceSettingsResponse {
	description := ""
	if w.WorkspaceDescription != nil {
		description = *w.WorkspaceDescription
	}
	return WorkspaceSettingsResponse{
		WorkspaceID:          w.ID,
		WorkspaceName:        w.WorkspaceName,
		WorkspaceDescription: description,
		IsPublic:             w.IsPublic,
		RequiresApproval:     w.NeedApproved,
		OnlyOwnerCanInvite:   w.OnlyOwnerCanInvite,
	}
}

// UpdateWorkspaceSettingsRequest represents the request to update workspace settings
type UpdateWorkspaceSettingsRequest struct {
	WorkspaceName        *string `json:"workspaceName,omitempty"`
	WorkspaceDescription *string `json:"workspaceDescription,omitempty"`
	IsPublic             *bool   `json:"isPublic,omitempty"`
	RequiresApproval     *bool   `json:"requiresApproval,omitempty"`
	OnlyOwnerCanInvite   *bool   `json:"onlyOwnerCanInvite,omitempty"`
}
