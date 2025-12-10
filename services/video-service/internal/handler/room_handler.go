package handler

import (
	"errors"
	"fmt"
	"net/http"
	"video-service/internal/domain"
	"video-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type RoomHandler struct {
	roomService service.RoomService
	logger      *zap.Logger
}

func NewRoomHandler(roomService service.RoomService, logger *zap.Logger) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
		logger:      logger,
	}
}

// CreateRoom godoc
// @Summary Create a new video call room
// @Tags rooms
// @Accept json
// @Produce json
// @Param request body domain.CreateRoomRequest true "Room details"
// @Success 201 {object} domain.RoomResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Security BearerAuth
// @Router /rooms [post]
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	token := c.MustGet("token").(string)

	var req domain.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": err.Error()},
		})
		return
	}

	room, err := h.roomService.CreateRoom(c.Request.Context(), &req, userID, token)
	if err != nil {
		if errors.Is(err, service.ErrNotWorkspaceMember) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   gin.H{"code": "FORBIDDEN", "message": "You are not a member of this workspace"},
			})
			return
		}
		h.logger.Error("Failed to create room", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"code": "INTERNAL_ERROR", "message": "Failed to create room"},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    room,
	})
}

// GetRoom godoc
// @Summary Get room details
// @Tags rooms
// @Produce json
// @Param roomId path string true "Room ID"
// @Success 200 {object} domain.RoomResponse
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /rooms/{roomId} [get]
func (h *RoomHandler) GetRoom(c *gin.Context) {
	roomIDStr := c.Param("roomId")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": "Invalid room ID"},
		})
		return
	}

	room, err := h.roomService.GetRoom(c.Request.Context(), roomID)
	if err != nil {
		if errors.Is(err, service.ErrRoomNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   gin.H{"code": "NOT_FOUND", "message": "Room not found"},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get room"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    room,
	})
}

// GetWorkspaceRooms godoc
// @Summary Get rooms for a workspace
// @Tags rooms
// @Produce json
// @Param workspaceId path string true "Workspace ID"
// @Param active query bool false "Filter active rooms only"
// @Success 200 {array} domain.RoomResponse
// @Security BearerAuth
// @Router /rooms/workspace/{workspaceId} [get]
func (h *RoomHandler) GetWorkspaceRooms(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	token := c.MustGet("token").(string)

	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": "Invalid workspace ID"},
		})
		return
	}

	activeOnly := c.Query("active") == "true"

	rooms, err := h.roomService.GetWorkspaceRooms(c.Request.Context(), workspaceID, userID, token, activeOnly)
	if err != nil {
		if errors.Is(err, service.ErrNotWorkspaceMember) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   gin.H{"code": "FORBIDDEN", "message": "You are not a member of this workspace"},
			})
			return
		}
		h.logger.Error("Failed to get workspace rooms", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get rooms"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rooms,
	})
}

// JoinRoom godoc
// @Summary Join a video call room
// @Tags rooms
// @Accept json
// @Produce json
// @Param roomId path string true "Room ID"
// @Success 200 {object} domain.JoinRoomResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /rooms/{roomId}/join [post]
func (h *RoomHandler) JoinRoom(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	token := c.MustGet("token").(string)

	roomIDStr := c.Param("roomId")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": "Invalid room ID"},
		})
		return
	}

	// Get user name from query or use default
	userName := c.Query("userName")
	if userName == "" {
		userName = userID.String()[:8]
	}

	response, err := h.roomService.JoinRoom(c.Request.Context(), roomID, userID, userName, token)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRoomNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   gin.H{"code": "NOT_FOUND", "message": "Room not found"},
			})
		case errors.Is(err, service.ErrRoomFull):
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   gin.H{"code": "ROOM_FULL", "message": "Room is full"},
			})
		case errors.Is(err, service.ErrAlreadyInRoom):
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   gin.H{"code": "ALREADY_IN_ROOM", "message": "User is already in room"},
			})
		case errors.Is(err, service.ErrRoomNotActive):
			c.JSON(http.StatusGone, gin.H{
				"success": false,
				"error":   gin.H{"code": "ROOM_ENDED", "message": "Room has ended"},
			})
		case errors.Is(err, service.ErrNotWorkspaceMember):
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   gin.H{"code": "FORBIDDEN", "message": "You are not a member of this workspace"},
			})
		default:
			h.logger.Error("Failed to join room", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   gin.H{"code": "INTERNAL_ERROR", "message": "Failed to join room"},
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// LeaveRoom godoc
// @Summary Leave a video call room
// @Tags rooms
// @Param roomId path string true "Room ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /rooms/{roomId}/leave [post]
func (h *RoomHandler) LeaveRoom(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	roomIDStr := c.Param("roomId")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": "Invalid room ID"},
		})
		return
	}

	if err := h.roomService.LeaveRoom(c.Request.Context(), roomID, userID); err != nil {
		if errors.Is(err, service.ErrNotInRoom) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   gin.H{"code": "NOT_IN_ROOM", "message": "User is not in room"},
			})
			return
		}
		h.logger.Error("Failed to leave room", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"code": "INTERNAL_ERROR", "message": "Failed to leave room"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Left room successfully",
	})
}

