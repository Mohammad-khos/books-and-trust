package domain

import (
	"context"
)

type UserService interface {
	CreateUser(context.Context, *User) error
	LoginUser(ctx context.Context, credential string, password string) (*User, string, error)
	UpdateUser(context.Context, *User) (*User, error)
	GetUserByID(context.Context, string) (*User, error)
	DeleteUserByID(context.Context, string) error
	VerifyToken(context.Context,string) (string, error)
}
