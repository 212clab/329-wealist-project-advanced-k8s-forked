package domain

import (
	"time"

	"github.com/google/uuid"
)

// Folder represents a folder in the storage system (like Google Drive folder)
type Folder struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID uuid.UUID  `gorm:"type:uuid;not null;index" json:"workspaceId"`
	ParentID    *uuid.UUID `gorm:"type:uuid;index" json:"parentId,omitempty"` // nil means root folder
	Name        string     `gorm:"size:255;not null" json:"name"`
	Path        string     `gorm:"size:2048;not null;index" json:"path"` // Full path like /documents/projects
	Color       *string    `gorm:"size:7" json:"color,omitempty"`        // Hex color code like #FF5733
	CreatedBy   uuid.UUID  `gorm:"type:uuid;not null" json:"createdBy"`
	CreatedAt   time.Time  `gorm:"not null" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"not null" json:"updatedAt"`
	DeletedAt   *time.Time `gorm:"index" json:"deletedAt,omitempty"` // Soft delete for trash

	// Relations (not stored in DB, populated via JOIN)
	Parent   *Folder  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Folder `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Files    []File   `gorm:"foreignKey:FolderID" json:"files,omitempty"`
}

// TableName returns the table name for Folder
func (Folder) TableName() string {
	return "storage_folders"
}

// IsRoot returns true if this folder is a root folder
func (f *Folder) IsRoot() bool {
	return f.ParentID == nil
}

// IsDeleted returns true if this folder is in trash
func (f *Folder) IsDeleted() bool {
	return f.DeletedAt != nil
}

// CreateFolderRequest represents request for creating a new folder
type CreateFolderRequest struct {
	WorkspaceID uuid.UUID  `json:"workspaceId" binding:"required"`
	ParentID    *uuid.UUID `json:"parentId,omitempty"`
	Name        string     `json:"name" binding:"required,min=1,max=255"`
	Color       *string    `json:"color,omitempty"`
}

// UpdateFolderRequest represents request for updating a folder
type UpdateFolderRequest struct {
	Name     *string    `json:"name,omitempty"`
	Color    *string    `json:"color,omitempty"`
	ParentID *uuid.UUID `json:"parentId,omitempty"` // For moving folder
}

// FolderResponse represents folder data returned to client
type FolderResponse struct {
	ID          uuid.UUID         `json:"id"`
	WorkspaceID uuid.UUID         `json:"workspaceId"`
	ParentID    *uuid.UUID        `json:"parentId,omitempty"`
	Name        string            `json:"name"`
	Path        string            `json:"path"`
	Color       *string           `json:"color,omitempty"`
	CreatedBy   uuid.UUID         `json:"createdBy"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	IsDeleted   bool              `json:"isDeleted"`
	Children    []FolderResponse  `json:"children,omitempty"`
	Files       []FileResponse    `json:"files,omitempty"`
	FileCount   int64             `json:"fileCount"`
	FolderCount int64             `json:"folderCount"`
	TotalSize   int64             `json:"totalSize"` // Total size in bytes
}

// ToResponse converts Folder to FolderResponse
func (f *Folder) ToResponse() FolderResponse {
	return FolderResponse{
		ID:          f.ID,
		WorkspaceID: f.WorkspaceID,
		ParentID:    f.ParentID,
		Name:        f.Name,
		Path:        f.Path,
		Color:       f.Color,
		CreatedBy:   f.CreatedBy,
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
		IsDeleted:   f.IsDeleted(),
	}
}
