package service_test

import (
	"context"
	"testing"
	"time"

	"books-and-trust/services/loan-service/internal/domain"
	"books-and-trust/services/loan-service/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBanUser(t *testing.T) {
	targetUserID := uuid.New()
	futureExpiration := time.Now().Add(time.Hour * 48)

	t.Run("success_ban_user", func(t *testing.T) {
		repo := &mockLoanRepository{
			onBanUser: func(b *domain.BannedUser) error {
				return nil
			},
		}

		repo.onBanUser = func(b *domain.BannedUser) error {
			assert.Equal(t, targetUserID, b.UserID)
			assert.Equal(t, "Overdue book return", b.Reason)
			return nil
		}

		svc := service.NewLoanService(repo)
		banInput := &domain.BannedUser{
			UserID:    targetUserID,
			Reason:    "Overdue book return",
			ExpiredAt: &futureExpiration,
		}

		err := svc.BanUser(context.Background(), banInput)
		assert.NoError(t, err)
	})

	t.Run("failed_validation_empty_reason", func(t *testing.T) {
		svc := service.NewLoanService(&mockLoanRepository{})
		banInput := &domain.BannedUser{
			UserID: targetUserID,
			Reason: "", // 🚨 دلیل بن نباید خالی باشد
		}

		err := svc.BanUser(context.Background(), banInput)
		assert.Error(t, err)
	})
}