package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"storage-service/internal/domain"
	"storage-service/internal/repository"
)

// FolderService handles folder business logic
type FolderService struct {
	folderRepo *repository.FolderRepository
	fileRepo   *repository.FileRepository
	logger     *zap.Logger
}

// NewFolderService creates a new FolderService
func NewFolderService(
	folderRepo *repository.FolderRepository,
	fileRepo *repository.FileRepository,
	logger *zap.Logger,
) *FolderService {
	return &FolderService{
		folderRepo: folderRepo,
		fileRepo:   fileRepo,
		logger:     logger,
	}
}

// CreateFolder creates a new folder
func (s *FolderService) CreateFolder(ctx context.Context, req domain.CreateFolderRequest, userID uuid.UUID) (*domain.Folder, error) {
	// Validate folder name
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("folder name cannot be empty")
	}

	// Build path
	var path string
	if req.ParentID == nil {
		path = "/" + req.Name
	} else {
		parent, err := s.folderRepo.FindByID(ctx, *req.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("parent folder not found")
			}
			return nil, fmt.Errorf("failed to find parent folder: %w", err)
		}
		if parent.WorkspaceID != req.WorkspaceID {
			return nil, errors.New("parent folder belongs to different workspace")
		}
		path = parent.Path + "/" + req.Name
	}

	// Generate unique name if necessary
	uniqueName, err := s.folderRepo.GenerateUniqueName(ctx, req.WorkspaceID, req.ParentID, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate unique name: %w", err)
	}
	if uniqueName != req.Name {
		// Update path with unique name
		if req.ParentID == nil {
			path = "/" + uniqueName
		} else {
			parent, _ := s.folderRepo.FindByID(ctx, *req.ParentID)
			path = parent.Path + "/" + uniqueName
		}
	}

	folder := &domain.Folder{
		ID:          uuid.New(),
		WorkspaceID: req.WorkspaceID,
		ParentID:    req.ParentID,
		Name:        uniqueName,
		Path:        path,
		Color:       req.Color,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.folderRepo.Create(ctx, folder); err != nil {
		s.logger.Error("Failed to create folder",
			zap.Error(err),
			zap.String("workspaceId", req.WorkspaceID.String()),
			zap.String("name", req.Name),
		)
		return nil, fmt.Errorf("failed to create folder: %w", err)
	}

	s.logger.Info("Folder created",
		zap.String("folderId", folder.ID.String()),
		zap.String("path", folder.Path),
		zap.String("userId", userID.String()),
	)

	return folder, nil
}

// GetFolder gets a folder by ID
func (s *FolderService) GetFolder(ctx context.Context, folderID uuid.UUID) (*domain.Folder, error) {
	folder, err := s.folderRepo.FindByID(ctx, folderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("folder not found")
		}
		return nil, fmt.Errorf("failed to get folder: %w", err)
	}
	return folder, nil
}

// GetFolderContents gets folder with its children and files
func (s *FolderService) GetFolderContents(ctx context.Context, workspaceID uuid.UUID, folderID *uuid.UUID) (*domain.FolderResponse, error) {
	var response domain.FolderResponse

	if folderID == nil {
		// Root folder
		response = domain.FolderResponse{
			ID:          uuid.Nil,
			WorkspaceID: workspaceID,
			Name:        "Root",
			Path:        "/",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	} else {
		folder, err := s.folderRepo.FindByID(ctx, *folderID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("folder not found")
			}
			return nil, fmt.Errorf("failed to get folder: %w", err)
		}
		response = folder.ToResponse()
	}

	// Get child folders
	childFolders, err := s.folderRepo.FindByParentID(ctx, workspaceID, folderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get child folders: %w", err)
	}
	for _, child := range childFolders {
		childResp := child.ToResponse()
		// Get counts for each child
		childResp.FolderCount, _ = s.folderRepo.CountByParentID(ctx, child.ID)
		childResp.FileCount, _ = s.fileRepo.CountByFolderID(ctx, child.ID)
		childResp.TotalSize, _ = s.fileRepo.SumSizeByFolderID(ctx, child.ID)
		response.Children = append(response.Children, childResp)
	}

	// Get files in folder
	files, err := s.fileRepo.FindByFolderID(ctx, workspaceID, folderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get files: %w", err)
	}
	for _, file := range files {
		response.Files = append(response.Files, file.ToResponse(file.FileKey)) // FileKey passed for URL generation in handler
	}

	// Set counts
	response.FolderCount = int64(len(childFolders))
	response.FileCount = int64(len(files))

	return &response, nil
}

// GetWorkspaceFolders gets all folders in a workspace
func (s *FolderService) GetWorkspaceFolders(ctx context.Context, workspaceID uuid.UUID) ([]domain.Folder, error) {
	return s.folderRepo.FindByWorkspaceID(ctx, workspaceID)
}

// GetRootFolders gets all root folders in a workspace
func (s *FolderService) GetRootFolders(ctx context.Context, workspaceID uuid.UUID) ([]domain.Folder, error) {
	return s.folderRepo.FindRootFolders(ctx, workspaceID)
}

