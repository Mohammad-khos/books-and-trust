package service_test

import (
	"books-and-trust/services/user-service/internal/domain"
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// --- MockUserRepository ---
type MockUserRepository struct{ mock.Mock }

func (m *MockUserRepository) Create(ctx context.Context, u *domain.User) error {
	 return m.Called(ctx, u).Error(0)
}
func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockUserRepository) GetByEmailOrUsername(ctx context.Context, cred string) (*domain.User, error) {
	args := m.Called(ctx, cred)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepository) Update(ctx context.Context, u *domain.User) error {
	args :=  m.Called(ctx , u)
	if args.Get(0) == nil {
		return nil
	}
	return args.Error(0)
}

// --- MockPasswordHasher ---
type MockPasswordHasher struct{ mock.Mock }

func (m *MockPasswordHasher) Hash(p string) ([]byte, error) {
	args := m.Called(p)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}
func (m *MockPasswordHasher) Compare(p string, h []byte) error { return m.Called(p, h).Error(0) }

// --- MockAuthenticator ---
type MockAuthenticator struct{ mock.Mock }

func (m *MockAuthenticator) GenerateToken(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}
func (m *MockAuthenticator) VerifyToken(t string) (string, error) {
	args := m.Called(t)
	return args.String(0), args.Error(1)
}