// EndRoom godoc
// @Summary End a video call room (creator only)
// @Tags rooms
// @Param roomId path string true "Room ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /rooms/{roomId}/end [post]
func (h *RoomHandler) EndRoom(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	roomIDStr := c.Param("roomId")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": "Invalid room ID"},
		})
		return
	}

	if err := h.roomService.EndRoom(c.Request.Context(), roomID, userID); err != nil {
		if errors.Is(err, service.ErrRoomNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   gin.H{"code": "NOT_FOUND", "message": "Room not found"},
			})
			return
		}
		h.logger.Error("Failed to end room", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"code": "INTERNAL_ERROR", "message": "Failed to end room"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Room ended successfully",
	})
}

// GetParticipants godoc
// @Summary Get room participants
// @Tags rooms
// @Produce json
// @Param roomId path string true "Room ID"
// @Success 200 {array} domain.ParticipantResponse
// @Security BearerAuth
// @Router /rooms/{roomId}/participants [get]
func (h *RoomHandler) GetParticipants(c *gin.Context) {
	roomIDStr := c.Param("roomId")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": "Invalid room ID"},
		})
		return
	}

	participants, err := h.roomService.GetParticipants(c.Request.Context(), roomID)
	if err != nil {
		h.logger.Error("Failed to get participants", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get participants"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    participants,
	})
}

// GetWorkspaceCallHistory godoc
// @Summary Get call history for a workspace
// @Tags history
// @Produce json
// @Param workspaceId path string true "Workspace ID"
// @Param limit query int false "Limit (default 20, max 100)"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /history/workspace/{workspaceId} [get]
func (h *RoomHandler) GetWorkspaceCallHistory(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	token := c.MustGet("token").(string)

	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": "Invalid workspace ID"},
		})
		return
	}

	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
			limit = 20
		}
	}
	if o := c.Query("offset"); o != "" {
		if _, err := fmt.Sscanf(o, "%d", &offset); err != nil {
			offset = 0
		}
	}

	histories, total, err := h.roomService.GetWorkspaceCallHistory(c.Request.Context(), workspaceID, userID, token, limit, offset)
	if err != nil {
		if errors.Is(err, service.ErrNotWorkspaceMember) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   gin.H{"code": "FORBIDDEN", "message": "You are not a member of this workspace"},
			})
			return
		}
		h.logger.Error("Failed to get call history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get call history"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    histories,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetMyCallHistory godoc
// @Summary Get current user's call history
// @Tags history
// @Produce json
// @Param limit query int false "Limit (default 20, max 100)"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /history/me [get]
func (h *RoomHandler) GetMyCallHistory(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
			limit = 20
		}
	}
	if o := c.Query("offset"); o != "" {
		if _, err := fmt.Sscanf(o, "%d", &offset); err != nil {
			offset = 0
		}
	}

	histories, total, err := h.roomService.GetUserCallHistory(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get call history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get call history"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    histories,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetCallHistory godoc
// @Summary Get a single call history by ID
// @Tags history
// @Produce json
// @Param historyId path string true "History ID"
// @Success 200 {object} domain.CallHistoryResponse
// @Security BearerAuth
// @Router /history/{historyId} [get]
func (h *RoomHandler) GetCallHistory(c *gin.Context) {
	historyIDStr := c.Param("historyId")
	historyID, err := uuid.Parse(historyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": "Invalid history ID"},
		})
		return
	}

	history, err := h.roomService.GetCallHistoryByID(c.Request.Context(), historyID)
	if err != nil {
		h.logger.Error("Failed to get call history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get call history"},
		})
		return
	}

	if history == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   gin.H{"code": "NOT_FOUND", "message": "Call history not found"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    history,
	})
}

// SaveTranscript godoc
// @Summary Save transcript for a room
// @Tags transcript
// @Accept json
// @Produce json
// @Param roomId path string true "Room ID"
// @Param request body domain.SaveTranscriptRequest true "Transcript content"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /rooms/{roomId}/transcript [post]
func (h *RoomHandler) SaveTranscript(c *gin.Context) {
	roomIDStr := c.Param("roomId")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": "Invalid room ID"},
		})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": err.Error()},
		})
		return
	}

	transcript, err := h.roomService.SaveTranscript(c.Request.Context(), roomID, req.Content)
	if err != nil {
		h.logger.Error("Failed to save transcript", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"code": "INTERNAL_ERROR", "message": "Failed to save transcript"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    transcript,
	})
}

// GetTranscript godoc
// @Summary Get transcript for a call history
// @Tags transcript
// @Produce json
// @Param historyId path string true "Call History ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /history/{historyId}/transcript [get]
func (h *RoomHandler) GetTranscript(c *gin.Context) {
	historyIDStr := c.Param("historyId")
	historyID, err := uuid.Parse(historyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": "Invalid history ID"},
		})
		return
	}

	transcript, err := h.roomService.GetTranscriptByCallHistoryID(c.Request.Context(), historyID)
	if err != nil {
		// Return empty content instead of error if not found
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    transcript,
	})
}
