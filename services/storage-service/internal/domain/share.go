package domain

import (
	"time"

	"github.com/google/uuid"
)

// ShareType represents the type of entity being shared
type ShareType string

const (
	ShareTypeFile   ShareType = "FILE"
	ShareTypeFolder ShareType = "FOLDER"
)

// PermissionLevel represents the level of access for a share
type PermissionLevel string

const (
	PermissionViewer    PermissionLevel = "VIEWER"    // Can view only
	PermissionCommenter PermissionLevel = "COMMENTER" // Can view and comment
	PermissionEditor    PermissionLevel = "EDITOR"    // Can view, comment, and edit
)

// FileShare represents a sharing configuration for a file
type FileShare struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FileID          uuid.UUID       `gorm:"type:uuid;not null;index" json:"fileId"`
	SharedWithID    *uuid.UUID      `gorm:"type:uuid;index" json:"sharedWithId,omitempty"` // nil means shared via link
	SharedByID      uuid.UUID       `gorm:"type:uuid;not null" json:"sharedById"`
	Permission      PermissionLevel `gorm:"size:20;not null;default:'VIEWER'" json:"permission"`
	ShareLink       *string         `gorm:"size:128;uniqueIndex" json:"shareLink,omitempty"` // Unique share link token
	LinkExpiresAt   *time.Time      `json:"linkExpiresAt,omitempty"`                          // Expiration for share link
	IsPublic        bool            `gorm:"not null;default:false" json:"isPublic"`          // Anyone with link can access
	CreatedAt       time.Time       `gorm:"not null" json:"createdAt"`
	UpdatedAt       time.Time       `gorm:"not null" json:"updatedAt"`

	// Relations
	File *File `gorm:"foreignKey:FileID" json:"file,omitempty"`
}

// TableName returns the table name for FileShare
func (FileShare) TableName() string {
	return "storage_file_shares"
}

// IsExpired returns true if the share link is expired
func (s *FileShare) IsExpired() bool {
	if s.LinkExpiresAt == nil {
		return false
	}
	return time.Now().After(*s.LinkExpiresAt)
}

// FolderShare represents a sharing configuration for a folder
type FolderShare struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FolderID        uuid.UUID       `gorm:"type:uuid;not null;index" json:"folderId"`
	SharedWithID    *uuid.UUID      `gorm:"type:uuid;index" json:"sharedWithId,omitempty"`
	SharedByID      uuid.UUID       `gorm:"type:uuid;not null" json:"sharedById"`
	Permission      PermissionLevel `gorm:"size:20;not null;default:'VIEWER'" json:"permission"`
	ShareLink       *string         `gorm:"size:128;uniqueIndex" json:"shareLink,omitempty"`
	LinkExpiresAt   *time.Time      `json:"linkExpiresAt,omitempty"`
	IsPublic        bool            `gorm:"not null;default:false" json:"isPublic"`
	IncludeChildren bool            `gorm:"not null;default:true" json:"includeChildren"` // Share includes subfolders
	CreatedAt       time.Time       `gorm:"not null" json:"createdAt"`
	UpdatedAt       time.Time       `gorm:"not null" json:"updatedAt"`

	// Relations
	Folder *Folder `gorm:"foreignKey:FolderID" json:"folder,omitempty"`
}

// TableName returns the table name for FolderShare
func (FolderShare) TableName() string {
	return "storage_folder_shares"
}

// IsExpired returns true if the share link is expired
func (s *FolderShare) IsExpired() bool {
	if s.LinkExpiresAt == nil {
		return false
	}
	return time.Now().After(*s.LinkExpiresAt)
}

// CreateShareRequest represents request for creating a share
type CreateShareRequest struct {
	EntityType      ShareType       `json:"entityType" binding:"required"`      // FILE or FOLDER
	EntityID        uuid.UUID       `json:"entityId" binding:"required"`
	SharedWithID    *uuid.UUID      `json:"sharedWithId,omitempty"`             // User ID to share with
	Permission      PermissionLevel `json:"permission" binding:"required"`
	IsPublic        bool            `json:"isPublic"`                           // Create public link
	ExpiresInDays   *int            `json:"expiresInDays,omitempty"`            // Expiration in days (nil = never)
	IncludeChildren *bool           `json:"includeChildren,omitempty"`          // For folder shares
}

// UpdateShareRequest represents request for updating a share
type UpdateShareRequest struct {
	Permission    *PermissionLevel `json:"permission,omitempty"`
	ExpiresInDays *int             `json:"expiresInDays,omitempty"` // nil to remove expiration
}

// ShareResponse represents share data returned to client
type ShareResponse struct {
	ID              uuid.UUID       `json:"id"`
	EntityType      ShareType       `json:"entityType"`
	EntityID        uuid.UUID       `json:"entityId"`
	EntityName      string          `json:"entityName"`
	SharedWithID    *uuid.UUID      `json:"sharedWithId,omitempty"`
	SharedWithName  *string         `json:"sharedWithName,omitempty"` // Resolved from user service
	SharedByID      uuid.UUID       `json:"sharedById"`
	SharedByName    string          `json:"sharedByName"` // Resolved from user service
	Permission      PermissionLevel `json:"permission"`
	ShareLink       *string         `json:"shareLink,omitempty"`
	ShareURL        *string         `json:"shareUrl,omitempty"` // Full URL for sharing
	LinkExpiresAt   *time.Time      `json:"linkExpiresAt,omitempty"`
	IsPublic        bool            `json:"isPublic"`
	IsExpired       bool            `json:"isExpired"`
	IncludeChildren bool            `json:"includeChildren,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

// SharedItem represents an item shared with the current user
type SharedItem struct {
	EntityType  ShareType       `json:"entityType"`
	EntityID    uuid.UUID       `json:"entityId"`
	EntityName  string          `json:"entityName"`
	Permission  PermissionLevel `json:"permission"`
	SharedByID  uuid.UUID       `json:"sharedById"`
	SharedByName string         `json:"sharedByName"`
	SharedAt    time.Time       `json:"sharedAt"`
}

// SharedWithMeResponse represents list of items shared with the user
type SharedWithMeResponse struct {
	Items      []SharedItem `json:"items"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"pageSize"`
	TotalPages int          `json:"totalPages"`
}
