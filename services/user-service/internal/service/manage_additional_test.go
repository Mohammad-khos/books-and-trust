package service_test

import (
	"context"
	"testing"

	"books-and-trust/services/user-service/internal/domain"
	"books-and-trust/services/user-service/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteUserByID(t *testing.T) {
    t.Run("invalid_uuid_returns_error", func(t *testing.T) {
        repo := new(MockUserRepository)
        svc := service.NewUserService(repo, nil, nil)

        err := svc.DeleteUserByID(context.Background(), "not-a-uuid")

        assert.Error(t, err)
    })

    t.Run("repo_return_not_found", func(t *testing.T) {
        repo := new(MockUserRepository)
        repo.On("DeleteByID", mock.Anything, mock.Anything).Return(domain.ErrResourceNotFound)

        svc := service.NewUserService(repo, nil, nil)
        randomUUID := uuid.New().String()

        err := svc.DeleteUserByID(context.Background(), randomUUID)

        assert.ErrorIs(t, err, domain.ErrResourceNotFound)
        repo.AssertExpectations(t)
    })
}

func TestGetUserByID(t *testing.T) {
    t.Run("invalid_uuid_returns_error", func(t *testing.T) {
        repo := new(MockUserRepository)
        svc := service.NewUserService(repo, nil, nil)

        _, err := svc.GetUserByID(context.Background(), "bad-uuid")

        assert.Error(t, err)
    })
}

func TestUpdateUser(t *testing.T) {
    validUserID := uuid.New()

    t.Run("password_regex_mismatch_returns_error", func(t *testing.T) {
        repo := new(MockUserRepository)
        repo.On("GetByID", mock.Anything, validUserID).Return(&domain.User{ID: validUserID}, nil)

        svc := service.NewUserService(repo, nil, nil)
        invalidPassword := "weak"
        _, err := svc.UpdateUser(context.Background(), &domain.User{ID: validUserID, Password: domain.Password{Text: &invalidPassword}})

        assert.ErrorIs(t, err, domain.ErrRegexNotMatched)
        repo.AssertExpectations(t)
    })
}
