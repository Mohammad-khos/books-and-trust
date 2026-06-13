package domain

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(context.Context, *User) error
	GetByID(context.Context, uuid.UUID) (*User, error)
	GetByEmailOrUsername(context.Context , string) (*User , error)
	DeleteByID(context.Context, uuid.UUID) error
	Update(context.Context, *User) error
}
