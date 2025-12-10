package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Room represents a video call room
type Room struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Name        string         `gorm:"size:255;not null" json:"name"`
	WorkspaceID uuid.UUID      `gorm:"type:uuid;not null;index" json:"workspaceId"`
	CreatorID   uuid.UUID      `gorm:"type:uuid;not null" json:"creatorId"`
	MaxParticipants int        `gorm:"default:10" json:"maxParticipants"`
	IsActive    bool           `gorm:"default:true" json:"isActive"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Participants []RoomParticipant `gorm:"foreignKey:RoomID" json:"participants,omitempty"`
}

func (r *Room) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// RoomParticipant represents a user in a video call room
type RoomParticipant struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	RoomID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"roomId"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"userId"`
	JoinedAt  time.Time      `json:"joinedAt"`
	LeftAt    *time.Time     `json:"leftAt,omitempty"`
	IsActive  bool           `gorm:"default:true" json:"isActive"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (rp *RoomParticipant) BeforeCreate(tx *gorm.DB) error {
	if rp.ID == uuid.Nil {
		rp.ID = uuid.New()
	}
	if rp.JoinedAt.IsZero() {
		rp.JoinedAt = time.Now()
	}
	return nil
}

// DTOs
type CreateRoomRequest struct {
	Name            string `json:"name" binding:"required"`
	WorkspaceID     string `json:"workspaceId" binding:"required,uuid"`
	MaxParticipants int    `json:"maxParticipants"`
}

type JoinRoomRequest struct {
	RoomID string `json:"roomId" binding:"required,uuid"`
}

type RoomResponse struct {
	ID              string                `json:"id"`
	Name            string                `json:"name"`
	WorkspaceID     string                `json:"workspaceId"`
	CreatorID       string                `json:"creatorId"`
	MaxParticipants int                   `json:"maxParticipants"`
	IsActive        bool                  `json:"isActive"`
	ParticipantCount int                  `json:"participantCount"`
	Participants    []ParticipantResponse `json:"participants,omitempty"`
	CreatedAt       time.Time             `json:"createdAt"`
	UpdatedAt       time.Time             `json:"updatedAt"`
}

type ParticipantResponse struct {
	ID       string     `json:"id"`
	UserID   string     `json:"userId"`
	JoinedAt time.Time  `json:"joinedAt"`
	LeftAt   *time.Time `json:"leftAt,omitempty"`
	IsActive bool       `json:"isActive"`
}

type JoinRoomResponse struct {
	Room     RoomResponse `json:"room"`
	Token    string       `json:"token"`
	WSUrl    string       `json:"wsUrl"`
}

// CallHistory represents a completed video call record
type CallHistory struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	RoomID          uuid.UUID      `gorm:"type:uuid;not null;index" json:"roomId"`
	RoomName        string         `gorm:"size:255;not null" json:"roomName"`
	WorkspaceID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"workspaceId"`
	CreatorID       uuid.UUID      `gorm:"type:uuid;not null" json:"creatorId"`
	StartedAt       time.Time      `json:"startedAt"`
	EndedAt         time.Time      `json:"endedAt"`
	DurationSeconds int            `json:"durationSeconds"`
	MaxParticipants int            `json:"maxParticipants"`
	TotalParticipants int          `json:"totalParticipants"`
	CreatedAt       time.Time      `json:"createdAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Participants who joined this call
	Participants []CallHistoryParticipant `gorm:"foreignKey:CallHistoryID" json:"participants,omitempty"`
}

func (ch *CallHistory) BeforeCreate(tx *gorm.DB) error {
	if ch.ID == uuid.Nil {
		ch.ID = uuid.New()
	}
	return nil
}

// CallHistoryParticipant represents a participant in a completed call
type CallHistoryParticipant struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	CallHistoryID uuid.UUID      `gorm:"type:uuid;not null;index" json:"callHistoryId"`
	UserID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"userId"`
	JoinedAt      time.Time      `json:"joinedAt"`
	LeftAt        time.Time      `json:"leftAt"`
	DurationSeconds int          `json:"durationSeconds"`
	CreatedAt     time.Time      `json:"createdAt"`
}

