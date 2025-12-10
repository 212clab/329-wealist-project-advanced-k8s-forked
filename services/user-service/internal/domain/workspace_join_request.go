package domain

import (
	"time"

	"github.com/google/uuid"
)

// JoinRequestStatus represents the status of a join request
type JoinRequestStatus string

const (
	JoinStatusPending  JoinRequestStatus = "PENDING"
	JoinStatusApproved JoinRequestStatus = "APPROVED"
	JoinStatusRejected JoinRequestStatus = "REJECTED"
)

// WorkspaceJoinRequest represents a request to join a workspace
type WorkspaceJoinRequest struct {
	ID          uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"joinRequestId"`
	WorkspaceID uuid.UUID         `gorm:"type:uuid;not null;index" json:"workspaceId"`
	UserID      uuid.UUID         `gorm:"type:uuid;not null;index" json:"userId"`
	Status      JoinRequestStatus `gorm:"type:varchar(20);not null;default:'PENDING'" json:"status"`
	RequestedAt time.Time         `gorm:"not null" json:"requestedAt"`
	UpdatedAt   time.Time         `gorm:"not null" json:"updatedAt"`

	// Relations
	Workspace *Workspace `gorm:"foreignKey:WorkspaceID" json:"workspace,omitempty"`
	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name for WorkspaceJoinRequest
func (WorkspaceJoinRequest) TableName() string {
	return "workspace_join_requests"
}

// CreateJoinRequestRequest represents the request to create a join request
type CreateJoinRequestRequest struct {
	WorkspaceID uuid.UUID `json:"workspaceId" binding:"required"`
}

// ProcessJoinRequestRequest represents the request to process a join request
type ProcessJoinRequestRequest struct {
	Status JoinRequestStatus `json:"status" binding:"required"`
}

// JoinRequestResponse represents the join request response
type JoinRequestResponse struct {
	JoinRequestID uuid.UUID         `json:"joinRequestId"`
	WorkspaceID   uuid.UUID         `json:"workspaceId"`
	UserID        uuid.UUID         `json:"userId"`
	Status        JoinRequestStatus `json:"status"`
	RequestedAt   time.Time         `json:"requestedAt"`
	UpdatedAt     time.Time         `json:"updatedAt"`
	UserEmail     string            `json:"userEmail,omitempty"`
	NickName      string            `json:"nickName,omitempty"`
	WorkspaceName string            `json:"workspaceName,omitempty"`
}

// ToResponse converts WorkspaceJoinRequest to JoinRequestResponse
func (r *WorkspaceJoinRequest) ToResponse() JoinRequestResponse {
	resp := JoinRequestResponse{
		JoinRequestID: r.ID,
		WorkspaceID:   r.WorkspaceID,
		UserID:        r.UserID,
		Status:        r.Status,
		RequestedAt:   r.RequestedAt,
		UpdatedAt:     r.UpdatedAt,
	}
	if r.User != nil {
		resp.UserEmail = r.User.Email
	}
	if r.Workspace != nil {
		resp.WorkspaceName = r.Workspace.WorkspaceName
	}
	return resp
}