// UpdateFolder updates a folder
func (s *FolderService) UpdateFolder(ctx context.Context, folderID uuid.UUID, req domain.UpdateFolderRequest, userID uuid.UUID) (*domain.Folder, error) {
	folder, err := s.folderRepo.FindByID(ctx, folderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("folder not found")
		}
		return nil, fmt.Errorf("failed to get folder: %w", err)
	}

	oldPath := folder.Path

	// Update name if provided
	if req.Name != nil && *req.Name != folder.Name {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, errors.New("folder name cannot be empty")
		}

		// Check for duplicate name
		exists, err := s.folderRepo.ExistsByNameInParent(ctx, folder.WorkspaceID, folder.ParentID, name)
		if err != nil {
			return nil, fmt.Errorf("failed to check duplicate name: %w", err)
		}
		if exists {
			return nil, errors.New("folder with this name already exists")
		}

		folder.Name = name
		// Update path
		if folder.ParentID == nil {
			folder.Path = "/" + name
		} else {
			parent, _ := s.folderRepo.FindByID(ctx, *folder.ParentID)
			folder.Path = parent.Path + "/" + name
		}
	}

	// Update color if provided
	if req.Color != nil {
		folder.Color = req.Color
	}

	// Move folder if parent changed
	if req.ParentID != nil && (folder.ParentID == nil || *req.ParentID != *folder.ParentID) {
		// Validate new parent
		if *req.ParentID != uuid.Nil {
			newParent, err := s.folderRepo.FindByID(ctx, *req.ParentID)
			if err != nil {
				return nil, errors.New("new parent folder not found")
			}
			if newParent.WorkspaceID != folder.WorkspaceID {
				return nil, errors.New("cannot move folder to different workspace")
			}
			// Prevent moving to self or descendant
			if strings.HasPrefix(newParent.Path, folder.Path+"/") || newParent.ID == folder.ID {
				return nil, errors.New("cannot move folder into itself or its descendants")
			}
			folder.ParentID = &newParent.ID
			folder.Path = newParent.Path + "/" + folder.Name
		} else {
			folder.ParentID = nil
			folder.Path = "/" + folder.Name
		}
	}

	folder.UpdatedAt = time.Now()

	if err := s.folderRepo.Update(ctx, folder); err != nil {
		return nil, fmt.Errorf("failed to update folder: %w", err)
	}

	// Update paths of descendants if path changed
	if oldPath != folder.Path {
		if err := s.folderRepo.UpdatePath(ctx, folder.WorkspaceID, oldPath, folder.Path); err != nil {
			s.logger.Error("Failed to update descendant paths", zap.Error(err))
		}
	}

	s.logger.Info("Folder updated",
		zap.String("folderId", folder.ID.String()),
		zap.String("userId", userID.String()),
	)

	return folder, nil
}

// DeleteFolder soft deletes a folder (move to trash)
func (s *FolderService) DeleteFolder(ctx context.Context, folderID uuid.UUID, userID uuid.UUID) error {
	folder, err := s.folderRepo.FindByID(ctx, folderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("folder not found")
		}
		return fmt.Errorf("failed to get folder: %w", err)
	}

	// Soft delete folder and all descendants
	if err := s.folderRepo.SoftDeleteByPath(ctx, folder.WorkspaceID, folder.Path); err != nil {
		return fmt.Errorf("failed to delete folder: %w", err)
	}

	// Soft delete all files in folder and descendants
	if err := s.fileRepo.SoftDeleteByFolderID(ctx, folderID); err != nil {
		s.logger.Error("Failed to delete files in folder", zap.Error(err))
	}

	s.logger.Info("Folder deleted (moved to trash)",
		zap.String("folderId", folder.ID.String()),
		zap.String("userId", userID.String()),
	)

	return nil
}

// RestoreFolder restores a deleted folder
func (s *FolderService) RestoreFolder(ctx context.Context, folderID uuid.UUID, userID uuid.UUID) error {
	folder, err := s.folderRepo.FindByIDWithDeleted(ctx, folderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("folder not found")
		}
		return fmt.Errorf("failed to get folder: %w", err)
	}

	if folder.DeletedAt == nil {
		return errors.New("folder is not deleted")
	}

	if err := s.folderRepo.Restore(ctx, folderID); err != nil {
		return fmt.Errorf("failed to restore folder: %w", err)
	}

	s.logger.Info("Folder restored",
		zap.String("folderId", folder.ID.String()),
		zap.String("userId", userID.String()),
	)

	return nil
}

// PermanentDeleteFolder permanently deletes a folder
func (s *FolderService) PermanentDeleteFolder(ctx context.Context, folderID uuid.UUID, userID uuid.UUID) error {
	folder, err := s.folderRepo.FindByIDWithDeleted(ctx, folderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("folder not found")
		}
		return fmt.Errorf("failed to get folder: %w", err)
	}

	if err := s.folderRepo.PermanentDelete(ctx, folderID); err != nil {
		return fmt.Errorf("failed to permanently delete folder: %w", err)
	}

	s.logger.Info("Folder permanently deleted",
		zap.String("folderId", folder.ID.String()),
		zap.String("userId", userID.String()),
	)

	return nil
}

// GetTrashFolders gets all deleted folders in a workspace
func (s *FolderService) GetTrashFolders(ctx context.Context, workspaceID uuid.UUID) ([]domain.Folder, error) {
	return s.folderRepo.FindDeleted(ctx, workspaceID)
}
