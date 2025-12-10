package domain

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// FileStatus represents the status of a file
type FileStatus string

const (
	FileStatusUploading FileStatus = "UPLOADING" // File is being uploaded
	FileStatusActive    FileStatus = "ACTIVE"    // File is active and accessible
	FileStatusDeleted   FileStatus = "DELETED"   // File is in trash
)

// File represents a file in the storage system
type File struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID uuid.UUID  `gorm:"type:uuid;not null;index" json:"workspaceId"`
	FolderID    *uuid.UUID `gorm:"type:uuid;index" json:"folderId,omitempty"` // nil means root folder
	Name        string     `gorm:"size:255;not null" json:"name"`
	OriginalName string    `gorm:"size:255;not null" json:"originalName"`
	FileKey     string     `gorm:"size:512;not null;uniqueIndex" json:"fileKey"` // S3 key
	FileSize    int64      `gorm:"not null" json:"fileSize"`                      // Size in bytes
	ContentType string     `gorm:"size:128;not null" json:"contentType"`
	Status      FileStatus `gorm:"size:20;not null;default:'ACTIVE'" json:"status"`
	Version     int        `gorm:"not null;default:1" json:"version"` // File versioning
	UploadedBy  uuid.UUID  `gorm:"type:uuid;not null;index" json:"uploadedBy"`
	CreatedAt   time.Time  `gorm:"not null" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"not null" json:"updatedAt"`
	DeletedAt   *time.Time `gorm:"index" json:"deletedAt,omitempty"` // Soft delete for trash

	// Relations
	Folder *Folder     `gorm:"foreignKey:FolderID" json:"folder,omitempty"`
	Shares []FileShare `gorm:"foreignKey:FileID" json:"shares,omitempty"`
}

// TableName returns the table name for File
func (File) TableName() string {
	return "storage_files"
}

// IsDeleted returns true if this file is in trash
func (f *File) IsDeleted() bool {
	return f.DeletedAt != nil || f.Status == FileStatusDeleted
}

// GetExtension returns the file extension
func (f *File) GetExtension() string {
	return strings.ToLower(filepath.Ext(f.Name))
}

// IsImage returns true if the file is an image
func (f *File) IsImage() bool {
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg", ".ico"}
	ext := f.GetExtension()
	for _, e := range imageExts {
		if ext == e {
			return true
		}
	}
	return false
}

// IsDocument returns true if the file is a document
func (f *File) IsDocument() bool {
	docExts := []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".rtf", ".odt"}
	ext := f.GetExtension()
	for _, e := range docExts {
		if ext == e {
			return true
		}
	}
	return false
}

// CreateFileRequest represents request for creating a new file record
type CreateFileRequest struct {
	WorkspaceID  uuid.UUID  `json:"workspaceId" binding:"required"`
	FolderID     *uuid.UUID `json:"folderId,omitempty"`
	Name         string     `json:"name" binding:"required"`
	OriginalName string     `json:"originalName" binding:"required"`
	FileKey      string     `json:"fileKey" binding:"required"`
	FileSize     int64      `json:"fileSize" binding:"required,min=1"`
	ContentType  string     `json:"contentType" binding:"required"`
}

// UpdateFileRequest represents request for updating a file
type UpdateFileRequest struct {
	Name     *string    `json:"name,omitempty"`
	FolderID *uuid.UUID `json:"folderId,omitempty"` // For moving file
}

// GenerateUploadURLRequest represents request for generating presigned upload URL
type GenerateUploadURLRequest struct {
	WorkspaceID uuid.UUID  `json:"workspaceId" binding:"required"`
	FolderID    *uuid.UUID `json:"folderId,omitempty"`
	FileName    string     `json:"fileName" binding:"required"`
	ContentType string     `json:"contentType" binding:"required"`
	FileSize    int64      `json:"fileSize" binding:"required,min=1"`
}

// GenerateUploadURLResponse represents response with presigned upload URL
type GenerateUploadURLResponse struct {
	UploadURL string    `json:"uploadUrl"`
	FileKey   string    `json:"fileKey"`
	FileID    uuid.UUID `json:"fileId"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// ConfirmUploadRequest represents request to confirm file upload completion
type ConfirmUploadRequest struct {
	FileID uuid.UUID `json:"fileId" binding:"required"`
}

// FileResponse represents file data returned to client
type FileResponse struct {
	ID           uuid.UUID  `json:"id"`
	WorkspaceID  uuid.UUID  `json:"workspaceId"`
	FolderID     *uuid.UUID `json:"folderId,omitempty"`
	Name         string     `json:"name"`
	OriginalName string     `json:"originalName"`
	FileURL      string     `json:"fileUrl"` // Public URL to access the file
	FileSize     int64      `json:"fileSize"`
	ContentType  string     `json:"contentType"`
	Status       FileStatus `json:"status"`
	Version      int        `json:"version"`
	UploadedBy   uuid.UUID  `json:"uploadedBy"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	IsDeleted    bool       `json:"isDeleted"`
	IsImage      bool       `json:"isImage"`
	IsDocument   bool       `json:"isDocument"`
	Extension    string     `json:"extension"`
}

// ToResponse converts File to FileResponse
func (f *File) ToResponse(fileURL string) FileResponse {
	return FileResponse{
		ID:           f.ID,
		WorkspaceID:  f.WorkspaceID,
		FolderID:     f.FolderID,
		Name:         f.Name,
		OriginalName: f.OriginalName,
		FileURL:      fileURL,
		FileSize:     f.FileSize,
		ContentType:  f.ContentType,
		Status:       f.Status,
		Version:      f.Version,
		UploadedBy:   f.UploadedBy,
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
		IsDeleted:    f.IsDeleted(),
		IsImage:      f.IsImage(),
		IsDocument:   f.IsDocument(),
		Extension:    f.GetExtension(),
	}
}

// FileListResponse represents list of files with pagination
type FileListResponse struct {
	Files      []FileResponse `json:"files"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"pageSize"`
	TotalPages int            `json:"totalPages"`
}
