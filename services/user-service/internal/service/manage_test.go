package service_test

import (
	"books-and-trust/services/user-service/internal/domain"
	"books-and-trust/services/user-service/internal/service"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_GetUserByID(t *testing.T) {
	ctx := context.Background()
	validUUID := uuid.New()
	dummyUser := &domain.User{ID: validUUID, Username: "mamad_dev"}

	t.Run("Success_Get_User", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		mockRepo.On("GetByID", mock.Anything, validUUID).Return(dummyUser, nil)

		svc := service.NewUserService(mockRepo, new(MockPasswordHasher), new(MockAuthenticator))
		user, err := svc.GetUserByID(ctx, validUUID.String())

		assert.NoError(t, err)
		assert.Equal(t, "mamad_dev", user.Username)
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	existingName := "Mamad Old Name"
	newName := "Mohammad Mehdi"
	newPassword := "MamadNewSecure123!"
	fakeHash := []byte("$2a$10$FakeHashForNewPassword")

	existingUser := &domain.User{
		ID:       userID,
		Name:     existingName,
		Username: "mamad_dev",
		Email:    "mamad@example.com",
	}

	t.Run("Success_Update_User_With_Password", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		mockHasher := new(MockPasswordHasher)
		mockAuth := new(MockAuthenticator)

		updateInput := &domain.User{
			ID:   userID,
			Name: newName,
			Password: domain.Password{
				Text: &newPassword,
			},
		}

		mockRepo.On("GetByID", mock.Anything, userID).Return(existingUser, nil).Once()
		mockHasher.On("Hash", newPassword).Return(fakeHash, nil)
		mockRepo.On("Update", mock.Anything, updateInput).Return(nil)

		updatedUserMock := &domain.User{
			ID:    userID,
			Name:  newName,
			Email: "mamad@example.com",
		}
		mockRepo.On("GetByID", mock.Anything, userID).Return(updatedUserMock, nil).Once()

		svc := service.NewUserService(mockRepo, mockHasher, mockAuth)
		result, err := svc.UpdateUser(ctx, updateInput)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newName, result.Name)
		assert.Equal(t, fakeHash, updateInput.Password.Hash)

		mockRepo.AssertExpectations(t)
		mockHasher.AssertExpectations(t)
	})

	t.Run("Failed_Update_User_Not_Found", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		mockHasher := new(MockPasswordHasher)
		mockAuth := new(MockAuthenticator)

		unknownID := uuid.New()
		updateInput := &domain.User{ID: unknownID, Name: "New Name"}

		mockRepo.On("GetByID", mock.Anything, unknownID).Return(nil, domain.ErrResourceNotFound)

		svc := service.NewUserService(mockRepo, mockHasher, mockAuth)
		result, err := svc.UpdateUser(ctx, updateInput)

		assert.ErrorIs(t, err, domain.ErrResourceNotFound)
		assert.Nil(t, result)

		mockRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
		mockHasher.AssertNotCalled(t, "Hash", mock.Anything)
	})
}
