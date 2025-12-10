package domain

import (
	"time"

	"github.com/google/uuid"
)

// BaseModel contains common fields for all domain entities
type BaseModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time  `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"not null" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"index" json:"deletedAt,omitempty"`
}
