package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"user-service/internal/domain"
	"user-service/internal/middleware"
	"user-service/internal/service"
)

// WorkspaceHandler handles workspace HTTP requests
type WorkspaceHandler struct {
	workspaceService *service.WorkspaceService
}

// NewWorkspaceHandler creates a new WorkspaceHandler
func NewWorkspaceHandler(workspaceService *service.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{workspaceService: workspaceService}
}

// CreateWorkspace godoc
// @Summary Create a new workspace
// @Tags Workspaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CreateWorkspaceRequest true "Create workspace request"
// @Success 201 {object} domain.WorkspaceResponse
// @Failure 400 {object} ErrorResponse
// @Router /workspaces/create [post]
func (h *WorkspaceHandler) CreateWorkspace(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: ErrorDetail{Code: "UNAUTHORIZED", Message: "User not authenticated"},
		})
		return
	}

	var req domain.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	workspace, err := h.workspaceService.CreateWorkspace(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "CREATE_FAILED", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusCreated, workspace.ToResponse())
}

// GetWorkspace godoc
// @Summary Get workspace by ID
// @Tags Workspaces
// @Produce json
// @Security BearerAuth
// @Param workspaceId path string true "Workspace ID"
// @Success 200 {object} domain.WorkspaceResponse
// @Failure 404 {object} ErrorResponse
// @Router /workspaces/{workspaceId} [get]
func (h *WorkspaceHandler) GetWorkspace(c *gin.Context) {
	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "INVALID_ID", Message: "Invalid workspace ID"},
		})
		return
	}

	workspace, err := h.workspaceService.GetWorkspace(workspaceID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: ErrorDetail{Code: "NOT_FOUND", Message: "Workspace not found"},
		})
		return
	}

	c.JSON(http.StatusOK, workspace.ToResponse())
}

// GetAllWorkspaces godoc
// @Summary Get all workspaces for current user
// @Tags Workspaces
// @Produce json
// @Security BearerAuth
// @Success 200 {array} domain.UserWorkspaceResponse
// @Router /workspaces/all [get]
func (h *WorkspaceHandler) GetAllWorkspaces(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: ErrorDetail{Code: "UNAUTHORIZED", Message: "User not authenticated"},
		})
		return
	}

	members, err := h.workspaceService.GetUserWorkspaces(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{Code: "FETCH_FAILED", Message: err.Error()},
		})
		return
	}

	// Convert to UserWorkspaceResponse with workspace details
	responses := make([]domain.UserWorkspaceResponse, 0, len(members))
	for _, m := range members {
		if m.Workspace == nil {
			continue // Skip if workspace not loaded
		}
		description := ""
		if m.Workspace.WorkspaceDescription != nil {
			description = *m.Workspace.WorkspaceDescription
		}
		responses = append(responses, domain.UserWorkspaceResponse{
			WorkspaceID:          m.WorkspaceID,
			WorkspaceName:        m.Workspace.WorkspaceName,
			WorkspaceDescription: description,
			Owner:                m.RoleName == domain.RoleOwner,
			Role:                 string(m.RoleName),
			CreatedAt:            m.Workspace.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, responses)
}

// SearchPublicWorkspaces godoc
// @Summary Search public workspaces by name
// @Tags Workspaces
// @Produce json
// @Param workspaceName path string true "Workspace name"
// @Success 200 {array} domain.WorkspaceResponse
// @Router /workspaces/public/{workspaceName} [get]
func (h *WorkspaceHandler) SearchPublicWorkspaces(c *gin.Context) {
	name := c.Param("workspaceName")

	workspaces, err := h.workspaceService.SearchPublicWorkspaces(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{Code: "SEARCH_FAILED", Message: err.Error()},
		})
		return
	}

	responses := make([]domain.WorkspaceResponse, len(workspaces))
	for i, w := range workspaces {
		responses[i] = w.ToResponse()
	}

	c.JSON(http.StatusOK, responses)
}

// UpdateWorkspace godoc
// @Summary Update workspace
// @Tags Workspaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspaceId path string true "Workspace ID"
// @Param request body domain.UpdateWorkspaceRequest true "Update workspace request"
// @Success 200 {object} domain.WorkspaceResponse
// @Failure 403 {object} ErrorResponse
// @Router /workspaces/ids/{workspaceId} [put]
func (h *WorkspaceHandler) UpdateWorkspace(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: ErrorDetail{Code: "UNAUTHORIZED", Message: "User not authenticated"},
		})
		return
	}

	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "INVALID_ID", Message: "Invalid workspace ID"},
		})
		return
	}

	var req domain.UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	workspace, err := h.workspaceService.UpdateWorkspace(workspaceID, userID, req)
	if err != nil {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error: ErrorDetail{Code: "UPDATE_FAILED", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, workspace.ToResponse())
}