func (chp *CallHistoryParticipant) BeforeCreate(tx *gorm.DB) error {
	if chp.ID == uuid.Nil {
		chp.ID = uuid.New()
	}
	return nil
}

// CallHistoryResponse represents call history for API response
type CallHistoryResponse struct {
	ID                string                        `json:"id"`
	RoomName          string                        `json:"roomName"`
	WorkspaceID       string                        `json:"workspaceId"`
	CreatorID         string                        `json:"creatorId"`
	StartedAt         time.Time                     `json:"startedAt"`
	EndedAt           time.Time                     `json:"endedAt"`
	DurationSeconds   int                           `json:"durationSeconds"`
	TotalParticipants int                           `json:"totalParticipants"`
	Participants      []CallHistoryParticipantResponse `json:"participants,omitempty"`
}

type CallHistoryParticipantResponse struct {
	UserID          string    `json:"userId"`
	JoinedAt        time.Time `json:"joinedAt"`
	LeftAt          time.Time `json:"leftAt"`
	DurationSeconds int       `json:"durationSeconds"`
}

// CallTranscript represents the transcript/subtitles of a video call
type CallTranscript struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	CallHistoryID uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex" json:"callHistoryId"`
	RoomID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"roomId"`
	Content       string         `gorm:"type:text" json:"content"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}

func (ct *CallTranscript) BeforeCreate(tx *gorm.DB) error {
	if ct.ID == uuid.Nil {
		ct.ID = uuid.New()
	}
	return nil
}

// SaveTranscriptRequest represents request to save a transcript
type SaveTranscriptRequest struct {
	RoomID  string `json:"roomId" binding:"required,uuid"`
	Content string `json:"content" binding:"required"`
}

// TranscriptResponse represents transcript for API response
type TranscriptResponse struct {
	ID            string    `json:"id"`
	CallHistoryID string    `json:"callHistoryId"`
	RoomID        string    `json:"roomId"`
	Content       string    `json:"content"`
	CreatedAt     time.Time `json:"createdAt"`
}

func (ct *CallTranscript) ToResponse() TranscriptResponse {
	return TranscriptResponse{
		ID:            ct.ID.String(),
		CallHistoryID: ct.CallHistoryID.String(),
		RoomID:        ct.RoomID.String(),
		Content:       ct.Content,
		CreatedAt:     ct.CreatedAt,
	}
}

func (ch *CallHistory) ToResponse() CallHistoryResponse {
	participants := make([]CallHistoryParticipantResponse, len(ch.Participants))
	for i, p := range ch.Participants {
		participants[i] = CallHistoryParticipantResponse{
			UserID:          p.UserID.String(),
			JoinedAt:        p.JoinedAt,
			LeftAt:          p.LeftAt,
			DurationSeconds: p.DurationSeconds,
		}
	}

	return CallHistoryResponse{
		ID:                ch.ID.String(),
		RoomName:          ch.RoomName,
		WorkspaceID:       ch.WorkspaceID.String(),
		CreatorID:         ch.CreatorID.String(),
		StartedAt:         ch.StartedAt,
		EndedAt:           ch.EndedAt,
		DurationSeconds:   ch.DurationSeconds,
		TotalParticipants: ch.TotalParticipants,
		Participants:      participants,
	}
}

func (r *Room) ToResponse() RoomResponse {
	participantCount := 0
	participants := make([]ParticipantResponse, 0)

	for _, p := range r.Participants {
		if p.IsActive {
			participantCount++
			participants = append(participants, ParticipantResponse{
				ID:       p.ID.String(),
				UserID:   p.UserID.String(),
				JoinedAt: p.JoinedAt,
				LeftAt:   p.LeftAt,
				IsActive: p.IsActive,
			})
		}
	}

	return RoomResponse{
		ID:               r.ID.String(),
		Name:             r.Name,
		WorkspaceID:      r.WorkspaceID.String(),
		CreatorID:        r.CreatorID.String(),
		MaxParticipants:  r.MaxParticipants,
		IsActive:         r.IsActive,
		ParticipantCount: participantCount,
		Participants:     participants,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}
