// Package service implements business logic for video-service.
//
// This package provides the RoomService interface and implementation
// for managing video call rooms, participants, call history, and transcripts.
// It integrates with LiveKit for WebRTC functionality and validates
// workspace membership through the user-service.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"
	"video-service/internal/client"
	"video-service/internal/config"
	"video-service/internal/domain"
	"video-service/internal/repository"

	"github.com/google/uuid"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Service-level errors for room operations.
// These errors are used to communicate specific failure conditions
// to handlers for appropriate HTTP response mapping.
var (
	// ErrRoomNotFound indicates the requested room does not exist.
	ErrRoomNotFound = errors.New("room not found")
	// ErrRoomFull indicates the room has reached its maximum participant capacity.
	ErrRoomFull = errors.New("room is full")
	// ErrAlreadyInRoom indicates the user is already a participant in the room.
	ErrAlreadyInRoom = errors.New("user is already in room")
	// ErrNotInRoom indicates the user is not a participant in the room.
	ErrNotInRoom = errors.New("user is not in room")
	// ErrRoomNotActive indicates the room has ended and is no longer accepting participants.
	ErrRoomNotActive = errors.New("room is not active")
	// ErrNotWorkspaceMember indicates the user is not authorized to access the workspace.
	ErrNotWorkspaceMember = errors.New("user is not a member of this workspace")
)

// RoomService defines the interface for video room business operations.
// It provides methods for room lifecycle management, participant tracking,
// call history retrieval, and transcript storage.
type RoomService interface {
	// CreateRoom creates a new video call room in the specified workspace.
	// It validates workspace membership and initializes a LiveKit room.
	CreateRoom(ctx context.Context, req *domain.CreateRoomRequest, creatorID uuid.UUID, token string) (*domain.RoomResponse, error)

	// GetRoom retrieves a single room by its ID.
	GetRoom(ctx context.Context, roomID uuid.UUID) (*domain.RoomResponse, error)

	// GetWorkspaceRooms retrieves all rooms for a workspace.
	// Set activeOnly to true to filter for only active rooms.
	GetWorkspaceRooms(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, token string, activeOnly bool) ([]domain.RoomResponse, error)

	// JoinRoom adds a user to a room and returns a LiveKit access token.
	// Returns ErrRoomFull if the room has reached capacity.
	JoinRoom(ctx context.Context, roomID, userID uuid.UUID, userName string, token string) (*domain.JoinRoomResponse, error)

	// LeaveRoom removes a user from a room.
	// If the creator leaves, the room is automatically ended.
	LeaveRoom(ctx context.Context, roomID, userID uuid.UUID) error

	// EndRoom closes a room and creates a call history record.
	// Only the room creator can end the room.
	EndRoom(ctx context.Context, roomID, userID uuid.UUID) error

	// GetParticipants returns all active participants in a room.
	GetParticipants(ctx context.Context, roomID uuid.UUID) ([]domain.ParticipantResponse, error)

	// GetWorkspaceCallHistory returns paginated call history for a workspace.
	GetWorkspaceCallHistory(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, token string, limit, offset int) ([]domain.CallHistoryResponse, int64, error)

	// GetUserCallHistory returns paginated call history for a specific user.
	GetUserCallHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.CallHistoryResponse, int64, error)

	// GetCallHistoryByID retrieves a single call history record by ID.
	GetCallHistoryByID(ctx context.Context, historyID uuid.UUID) (*domain.CallHistoryResponse, error)

	// SaveTranscript stores or updates the transcript content for a room.
	SaveTranscript(ctx context.Context, roomID uuid.UUID, content string) (*domain.TranscriptResponse, error)

	// GetTranscriptByCallHistoryID retrieves the transcript for a completed call.
	GetTranscriptByCallHistoryID(ctx context.Context, callHistoryID uuid.UUID) (*domain.TranscriptResponse, error)
}

// roomService implements RoomService interface.
// It coordinates between repository, LiveKit, and user-service client.
type roomService struct {
	roomRepo    repository.RoomRepository
	userClient  client.UserClient
	lkClient    *lksdk.RoomServiceClient
	lkConfig    config.LiveKitConfig
	redisClient *redis.Client
	logger      *zap.Logger
}

