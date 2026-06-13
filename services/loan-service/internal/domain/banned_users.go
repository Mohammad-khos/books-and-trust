package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BannedUser struct {
	UserID    uuid.UUID      `json:"user_id" gorm:"primaryKey" validate:"required,uuid"`
	Reason    string         `json:"reason" validate:"required,min=3,max=500"`
	ExpiredAt *time.Time     `json:"expired_at,omitempty" validate:"omitempty,gtfield=CreatedAt"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
