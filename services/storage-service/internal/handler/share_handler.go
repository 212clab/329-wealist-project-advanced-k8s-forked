package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"storage-service/internal/domain"
	"storage-service/internal/service"
)

// ShareHandler handles share HTTP requests
type ShareHandler struct {
	shareService *service.ShareService
}

// NewShareHandler creates a new ShareHandler
func NewShareHandler(shareService *service.ShareService) *ShareHandler {
	return &ShareHandler{
		shareService: shareService,
	}
}

// CreateShare godoc
// @Summary Create a share
// @Description Creates a new share for a file or folder
// @Tags shares
// @Accept json
// @Produce json
// @Param request body domain.CreateShareRequest true "Create share request"
// @Success 201 {object} domain.ShareResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/shares [post]
func (h *ShareHandler) CreateShare(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		handleUnauthorized(c, "User not authenticated")
		return
	}

	var req domain.CreateShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	response, err := h.shareService.CreateShare(c.Request.Context(), req, userID)
	if err != nil {
		handleBadRequest(c, err.Error())
		return
	}

	respondWithData(c, http.StatusCreated, response)
}

// GetFileShares godoc
// @Summary Get file shares
// @Description Gets all shares for a file
// @Tags shares
// @Produce json
// @Param fileId path string true "File ID"
// @Success 200 {array} domain.ShareResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/files/{fileId}/shares [get]
func (h *ShareHandler) GetFileShares(c *gin.Context) {
	fileIDStr := c.Param("fileId")
	fileID, err := parseUUID(fileIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid file ID")
		return
	}

	shares, err := h.shareService.GetFileShares(c.Request.Context(), fileID)
	if err != nil {
		handleNotFound(c, err.Error())
		return
	}

	respondWithData(c, http.StatusOK, shares)
}

// GetFolderShares godoc
// @Summary Get folder shares
// @Description Gets all shares for a folder
// @Tags shares
// @Produce json
// @Param folderId path string true "Folder ID"
// @Success 200 {array} domain.ShareResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/folders/{folderId}/shares [get]
func (h *ShareHandler) GetFolderShares(c *gin.Context) {
	folderIDStr := c.Param("folderId")
	folderID, err := parseUUID(folderIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid folder ID")
		return
	}

	shares, err := h.shareService.GetFolderShares(c.Request.Context(), folderID)
	if err != nil {
		handleNotFound(c, err.Error())
		return
	}

	respondWithData(c, http.StatusOK, shares)
}

// GetSharedWithMe godoc
// @Summary Get items shared with me
// @Description Gets all files and folders shared with the current user
// @Tags shares
// @Produce json
// @Success 200 {array} domain.SharedItem
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/shared-with-me [get]
func (h *ShareHandler) GetSharedWithMe(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		handleUnauthorized(c, "User not authenticated")
		return
	}

	items, err := h.shareService.GetSharedWithMe(c.Request.Context(), userID)
	if err != nil {
		handleInternalError(c, err.Error())
		return
	}

	respondWithData(c, http.StatusOK, items)
}

// GetShareByLink godoc
// @Summary Get share by link
// @Description Gets share details by public share link
// @Tags shares
// @Produce json
// @Param link path string true "Share link"
// @Success 200 {object} domain.ShareResponse
// @Failure 404 {object} ErrorResponse
// @Router /storage/shares/link/{link} [get]
func (h *ShareHandler) GetShareByLink(c *gin.Context) {
	link := c.Param("link")
	if link == "" {
		handleBadRequest(c, "Share link is required")
		return
	}

	share, err := h.shareService.GetShareByLink(c.Request.Context(), link)
	if err != nil {
		handleNotFound(c, err.Error())
		return
	}

	respondWithData(c, http.StatusOK, share)
}

// UpdateShare godoc
// @Summary Update a share
// @Description Updates share permission or expiration
// @Tags shares
// @Accept json
// @Produce json
// @Param shareId path string true "Share ID"
// @Param type query string true "Entity type (FILE or FOLDER)"
// @Param request body domain.UpdateShareRequest true "Update share request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/shares/{shareId} [put]
func (h *ShareHandler) UpdateShare(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		handleUnauthorized(c, "User not authenticated")
		return
	}

	shareIDStr := c.Param("shareId")
	shareID, err := parseUUID(shareIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid share ID")
		return
	}

	entityTypeStr := c.Query("type")
	var entityType domain.ShareType
	switch entityTypeStr {
	case "FILE":
		entityType = domain.ShareTypeFile
	case "FOLDER":
		entityType = domain.ShareTypeFolder
	default:
		handleBadRequest(c, "Invalid entity type, must be FILE or FOLDER")
		return
	}

	var req domain.UpdateShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	if err := h.shareService.UpdateShare(c.Request.Context(), shareID, entityType, req, userID); err != nil {
		handleBadRequest(c, err.Error())
		return
	}

	respondWithSuccess(c, http.StatusOK, "Share updated", nil)
}

// DeleteShare godoc
// @Summary Delete a share
// @Description Removes a share
// @Tags shares
// @Produce json
// @Param shareId path string true "Share ID"
// @Param type query string true "Entity type (FILE or FOLDER)"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /storage/shares/{shareId} [delete]
func (h *ShareHandler) DeleteShare(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		handleUnauthorized(c, "User not authenticated")
		return
	}

	shareIDStr := c.Param("shareId")
	shareID, err := parseUUID(shareIDStr)
	if err != nil {
		handleBadRequest(c, "Invalid share ID")
		return
	}

	entityTypeStr := c.Query("type")
	var entityType domain.ShareType
	switch entityTypeStr {
	case "FILE":
		entityType = domain.ShareTypeFile
	case "FOLDER":
		entityType = domain.ShareTypeFolder
	default:
		handleBadRequest(c, "Invalid entity type, must be FILE or FOLDER")
		return
	}

	if err := h.shareService.DeleteShare(c.Request.Context(), shareID, entityType, userID); err != nil {
		handleBadRequest(c, err.Error())
		return
	}

	respondWithSuccess(c, http.StatusOK, "Share removed", nil)
}