// NewRoomService creates a new RoomService with the given dependencies.
// If LiveKit configuration is provided, it initializes the LiveKit client.
// If userClient is nil, workspace validation will be skipped.
func NewRoomService(
	roomRepo repository.RoomRepository,
	userClient client.UserClient,
	lkConfig config.LiveKitConfig,
	redisClient *redis.Client,
	logger *zap.Logger,
) RoomService {
	var lkClient *lksdk.RoomServiceClient
	if lkConfig.Host != "" && lkConfig.APIKey != "" && lkConfig.APISecret != "" {
		lkClient = lksdk.NewRoomServiceClient(lkConfig.Host, lkConfig.APIKey, lkConfig.APISecret)
	}

	return &roomService{
		roomRepo:    roomRepo,
		userClient:  userClient,
		lkClient:    lkClient,
		lkConfig:    lkConfig,
		redisClient: redisClient,
		logger:      logger,
	}
}

func (s *roomService) CreateRoom(ctx context.Context, req *domain.CreateRoomRequest, creatorID uuid.UUID, token string) (*domain.RoomResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("invalid workspace ID: %w", err)
	}

	// Validate workspace membership
	if s.userClient != nil {
		isMember, err := s.userClient.ValidateWorkspaceMember(ctx, workspaceID, creatorID, token)
		if err != nil {
			s.logger.Error("Failed to validate workspace membership", zap.Error(err))
			return nil, fmt.Errorf("failed to validate workspace membership: %w", err)
		}
		if !isMember {
			return nil, ErrNotWorkspaceMember
		}
	}

	maxParticipants := req.MaxParticipants
	if maxParticipants <= 0 {
		maxParticipants = 10
	}

	room := &domain.Room{
		Name:            req.Name,
		WorkspaceID:     workspaceID,
		CreatorID:       creatorID,
		MaxParticipants: maxParticipants,
		IsActive:        true,
	}

	if err := s.roomRepo.Create(room); err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	// Create room in LiveKit
	if s.lkClient != nil {
		_, err := s.lkClient.CreateRoom(ctx, &livekit.CreateRoomRequest{
			Name:            room.ID.String(),
			EmptyTimeout:    300, // 5 minutes
			MaxParticipants: uint32(maxParticipants),
		})
		if err != nil {
			s.logger.Warn("Failed to create LiveKit room", zap.Error(err))
		}
	}

	response := room.ToResponse()
	return &response, nil
}

func (s *roomService) GetRoom(ctx context.Context, roomID uuid.UUID) (*domain.RoomResponse, error) {
	room, err := s.roomRepo.GetByID(roomID)
	if err != nil {
		return nil, ErrRoomNotFound
	}

	response := room.ToResponse()
	return &response, nil
}

