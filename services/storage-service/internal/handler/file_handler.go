package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"storage-service/internal/domain"
	"storage-service/internal/service"
)

// FileHandler handles file HTTP requests
type FileHandler struct {
	fileService *service.FileService
}

// NewFileHandler creates a new FileHandler
func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

// GenerateUploadURL godoc
// @Summary Generate presigned upload URL
// @Description Generates a presigned URL for direct file upload to S3
// @Tags files
// @Accept json
// @Produce json
// @Param request body domain.GenerateUploadURLRequest true "Upload request"
// @Success 200 {object} domain.GenerateUploadURLResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/files/upload-url [post]
func (h *FileHandler) GenerateUploadURL(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		handleUnauthorized(c, "User not authenticated")
		return
	}

	var req domain.GenerateUploadURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	response, err := h.fileService.GenerateUploadURL(c.Request.Context(), req, userID)
	if err != nil {
		handleBadRequest(c, err.Error())
		return
	}

	respondWithData(c, http.StatusOK, response)
}

// ConfirmUpload godoc
// @Summary Confirm file upload
// @Description Confirms that file upload is complete and activates the file
// @Tags files
// @Accept json
// @Produce json
// @Param request body domain.ConfirmUploadRequest true "Confirm request"
// @Success 200 {object} domain.FileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/files/confirm [post]
func (h *FileHandler) ConfirmUpload(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		handleUnauthorized(c, "User not authenticated")
		return
	}

	var req domain.ConfirmUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	file, err := h.fileService.ConfirmUpload(c.Request.Context(), req.FileID, userID)
	if err != nil {
		handleBadRequest(c, err.Error())
		return
	}

	respondWithData(c, http.StatusOK, file.ToResponse(h.fileService.GetFileURL(file.FileKey)))
}

// GetFile godoc
// @Summary Get file by ID
// @Description Gets file details by ID
// @Tags files
// @Produce json
// @Param fileId path string true "File ID"
// @Success 200 {object} domain.FileResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/files/{fileId} [get]
func (h *FileHandler) GetFile(c *gin.Context) {
	fileIDStr := c.Param("fileId")
	fileID, err := parseUUID(fileIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid file ID")
		return
	}

	file, err := h.fileService.GetFile(c.Request.Context(), fileID)
	if err != nil {
		handleNotFound(c, err.Error())
		return
	}

	respondWithData(c, http.StatusOK, file.ToResponse(h.fileService.GetFileURL(file.FileKey)))
}

// GetWorkspaceFiles godoc
// @Summary Get files in workspace
// @Description Gets all files in workspace with pagination
// @Tags files
// @Produce json
// @Param workspaceId path string true "Workspace ID"
// @Param page query int false "Page number (default 1)"
// @Param pageSize query int false "Page size (default 20, max 100)"
// @Success 200 {object} domain.FileListResponse
// @Failure 400 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/workspaces/{workspaceId}/files [get]
func (h *FileHandler) GetWorkspaceFiles(c *gin.Context) {
	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := parseUUID(workspaceIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid workspace ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	response, err := h.fileService.GetWorkspaceFiles(c.Request.Context(), workspaceID, page, pageSize)
	if err != nil {
		handleInternalError(c, err.Error())
		return
	}

	respondWithData(c, http.StatusOK, response)
}

// UpdateFile godoc
// @Summary Update a file
// @Description Updates file name or location
// @Tags files
// @Accept json
// @Produce json
// @Param fileId path string true "File ID"
// @Param request body domain.UpdateFileRequest true "Update file request"
// @Success 200 {object} domain.FileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/files/{fileId} [put]
func (h *FileHandler) UpdateFile(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		handleUnauthorized(c, "User not authenticated")
		return
	}

	fileIDStr := c.Param("fileId")
	fileID, err := parseUUID(fileIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid file ID")
		return
	}

	var req domain.UpdateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	file, err := h.fileService.UpdateFile(c.Request.Context(), fileID, req, userID)
	if err != nil {
		handleBadRequest(c, err.Error())
		return
	}

	respondWithData(c, http.StatusOK, file.ToResponse(h.fileService.GetFileURL(file.FileKey)))
}

// DeleteFile godoc
// @Summary Delete a file
// @Description Moves file to trash (soft delete)
// @Tags files
// @Produce json
// @Param fileId path string true "File ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/files/{fileId} [delete]
func (h *FileHandler) DeleteFile(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		handleUnauthorized(c, "User not authenticated")
		return
	}

	fileIDStr := c.Param("fileId")
	fileID, err := parseUUID(fileIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid file ID")
		return
	}

	if err := h.fileService.DeleteFile(c.Request.Context(), fileID, userID); err != nil {
		handleBadRequest(c, err.Error())
		return
	}

	respondWithSuccess(c, http.StatusOK, "File moved to trash", nil)
}

