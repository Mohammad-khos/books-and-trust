package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Loan struct {
	ID         uuid.UUID      `json:"id" gorm:"primaryKey" validate:"omitempty,uuid"`
	OwnerID    uuid.UUID      `json:"owner_id" gorm:"index" validate:"required,uuid"`
	UserID     uuid.UUID      `json:"user_id" gorm:"index" validate:"omitempty,uuid"`
	BookName   string         `json:"book_name" validate:"required,min=3,max=255"`
	Status     string         `json:"status" validate:"max=255"`
	Deadline   *time.Time      `json:"deadline" validate:"omitempty,gtfield=CreatedAt"`
	DeliveryCode   string      `json:"delivery_code"`
	CreatedAt  time.Time      `json:"created_at"`
	ReturnedAt *time.Time      `json:"returned_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (l *Loan) BeforeCreate(tx *gorm.DB) (err error) {
	l.ID = uuid.New()
	return nil
}