func (s *roomService) GetWorkspaceRooms(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, token string, activeOnly bool) ([]domain.RoomResponse, error) {
	// Validate workspace membership
	if s.userClient != nil {
		isMember, err := s.userClient.ValidateWorkspaceMember(ctx, workspaceID, userID, token)
		if err != nil {
			s.logger.Error("Failed to validate workspace membership", zap.Error(err))
			return nil, fmt.Errorf("failed to validate workspace membership: %w", err)
		}
		if !isMember {
			return nil, ErrNotWorkspaceMember
		}
	}

	var rooms []domain.Room
	var err error

	if activeOnly {
		rooms, err = s.roomRepo.GetActiveByWorkspaceID(workspaceID)
	} else {
		rooms, err = s.roomRepo.GetByWorkspaceID(workspaceID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get rooms: %w", err)
	}

	responses := make([]domain.RoomResponse, len(rooms))
	for i, room := range rooms {
		responses[i] = room.ToResponse()
	}

	return responses, nil
}

func (s *roomService) JoinRoom(ctx context.Context, roomID, userID uuid.UUID, userName string, token string) (*domain.JoinRoomResponse, error) {
	room, err := s.roomRepo.GetByID(roomID)
	if err != nil {
		return nil, ErrRoomNotFound
	}

	if !room.IsActive {
		return nil, ErrRoomNotActive
	}

	// Validate workspace membership
	if s.userClient != nil {
		isMember, err := s.userClient.ValidateWorkspaceMember(ctx, room.WorkspaceID, userID, token)
		if err != nil {
			s.logger.Error("Failed to validate workspace membership", zap.Error(err))
			return nil, fmt.Errorf("failed to validate workspace membership: %w", err)
		}
		if !isMember {
			return nil, ErrNotWorkspaceMember
		}
	}

	// Check if user is already in room record (even if inactive)
	existingParticipant, _ := s.roomRepo.GetParticipant(roomID, userID)
	if existingParticipant != nil {
		// User is rejoining room
		s.logger.Info("User rejoining room",
			zap.String("room_id", roomID.String()),
			zap.String("user_id", userID.String()),
		)
		existingParticipant.IsActive = true
		existingParticipant.LeftAt = nil    // Clear leftAt since they are back
		existingParticipant.Name = userName // Update name

		if err := s.roomRepo.UpdateParticipant(existingParticipant); err != nil {
			return nil, fmt.Errorf("failed to update participant: %w", err)
		}

	} else {
		// New participant - check room capacity
		count, err := s.roomRepo.CountActiveParticipants(roomID)
		if err != nil {
			return nil, fmt.Errorf("failed to count participants: %w", err)
		}
		if count >= int64(room.MaxParticipants) {
			return nil, ErrRoomFull
		}

		// Add participant
		participant := &domain.RoomParticipant{
			RoomID:   roomID,
			UserID:   userID,
			IsActive: true,
			Name:     userName,
		}
		if err := s.roomRepo.AddParticipant(participant); err != nil {
			return nil, fmt.Errorf("failed to add participant: %w", err)
		}
	}

	// Generate LiveKit token
	lkToken, err := s.generateLiveKitToken(room.ID.String(), userID.String(), userName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Refresh room data
	room, _ = s.roomRepo.GetByID(roomID)

	return &domain.JoinRoomResponse{
		Room:  room.ToResponse(),
		Token: lkToken,
		WSUrl: s.lkConfig.WSUrl,
	}, nil
}

func (s *roomService) LeaveRoom(ctx context.Context, roomID, userID uuid.UUID) error {
	// First check if room exists and get details (for creator check)
	room, err := s.roomRepo.GetByID(roomID)
	if err != nil {
		return ErrRoomNotFound
	}

	// If the user leaving is the creator, End the room
	if room.CreatorID == userID {
		s.logger.Info("Creator leaving room, ending room", zap.String("room_id", roomID.String()))
		return s.EndRoom(ctx, roomID, userID)
	}

	// Normal check if user is in room
	_, err = s.roomRepo.GetParticipant(roomID, userID)
	if err != nil {
		return ErrNotInRoom
	}

	if err := s.roomRepo.RemoveParticipant(roomID, userID); err != nil {
		return fmt.Errorf("failed to remove participant: %w", err)
	}

	// Check if room is empty and close it (secondary check)
	count, _ := s.roomRepo.CountActiveParticipants(roomID)
	if count == 0 {
		room.IsActive = false
		s.roomRepo.Update(room)

		// Create call history
		s.createCallHistory(room)

		// Delete room from LiveKit
		if s.lkClient != nil {
			s.lkClient.DeleteRoom(ctx, &livekit.DeleteRoomRequest{
				Room: room.ID.String(),
			})
		}
	}

	return nil
}

func (s *roomService) EndRoom(ctx context.Context, roomID, userID uuid.UUID) error {
	room, err := s.roomRepo.GetByID(roomID)
	if err != nil {
		return ErrRoomNotFound
	}

	// Only creator can end the room
	if room.CreatorID != userID {
		return errors.New("only room creator can end the room")
	}

	room.IsActive = false
	if err := s.roomRepo.Update(room); err != nil {
		return fmt.Errorf("failed to end room: %w", err)
	}

	// Create call history
	s.createCallHistory(room)

	// Delete room from LiveKit
	if s.lkClient != nil {
		s.lkClient.DeleteRoom(ctx, &livekit.DeleteRoomRequest{
			Room: room.ID.String(),
		})
	}

	return nil
}

func (s *roomService) GetParticipants(ctx context.Context, roomID uuid.UUID) ([]domain.ParticipantResponse, error) {
	participants, err := s.roomRepo.GetActiveParticipants(roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}

	responses := make([]domain.ParticipantResponse, len(participants))
	for i, p := range participants {
		responses[i] = domain.ParticipantResponse{
			ID:       p.ID.String(),
			UserID:   p.UserID.String(),
			Name:     p.Name,
			JoinedAt: p.JoinedAt,
			LeftAt:   p.LeftAt,
			IsActive: p.IsActive,
		}
	}

	return responses, nil
}

func (s *roomService) generateLiveKitToken(roomName, userID, userName string) (string, error) {
	if s.lkConfig.APIKey == "" || s.lkConfig.APISecret == "" {
		return "", errors.New("LiveKit credentials not configured")
	}

	at := auth.NewAccessToken(s.lkConfig.APIKey, s.lkConfig.APISecret)
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomName,
	}
	at.AddGrant(grant).
		SetIdentity(userID).
		SetName(userName).
		SetValidFor(24 * time.Hour)

	return at.ToJWT()
}

// createCallHistory creates a call history record when a room ends
func (s *roomService) createCallHistory(room *domain.Room) {
	// Get all participants (including those who left)
	participants, err := s.roomRepo.GetAllParticipants(room.ID)
	if err != nil {
		s.logger.Error("Failed to get participants for call history", zap.Error(err))
		return
	}

	if len(participants) == 0 {
		s.logger.Debug("No participants found, skipping call history creation")
		return
	}

	// Calculate call duration
	endedAt := time.Now()
	startedAt := room.CreatedAt
	durationSeconds := int(endedAt.Sub(startedAt).Seconds())

	// Deduplication Logic: Group by UserID
	type UserStats struct {
		JoinedAt time.Time
		LeftAt   time.Time
	}
	userStatsMap := make(map[uuid.UUID]UserStats)

	for _, p := range participants {
		leftAt := endedAt
		if p.LeftAt != nil {
			leftAt = *p.LeftAt
		}

		stats, exists := userStatsMap[p.UserID]
		if !exists {
			userStatsMap[p.UserID] = UserStats{
				JoinedAt: p.JoinedAt,
				LeftAt:   leftAt,
			}
		} else {
			// If already exists, take the earliest join and latest leave
			if p.JoinedAt.Before(stats.JoinedAt) {
				stats.JoinedAt = p.JoinedAt
			}
			if leftAt.After(stats.LeftAt) {
				stats.LeftAt = leftAt
			}
			userStatsMap[p.UserID] = stats
		}
	}

	// Create call history
	history := &domain.CallHistory{
		RoomID:            room.ID,
		RoomName:          room.Name,
		WorkspaceID:       room.WorkspaceID,
		CreatorID:         room.CreatorID,
		StartedAt:         startedAt,
		EndedAt:           endedAt,
		DurationSeconds:   durationSeconds,
		MaxParticipants:   room.MaxParticipants,
		TotalParticipants: len(userStatsMap),
	}

	// Create CallHistoryParticipant list from map
	historyParticipants := make([]domain.CallHistoryParticipant, 0, len(userStatsMap))
	for uid, stats := range userStatsMap {
		participantDuration := int(stats.LeftAt.Sub(stats.JoinedAt).Seconds())

		historyParticipants = append(historyParticipants, domain.CallHistoryParticipant{
			UserID:          uid,
			JoinedAt:        stats.JoinedAt,
			LeftAt:          stats.LeftAt,
			DurationSeconds: participantDuration,
		})
	}

	history.Participants = historyParticipants

	if err := s.roomRepo.CreateCallHistory(history); err != nil {
		s.logger.Error("Failed to create call history", zap.Error(err))
		return
	}

	s.logger.Info("Call history created",
		zap.String("room_id", room.ID.String()),
		zap.String("room_name", room.Name),
		zap.Int("duration_seconds", durationSeconds),
		zap.Int("total_participants", len(historyParticipants)),
	)
}

// GetWorkspaceCallHistory returns call history for a workspace
func (s *roomService) GetWorkspaceCallHistory(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, token string, limit, offset int) ([]domain.CallHistoryResponse, int64, error) {
	// Validate workspace membership
	if s.userClient != nil {
		isMember, err := s.userClient.ValidateWorkspaceMember(ctx, workspaceID, userID, token)
		if err != nil {
			s.logger.Error("Failed to validate workspace membership", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to validate workspace membership: %w", err)
		}
		if !isMember {
			return nil, 0, ErrNotWorkspaceMember
		}
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	histories, total, err := s.roomRepo.GetCallHistoryByWorkspace(workspaceID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get call history: %w", err)
	}

	responses := make([]domain.CallHistoryResponse, len(histories))
	for i, h := range histories {
		responses[i] = h.ToResponse()
	}

	return responses, total, nil
}

// GetUserCallHistory returns call history for a user
func (s *roomService) GetUserCallHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.CallHistoryResponse, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	histories, total, err := s.roomRepo.GetCallHistoryByUser(userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get call history: %w", err)
	}

	responses := make([]domain.CallHistoryResponse, len(histories))
	for i, h := range histories {
		responses[i] = h.ToResponse()
	}

	return responses, total, nil
}

// GetCallHistoryByID gets a single call history by ID
func (s *roomService) GetCallHistoryByID(ctx context.Context, historyID uuid.UUID) (*domain.CallHistoryResponse, error) {
	history, err := s.roomRepo.GetCallHistoryByID(historyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get call history: %w", err)
	}
	if history == nil {
		return nil, nil
	}

	response := history.ToResponse()
	return &response, nil
}

// SaveTranscript saves or updates the transcript for a room
func (s *roomService) SaveTranscript(ctx context.Context, roomID uuid.UUID, content string) (*domain.TranscriptResponse, error) {
	// Find the call history for this room
	// First try to get existing transcript
	existingTranscript, _ := s.roomRepo.GetTranscriptByRoomID(roomID)

	if existingTranscript != nil {
		// Update existing transcript
		existingTranscript.Content = content
		if err := s.roomRepo.SaveTranscript(existingTranscript); err != nil {
			return nil, fmt.Errorf("failed to update transcript: %w", err)
		}
		response := existingTranscript.ToResponse()
		return &response, nil
	}

	// Create new transcript - we'll store it with roomID first
	// The callHistoryID will be nil until the room ends
	transcript := &domain.CallTranscript{
		RoomID:  roomID,
		Content: content,
	}

	// Try to find if there's already a call history for this room
	// (in case the room has ended but transcript is being saved)
	histories, _, _ := s.roomRepo.GetCallHistoryByWorkspace(uuid.Nil, 100, 0)
	for _, h := range histories {
		if h.RoomID == roomID {
			transcript.CallHistoryID = h.ID
			break
		}
	}

	if err := s.roomRepo.SaveTranscript(transcript); err != nil {
		return nil, fmt.Errorf("failed to save transcript: %w", err)
	}

	s.logger.Info("Transcript saved",
		zap.String("room_id", roomID.String()),
		zap.Int("content_length", len(content)),
	)

	response := transcript.ToResponse()
	return &response, nil
}

// GetTranscriptByCallHistoryID returns the transcript for a call history
func (s *roomService) GetTranscriptByCallHistoryID(ctx context.Context, callHistoryID uuid.UUID) (*domain.TranscriptResponse, error) {
	transcript, err := s.roomRepo.GetTranscriptByCallHistoryID(callHistoryID)
	if err != nil {
		// Try to find by room ID from call history
		history, histErr := s.roomRepo.GetCallHistoryByID(callHistoryID)
		if histErr != nil {
			return nil, fmt.Errorf("call history not found: %w", histErr)
		}

		// Try to find transcript by room ID
		transcript, err = s.roomRepo.GetTranscriptByRoomID(history.RoomID)
		if err != nil {
			return nil, fmt.Errorf("transcript not found: %w", err)
		}
	}

	response := transcript.ToResponse()
	return &response, nil
}
