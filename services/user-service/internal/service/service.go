package service

import (
	"books-and-trust/services/user-service/internal/domain"
)

type UserService struct {
	repo          domain.UserRepository
	hasher        domain.PasswordHasher
	authenticator domain.Authenticator
}

func NewUserService(repo domain.UserRepository, hasher domain.PasswordHasher, authenticator domain.Authenticator) *UserService {
	return &UserService{
		repo:          repo,
		hasher:        hasher,
		authenticator: authenticator,
	}
}
