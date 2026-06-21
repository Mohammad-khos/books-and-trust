package service_test

import (
	"context"
	"errors"
	"testing"

	"books-and-trust/services/user-service/internal/domain"
	"books-and-trust/services/user-service/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestVerifyToken(t *testing.T) {
    t.Run("success_returns_user_id", func(t *testing.T) {
        mockAuth := new(MockAuthenticator)
        expectedID := uuid.New().String()
        mockAuth.On("VerifyToken", "valid-token").Return(expectedID, nil)

        svc := service.NewUserService(nil, nil, mockAuth)
        userID, err := svc.VerifyToken(context.Background(), "valid-token")

        assert.NoError(t, err)
        assert.Equal(t, expectedID, userID)
        mockAuth.AssertExpectations(t)
    })

    t.Run("invalid_token_returns_error", func(t *testing.T) {
        mockAuth := new(MockAuthenticator)
        mockAuth.On("VerifyToken", "bad-token").Return("", errors.New("invalid token"))

        svc := service.NewUserService(nil, nil, mockAuth)
        _, err := svc.VerifyToken(context.Background(), "bad-token")

        assert.Error(t, err)
        mockAuth.AssertExpectations(t)
    })
}

func TestLoginUser_EmptyCredential(t *testing.T) {
    svc := service.NewUserService(nil, nil, nil)

    user, token, err := svc.LoginUser(context.Background(), "", "password")

    assert.ErrorIs(t, err, domain.ErrInvalidCredential)
    assert.Nil(t, user)
    assert.Empty(t, token)
}
