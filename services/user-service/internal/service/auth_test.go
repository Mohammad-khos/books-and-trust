package service_test

import (
	"context"
	"errors"
	"testing"

	"books-and-trust/services/user-service/internal/domain"
	"books-and-trust/services/user-service/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_CreateUser(t *testing.T) {
	ctx := context.Background()
	plainPassword := "MamadSecure123!"
	invalidPassword := "123"
	fakeHash := []byte("$2a$10$FakeHashForRegistration")

	t.Run("Success_User_Registration", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		mockHasher := new(MockPasswordHasher)
		mockAuth := new(MockAuthenticator)

		user := &domain.User{
			ID: uuid.New(),
			Name:     "Mohammad Mehdi",
			Username: "mamad_dev",
			Email:    "mamad@example.com",
			Password: domain.Password{
				Text: &plainPassword,
			},

		}

		mockHasher.On("Hash", plainPassword).Return(fakeHash, nil)
		mockRepo.On("Create", mock.Anything, user).Return(nil)

		svc := service.NewUserService(mockRepo, mockHasher, mockAuth)
		err := svc.CreateUser(ctx, user)

		assert.NoError(t, err)
		assert.Equal(t, fakeHash, user.Password.Hash)
		mockRepo.AssertExpectations(t)
		mockHasher.AssertExpectations(t)
	})

	t.Run("Failed_Registration_With_Invalid_Password_Regex", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		mockHasher := new(MockPasswordHasher)
		mockAuth := new(MockAuthenticator)

		user := &domain.User{
			Name:     "Mohammad Mehdi",
			Username: "mamad_dev",
			Email:    "mamad@example.com",
			Password: domain.Password{
				Text: &invalidPassword,
			},
		}

		svc := service.NewUserService(mockRepo, mockHasher, mockAuth)
		err := svc.CreateUser(ctx, user)

		assert.Error(t, err)
		mockHasher.AssertNotCalled(t, "Hash", mock.Anything)
		mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})
}

func TestUserService_LoginUser(t *testing.T) {
	ctx := context.Background()
	testUserID := uuid.New()
	correctPassword := "MamadSecure123!"
	wrongPassword := "WrongPass123!"
	fakeHash := []byte("$2a$10$FakeBcryptHashForTesting")
	expectedToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.dummy_token"

	dummyUser := &domain.User{
		ID:       testUserID,
		Name:     "Mohammad Mehdi",
		Username: "mamad_dev",
		Email:    "mamad@example.com",
		Password: domain.Password{
			Hash: fakeHash,
		},
	}

	t.Run("Success_Login_With_Valid_Credentials", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		mockHasher := new(MockPasswordHasher)
		mockAuth := new(MockAuthenticator)

		mockRepo.On("GetByEmailOrUsername", mock.Anything, "mamad_dev").Return(dummyUser, nil)
		mockHasher.On("Compare", correctPassword, fakeHash).Return(nil)
		mockAuth.On("GenerateToken", testUserID.String()).Return(expectedToken, nil)

		svc := service.NewUserService(mockRepo, mockHasher, mockAuth)
		user, token, err := svc.LoginUser(ctx, "mamad_dev", correctPassword)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, expectedToken, token)
		assert.Equal(t, "mamad_dev", user.Username)

		mockRepo.AssertExpectations(t)
		mockHasher.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})

	t.Run("Failed_Login_With_Wrong_Password", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		mockHasher := new(MockPasswordHasher)
		mockAuth := new(MockAuthenticator)

		mockRepo.On("GetByEmailOrUsername", mock.Anything, "mamad_dev").Return(dummyUser, nil)
		mockHasher.On("Compare", wrongPassword, fakeHash).Return(errors.New("crypto/bcrypt: hashed password mismatch"))

		svc := service.NewUserService(mockRepo, mockHasher, mockAuth)
		user, token, err := svc.LoginUser(ctx, "mamad_dev", wrongPassword)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Empty(t, token)

		mockAuth.AssertNotCalled(t, "GenerateToken", mock.Anything)
	})

	t.Run("Failed_Login_When_User_Not_Found", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		mockHasher := new(MockPasswordHasher)
		mockAuth := new(MockAuthenticator)

		mockRepo.On("GetByEmailOrUsername", mock.Anything, "unknown_user").Return(nil, domain.ErrResourceNotFound)

		svc := service.NewUserService(mockRepo, mockHasher, mockAuth)
		user, token, err := svc.LoginUser(ctx, "unknown_user", correctPassword)

		assert.ErrorIs(t, err, domain.ErrResourceNotFound)
		assert.Nil(t, user)
		assert.Empty(t, token)

		mockHasher.AssertNotCalled(t, "Compare", mock.Anything, mock.Anything)
		mockAuth.AssertNotCalled(t, "GenerateToken", mock.Anything)
	})
}
