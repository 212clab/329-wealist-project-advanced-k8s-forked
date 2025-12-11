package service

import (
	"chat-service/internal/client"
	"chat-service/internal/domain"
	"chat-service/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type ChatService struct {
	chatRepo    *repository.ChatRepository
	messageRepo *repository.MessageRepository
	userClient  client.UserClient
	redis       *redis.Client
	logger      *zap.Logger
}

func NewChatService(
	chatRepo *repository.ChatRepository,
	messageRepo *repository.MessageRepository,
	userClient client.UserClient,
	redis *redis.Client,
	logger *zap.Logger,
) *ChatService {
	return &ChatService{
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
		userClient:  userClient,
		redis:       redis,
		logger:      logger,
	}
}

func (s *ChatService) CreateChat(ctx context.Context, req *domain.CreateChatRequest, createdBy uuid.UUID) (*domain.Chat, error) {
	chat := &domain.Chat{
		ID:          uuid.New(),
		WorkspaceID: req.WorkspaceID,
		ProjectID:   req.ProjectID,
		ChatType:    req.ChatType,
		ChatName:    req.ChatName,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.chatRepo.Create(chat); err != nil {
		return nil, err
	}

	// Ensure creator is in participants
	participantIDs := append([]uuid.UUID{createdBy}, req.Participants...)
	uniqueIDs := make(map[uuid.UUID]bool)
	for _, id := range participantIDs {
		uniqueIDs[id] = true
	}

	var uniqueParticipants []uuid.UUID
	for id := range uniqueIDs {
		uniqueParticipants = append(uniqueParticipants, id)
	}

	if err := s.chatRepo.AddParticipants(chat.ID, uniqueParticipants); err != nil {
		return nil, err
	}

	// Reload with participants
	return s.chatRepo.GetByID(chat.ID)
}

func (s *ChatService) GetChatByID(ctx context.Context, chatID uuid.UUID) (*domain.Chat, error) {
	return s.chatRepo.GetByID(chatID)
}

func (s *ChatService) GetUserChats(ctx context.Context, userID uuid.UUID) ([]domain.ChatWithUnread, error) {
	return s.chatRepo.GetUserChats(userID)
}

func (s *ChatService) GetWorkspaceChats(ctx context.Context, workspaceID uuid.UUID) ([]domain.Chat, error) {
	return s.chatRepo.GetWorkspaceChats(workspaceID)
}

func (s *ChatService) DeleteChat(ctx context.Context, chatID uuid.UUID) error {
	return s.chatRepo.SoftDelete(chatID)
}

func (s *ChatService) AddParticipants(ctx context.Context, chatID uuid.UUID, userIDs []uuid.UUID) error {
	return s.chatRepo.AddParticipants(chatID, userIDs)
}

func (s *ChatService) RemoveParticipant(ctx context.Context, chatID, userID uuid.UUID) error {
	return s.chatRepo.RemoveParticipant(chatID, userID)
}

func (s *ChatService) IsUserInChat(ctx context.Context, chatID, userID uuid.UUID) (bool, error) {
	return s.chatRepo.IsUserInChat(chatID, userID)
}

func (s *ChatService) SendMessage(ctx context.Context, chatID, userID uuid.UUID, req *domain.SendMessageRequest) (*domain.Message, error) {
	messageType := domain.MessageTypeText
	if req.MessageType != "" {
		messageType = req.MessageType
	}

	message := &domain.Message{
		ID:          uuid.New(),
		ChatID:      chatID,
		UserID:      userID,
		Content:     req.Content,
		MessageType: messageType,
		FileURL:     req.FileURL,
		FileName:    req.FileName,
		FileSize:    req.FileSize,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.messageRepo.Create(message); err != nil {
		return nil, err
	}

	// Update chat timestamp
	s.chatRepo.UpdateTimestamp(chatID)

	// Publish to Redis for WebSocket broadcast
	s.publishMessage(ctx, chatID, message)

	return message, nil
}

func (s *ChatService) GetMessages(ctx context.Context, chatID uuid.UUID, limit int, before *uuid.UUID) ([]domain.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.messageRepo.GetByChatID(chatID, limit, before)
}

func (s *ChatService) DeleteMessage(ctx context.Context, messageID uuid.UUID) error {
	return s.messageRepo.SoftDelete(messageID)
}

func (s *ChatService) MarkMessagesAsRead(ctx context.Context, messageIDs []uuid.UUID, userID uuid.UUID) error {
	return s.messageRepo.MarkMultipleAsRead(messageIDs, userID)
}

func (s *ChatService) UpdateLastReadAt(ctx context.Context, chatID, userID uuid.UUID) error {
	return s.chatRepo.UpdateLastReadAt(chatID, userID)
}

func (s *ChatService) GetUnreadCount(ctx context.Context, chatID, userID uuid.UUID) (int64, error) {
	chat, err := s.chatRepo.GetByID(chatID)
	if err != nil {
		return 0, err
	}

	var lastReadAt *time.Time
	for _, p := range chat.Participants {
		if p.UserID == userID {
			lastReadAt = p.LastReadAt
			break
		}
	}

	return s.messageRepo.GetUnreadCount(chatID, userID, lastReadAt)
}

func (s *ChatService) publishMessage(ctx context.Context, chatID uuid.UUID, message *domain.Message) {
	if s.redis == nil {
		return
	}

	channel := fmt.Sprintf("chat:%s", chatID.String())
	data, err := json.Marshal(map[string]interface{}{
		"type":    "MESSAGE_RECEIVED",
		"message": message,
	})
	if err != nil {
		s.logger.Error("failed to marshal message for publish", zap.Error(err))
		return
	}

	if err := s.redis.Publish(ctx, channel, data).Err(); err != nil {
		s.logger.Error("failed to publish message", zap.Error(err))
	}
}
