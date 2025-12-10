package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"storage-service/internal/domain"
	"storage-service/internal/service"
)

// FolderHandler handles folder HTTP requests
type FolderHandler struct {
	folderService *service.FolderService
	fileService   *service.FileService
}

// NewFolderHandler creates a new FolderHandler
func NewFolderHandler(folderService *service.FolderService, fileService *service.FileService) *FolderHandler {
	return &FolderHandler{
		folderService: folderService,
		fileService:   fileService,
	}
}

// CreateFolder godoc
// @Summary Create a new folder
// @Description Creates a new folder in the workspace
// @Tags folders
// @Accept json
// @Produce json
// @Param request body domain.CreateFolderRequest true "Create folder request"
// @Success 201 {object} domain.FolderResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/folders [post]
func (h *FolderHandler) CreateFolder(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		handleUnauthorized(c, "User not authenticated")
		return
	}

	var req domain.CreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	folder, err := h.folderService.CreateFolder(c.Request.Context(), req, userID)
	if err != nil {
		handleBadRequest(c, err.Error())
		return
	}

	respondWithData(c, http.StatusCreated, folder.ToResponse())
}

// GetFolder godoc
// @Summary Get folder by ID
// @Description Gets folder details by ID
// @Tags folders
// @Produce json
// @Param folderId path string true "Folder ID"
// @Success 200 {object} domain.FolderResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/folders/{folderId} [get]
func (h *FolderHandler) GetFolder(c *gin.Context) {
	folderIDStr := c.Param("folderId")
	folderID, err := parseUUID(folderIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid folder ID")
		return
	}

	folder, err := h.folderService.GetFolder(c.Request.Context(), folderID)
	if err != nil {
		handleNotFound(c, err.Error())
		return
	}

	respondWithData(c, http.StatusOK, folder.ToResponse())
}

// GetFolderContents godoc
// @Summary Get folder contents
// @Description Gets folder with its children and files
// @Tags folders
// @Produce json
// @Param workspaceId query string true "Workspace ID"
// @Param folderId query string false "Folder ID (omit for root)"
// @Success 200 {object} domain.FolderResponse
// @Failure 400 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/folders/contents [get]
func (h *FolderHandler) GetFolderContents(c *gin.Context) {
	workspaceIDStr := c.Query("workspaceId")
	if workspaceIDStr == "" {
		handleBadRequest(c, "workspaceId is required")
		return
	}

	workspaceID, err := parseUUID(workspaceIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid workspace ID")
		return
	}

	var folderID *uuid.UUID
	folderIDStr := c.Query("folderId")
	if folderIDStr != "" {
		id, err := parseUUID(folderIDStr)
		if err != nil {
			handleBadRequest(c, "Invalid folder ID")
			return
		}
		folderID = &id
	}

	response, err := h.folderService.GetFolderContents(c.Request.Context(), workspaceID, folderID)
	if err != nil {
		handleNotFound(c, err.Error())
		return
	}

	// Fill in file URLs
	for i := range response.Files {
		response.Files[i].FileURL = h.fileService.GetFileURL(response.Files[i].FileURL)
	}

	respondWithData(c, http.StatusOK, response)
}

// GetWorkspaceFolders godoc
// @Summary Get all folders in workspace
// @Description Gets all folders in a workspace as a tree structure
// @Tags folders
// @Produce json
// @Param workspaceId path string true "Workspace ID"
// @Success 200 {array} domain.FolderResponse
// @Failure 400 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/workspaces/{workspaceId}/folders [get]
func (h *FolderHandler) GetWorkspaceFolders(c *gin.Context) {
	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := parseUUID(workspaceIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid workspace ID")
		return
	}

	folders, err := h.folderService.GetWorkspaceFolders(c.Request.Context(), workspaceID)
	if err != nil {
		handleInternalError(c, err.Error())
		return
	}

	var responses []domain.FolderResponse
	for _, folder := range folders {
		responses = append(responses, folder.ToResponse())
	}

	respondWithData(c, http.StatusOK, responses)
}