// DeleteWorkspace godoc
// @Summary Delete workspace (soft delete)
// @Tags Workspaces
// @Produce json
// @Security BearerAuth
// @Param workspaceId path string true "Workspace ID"
// @Success 200 {object} SuccessResponse
// @Failure 403 {object} ErrorResponse
// @Router /workspaces/{workspaceId} [delete]
func (h *WorkspaceHandler) DeleteWorkspace(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: ErrorDetail{Code: "UNAUTHORIZED", Message: "User not authenticated"},
		})
		return
	}

	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "INVALID_ID", Message: "Invalid workspace ID"},
		})
		return
	}

	if err := h.workspaceService.DeleteWorkspace(workspaceID, userID); err != nil {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error: ErrorDetail{Code: "DELETE_FAILED", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "Workspace deleted successfully"})
}

// SetDefaultWorkspace godoc
// @Summary Set default workspace for user
// @Tags Workspaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]string true "Workspace ID"
// @Success 200 {object} SuccessResponse
// @Router /workspaces/default [post]
func (h *WorkspaceHandler) SetDefaultWorkspace(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: ErrorDetail{Code: "UNAUTHORIZED", Message: "User not authenticated"},
		})
		return
	}

	var req struct {
		WorkspaceID string `json:"workspaceId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "INVALID_ID", Message: "Invalid workspace ID"},
		})
		return
	}

	if err := h.workspaceService.SetDefaultWorkspace(userID, workspaceID); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "SET_DEFAULT_FAILED", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "Default workspace set successfully"})
}

// GetMembers godoc
// @Summary Get workspace members
// @Tags Workspace Members
// @Produce json
// @Security BearerAuth
// @Param workspaceId path string true "Workspace ID"
// @Success 200 {array} domain.WorkspaceMemberResponse
// @Router /workspaces/{workspaceId}/members [get]
func (h *WorkspaceHandler) GetMembers(c *gin.Context) {
	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "INVALID_ID", Message: "Invalid workspace ID"},
		})
		return
	}

	// Use GetMembersWithProfiles to include nickName and profileImageUrl
	responses, err := h.workspaceService.GetMembersWithProfiles(workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{Code: "FETCH_FAILED", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, responses)
}

// InviteMember godoc
// @Summary Invite user to workspace
// @Tags Workspace Members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspaceId path string true "Workspace ID"
// @Param request body domain.InviteMemberRequest true "Invite member request"
// @Success 201 {object} domain.WorkspaceMemberResponse
// @Failure 403 {object} ErrorResponse
// @Router /workspaces/{workspaceId}/members/invite [post]
func (h *WorkspaceHandler) InviteMember(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: ErrorDetail{Code: "UNAUTHORIZED", Message: "User not authenticated"},
		})
		return
	}

	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "INVALID_ID", Message: "Invalid workspace ID"},
		})
		return
	}

	var req domain.InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	member, err := h.workspaceService.InviteMember(workspaceID, userID, req)
	if err != nil {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error: ErrorDetail{Code: "INVITE_FAILED", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusCreated, member.ToResponse())
}

// UpdateMemberRole godoc
// @Summary Update member role
// @Tags Workspace Members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspaceId path string true "Workspace ID"
// @Param memberId path string true "Member ID"
// @Param request body domain.UpdateMemberRoleRequest true "Update role request"
// @Success 200 {object} domain.WorkspaceMemberResponse
// @Failure 403 {object} ErrorResponse
// @Router /workspaces/{workspaceId}/members/{memberId}/role [put]
func (h *WorkspaceHandler) UpdateMemberRole(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: ErrorDetail{Code: "UNAUTHORIZED", Message: "User not authenticated"},
		})
		return
	}

	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "INVALID_ID", Message: "Invalid workspace ID"},
		})
		return
	}

	memberIDStr := c.Param("memberId")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "INVALID_ID", Message: "Invalid member ID"},
		})
		return
	}

	var req domain.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	member, err := h.workspaceService.UpdateMemberRole(workspaceID, memberID, userID, req)
	if err != nil {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error: ErrorDetail{Code: "UPDATE_FAILED", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, member.ToResponse())
}

// RemoveMember godoc
// @Summary Remove member from workspace
// @Tags Workspace Members
// @Produce json
// @Security BearerAuth
// @Param workspaceId path string true "Workspace ID"
// @Param memberId path string true "Member ID"
// @Success 200 {object} SuccessResponse
// @Failure 403 {object} ErrorResponse
// @Router /workspaces/{workspaceId}/members/{memberId} [delete]
func (h *WorkspaceHandler) RemoveMember(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: ErrorDetail{Code: "UNAUTHORIZED", Message: "User not authenticated"},
		})
		return
	}

	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "INVALID_ID", Message: "Invalid workspace ID"},
		})
		return
	}

	memberIDStr := c.Param("memberId")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "INVALID_ID", Message: "Invalid member ID"},
		})
		return
	}

	if err := h.workspaceService.RemoveMember(workspaceID, memberID, userID); err != nil {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error: ErrorDetail{Code: "REMOVE_FAILED", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "Member removed successfully"})
}

// ValidateMember godoc
// @Summary Validate user has access to workspace
// @Tags Workspace Members
// @Produce json
// @Security BearerAuth
// @Param workspaceId path string true "Workspace ID"
// @Param userId path string true "User ID"
// @Success 200 {object} map[string]bool
// @Router /workspaces/{workspaceId}/validate-member/{userId} [get]
func (h *WorkspaceHandler) ValidateMember(c *gin.Context) {
	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "INVALID_ID", Message: "Invalid workspace ID"},
		})
		return
	}

	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "INVALID_ID", Message: "Invalid user ID"},
		})
		return
	}

	isMember, err := h.workspaceService.ValidateMemberAccess(workspaceID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{Code: "VALIDATION_FAILED", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"isMember": isMember})
}

// CreateJoinRequest godoc
// @Summary Create join request for workspace
// @Tags Join Requests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CreateJoinRequestRequest true "Join request"
// @Success 201 {object} domain.JoinRequestResponse
// @Router /workspaces/join-requests [post]
func (h *WorkspaceHandler) CreateJoinRequest(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: ErrorDetail{Code: "UNAUTHORIZED", Message: "User not authenticated"},
		})
		return
	}

	var req domain.CreateJoinRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	request, err := h.workspaceService.CreateJoinRequest(req.WorkspaceID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "CREATE_FAILED", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusCreated, request.ToResponse())
}

// GetJoinRequests godoc
// @Summary Get join requests for workspace
// @Tags Join Requests
// @Produce json
// @Security BearerAuth
// @Param workspaceId path string true "Workspace ID"
// @Success 200 {array} domain.JoinRequestResponse
// @Router /workspaces/{workspaceId}/joinRequests [get]
func (h *WorkspaceHandler) GetJoinRequests(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: ErrorDetail{Code: "UNAUTHORIZED", Message: "User not authenticated"},
		})
		return
	}

	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "INVALID_ID", Message: "Invalid workspace ID"},
		})
		return
	}

	requests, err := h.workspaceService.GetJoinRequests(workspaceID, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error: ErrorDetail{Code: "FETCH_FAILED", Message: err.Error()},
		})
		return
	}

	responses := make([]domain.JoinRequestResponse, len(requests))
	for i, r := range requests {
		responses[i] = r.ToResponse()
	}

	c.JSON(http.StatusOK, responses)
}

// ProcessJoinRequest godoc
// @Summary Process join request (approve/reject)
// @Tags Join Requests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspaceId path string true "Workspace ID"
// @Param requestId path string true "Request ID"
// @Param request body domain.ProcessJoinRequestRequest true "Process request"
// @Success 200 {object} domain.JoinRequestResponse
// @Router /workspaces/{workspaceId}/joinRequests/{requestId} [put]
func (h *WorkspaceHandler) ProcessJoinRequest(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: ErrorDetail{Code: "UNAUTHORIZED", Message: "User not authenticated"},
		})
		return
	}

	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "INVALID_ID", Message: "Invalid workspace ID"},
		})
		return
	}

	requestIDStr := c.Param("requestId")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "INVALID_ID", Message: "Invalid request ID"},
		})
		return
	}

	var req domain.ProcessJoinRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	request, err := h.workspaceService.ProcessJoinRequest(workspaceID, requestID, userID, req)
	if err != nil {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error: ErrorDetail{Code: "PROCESS_FAILED", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, request.ToResponse())
}

// GetWorkspaceSettings godoc
// @Summary Get workspace settings
// @Tags Workspaces
// @Produce json
// @Security BearerAuth
// @Param workspaceId path string true "Workspace ID"
// @Success 200 {object} domain.WorkspaceSettingsResponse
// @Router /workspaces/{workspaceId}/settings [get]
func (h *WorkspaceHandler) GetWorkspaceSettings(c *gin.Context) {
	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "INVALID_ID", Message: "Invalid workspace ID"},
		})
		return
	}

	workspace, err := h.workspaceService.GetWorkspace(workspaceID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: ErrorDetail{Code: "NOT_FOUND", Message: "Workspace not found"},
		})
		return
	}

	c.JSON(http.StatusOK, workspace.ToSettingsResponse())
}

// UpdateWorkspaceSettings godoc
// @Summary Update workspace settings
// @Tags Workspaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspaceId path string true "Workspace ID"
// @Param request body domain.UpdateWorkspaceSettingsRequest true "Update settings request"
// @Success 200 {object} domain.WorkspaceSettingsResponse
// @Router /workspaces/{workspaceId}/settings [put]
func (h *WorkspaceHandler) UpdateWorkspaceSettings(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: ErrorDetail{Code: "UNAUTHORIZED", Message: "User not authenticated"},
		})
		return
	}

	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "INVALID_ID", Message: "Invalid workspace ID"},
		})
		return
	}

	var req domain.UpdateWorkspaceSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	workspace, err := h.workspaceService.UpdateWorkspaceSettings(workspaceID, userID, req)
	if err != nil {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error: ErrorDetail{Code: "UPDATE_FAILED", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, workspace.ToSettingsResponse())
}