// RestoreFile godoc
// @Summary Restore a file from trash
// @Description Restores a previously deleted file
// @Tags files
// @Produce json
// @Param fileId path string true "File ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/files/{fileId}/restore [post]
func (h *FileHandler) RestoreFile(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		handleUnauthorized(c, "User not authenticated")
		return
	}

	fileIDStr := c.Param("fileId")
	fileID, err := parseUUID(fileIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid file ID")
		return
	}

	if err := h.fileService.RestoreFile(c.Request.Context(), fileID, userID); err != nil {
		handleBadRequest(c, err.Error())
		return
	}

	respondWithSuccess(c, http.StatusOK, "File restored", nil)
}

// PermanentDeleteFile godoc
// @Summary Permanently delete a file
// @Description Permanently deletes a file from storage (cannot be undone)
// @Tags files
// @Produce json
// @Param fileId path string true "File ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/files/{fileId}/permanent [delete]
func (h *FileHandler) PermanentDeleteFile(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		handleUnauthorized(c, "User not authenticated")
		return
	}

	fileIDStr := c.Param("fileId")
	fileID, err := parseUUID(fileIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid file ID")
		return
	}

	if err := h.fileService.PermanentDeleteFile(c.Request.Context(), fileID, userID); err != nil {
		handleBadRequest(c, err.Error())
		return
	}

	respondWithSuccess(c, http.StatusOK, "File permanently deleted", nil)
}

// GetTrashFiles godoc
// @Summary Get files in trash
// @Description Gets all deleted files in workspace
// @Tags files
// @Produce json
// @Param workspaceId path string true "Workspace ID"
// @Success 200 {array} domain.FileResponse
// @Failure 400 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/workspaces/{workspaceId}/trash/files [get]
func (h *FileHandler) GetTrashFiles(c *gin.Context) {
	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := parseUUID(workspaceIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid workspace ID")
		return
	}

	files, err := h.fileService.GetTrashFiles(c.Request.Context(), workspaceID)
	if err != nil {
		handleInternalError(c, err.Error())
		return
	}

	var responses []domain.FileResponse
	for _, file := range files {
		responses = append(responses, file.ToResponse(h.fileService.GetFileURL(file.FileKey)))
	}

	respondWithData(c, http.StatusOK, responses)
}

// SearchFiles godoc
// @Summary Search files
// @Description Searches files by name in workspace
// @Tags files
// @Produce json
// @Param workspaceId path string true "Workspace ID"
// @Param q query string true "Search query"
// @Param page query int false "Page number (default 1)"
// @Param pageSize query int false "Page size (default 20, max 100)"
// @Success 200 {object} domain.FileListResponse
// @Failure 400 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/workspaces/{workspaceId}/files/search [get]
func (h *FileHandler) SearchFiles(c *gin.Context) {
	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := parseUUID(workspaceIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid workspace ID")
		return
	}

	query := c.Query("q")
	if query == "" {
		handleBadRequest(c, "Search query is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	response, err := h.fileService.SearchFiles(c.Request.Context(), workspaceID, query, page, pageSize)
	if err != nil {
		handleInternalError(c, err.Error())
		return
	}

	respondWithData(c, http.StatusOK, response)
}

// GetStorageUsage godoc
// @Summary Get storage usage
// @Description Gets storage usage statistics for workspace
// @Tags files
// @Produce json
// @Param workspaceId path string true "Workspace ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/workspaces/{workspaceId}/usage [get]
func (h *FileHandler) GetStorageUsage(c *gin.Context) {
	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := parseUUID(workspaceIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid workspace ID")
		return
	}

	totalSize, fileCount, err := h.fileService.GetStorageUsage(c.Request.Context(), workspaceID)
	if err != nil {
		handleInternalError(c, err.Error())
		return
	}

	respondWithData(c, http.StatusOK, gin.H{
		"totalSize":     totalSize,
		"totalSizeMB":   float64(totalSize) / (1024 * 1024),
		"totalSizeGB":   float64(totalSize) / (1024 * 1024 * 1024),
		"fileCount":     fileCount,
		"workspaceId":   workspaceID,
	})
}

// GetDownloadURL godoc
// @Summary Get download URL
// @Description Generates a presigned URL for file download
// @Tags files
// @Produce json
// @Param fileId path string true "File ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/files/{fileId}/download [get]
func (h *FileHandler) GetDownloadURL(c *gin.Context) {
	fileIDStr := c.Param("fileId")
	fileID, err := parseUUID(fileIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid file ID")
		return
	}

	url, err := h.fileService.GenerateDownloadURL(c.Request.Context(), fileID)
	if err != nil {
		handleNotFound(c, err.Error())
		return
	}

	respondWithData(c, http.StatusOK, gin.H{
		"downloadUrl": url,
	})
}