// UpdateFolder godoc
// @Summary Update a folder
// @Description Updates folder name, color, or location
// @Tags folders
// @Accept json
// @Produce json
// @Param folderId path string true "Folder ID"
// @Param request body domain.UpdateFolderRequest true "Update folder request"
// @Success 200 {object} domain.FolderResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/folders/{folderId} [put]
func (h *FolderHandler) UpdateFolder(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		handleUnauthorized(c, "User not authenticated")
		return
	}

	folderIDStr := c.Param("folderId")
	folderID, err := parseUUID(folderIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid folder ID")
		return
	}

	var req domain.UpdateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	folder, err := h.folderService.UpdateFolder(c.Request.Context(), folderID, req, userID)
	if err != nil {
		handleBadRequest(c, err.Error())
		return
	}

	respondWithData(c, http.StatusOK, folder.ToResponse())
}

// DeleteFolder godoc
// @Summary Delete a folder
// @Description Moves folder to trash (soft delete)
// @Tags folders
// @Produce json
// @Param folderId path string true "Folder ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/folders/{folderId} [delete]
func (h *FolderHandler) DeleteFolder(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		handleUnauthorized(c, "User not authenticated")
		return
	}

	folderIDStr := c.Param("folderId")
	folderID, err := parseUUID(folderIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid folder ID")
		return
	}

	if err := h.folderService.DeleteFolder(c.Request.Context(), folderID, userID); err != nil {
		handleBadRequest(c, err.Error())
		return
	}

	respondWithSuccess(c, http.StatusOK, "Folder moved to trash", nil)
}

// RestoreFolder godoc
// @Summary Restore a folder from trash
// @Description Restores a previously deleted folder
// @Tags folders
// @Produce json
// @Param folderId path string true "Folder ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/folders/{folderId}/restore [post]
func (h *FolderHandler) RestoreFolder(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		handleUnauthorized(c, "User not authenticated")
		return
	}

	folderIDStr := c.Param("folderId")
	folderID, err := parseUUID(folderIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid folder ID")
		return
	}

	if err := h.folderService.RestoreFolder(c.Request.Context(), folderID, userID); err != nil {
		handleBadRequest(c, err.Error())
		return
	}

	respondWithSuccess(c, http.StatusOK, "Folder restored", nil)
}

// PermanentDeleteFolder godoc
// @Summary Permanently delete a folder
// @Description Permanently deletes a folder (cannot be undone)
// @Tags folders
// @Produce json
// @Param folderId path string true "Folder ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/folders/{folderId}/permanent [delete]
func (h *FolderHandler) PermanentDeleteFolder(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		handleUnauthorized(c, "User not authenticated")
		return
	}

	folderIDStr := c.Param("folderId")
	folderID, err := parseUUID(folderIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid folder ID")
		return
	}

	if err := h.folderService.PermanentDeleteFolder(c.Request.Context(), folderID, userID); err != nil {
		handleBadRequest(c, err.Error())
		return
	}

	respondWithSuccess(c, http.StatusOK, "Folder permanently deleted", nil)
}

// GetTrashFolders godoc
// @Summary Get folders in trash
// @Description Gets all deleted folders in workspace
// @Tags folders
// @Produce json
// @Param workspaceId path string true "Workspace ID"
// @Success 200 {array} domain.FolderResponse
// @Failure 400 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/workspaces/{workspaceId}/trash/folders [get]
func (h *FolderHandler) GetTrashFolders(c *gin.Context) {
	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := parseUUID(workspaceIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid workspace ID")
		return
	}

	folders, err := h.folderService.GetTrashFolders(c.Request.Context(), workspaceID)
	if err != nil {
		handleInternalError(c, err.Error())
		return
	}

	var responses []domain.FolderResponse
	for _, folder := range folders {
		responses = append(responses, folder.ToResponse())
	}

	respondWithData(c, http.StatusOK, responses)
}
