package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"books-and-trust/services/loan-service/internal/domain"
	"books-and-trust/services/loan-service/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockLoanRepository struct {
	domain.LoanRepository
	onDropCreate   func(loan *domain.Loan) (*domain.Loan, error)
	onIsUserBanned func(userID uuid.UUID) (bool, error)
	onUpdate       func(loan *domain.Loan) error
	onBanUser      func(bannedUser *domain.BannedUser) error
	onGetByID            func(id uuid.UUID) (*domain.Loan, error)
	onUpdateDeliveryCode func(id uuid.UUID, code string) error
}

func (m *mockLoanRepository) Create(ctx context.Context, loan *domain.Loan) (*domain.Loan, error) {
	return m.onDropCreate(loan)
}

func (m *mockLoanRepository) IsUserBanned(ctx context.Context, userID uuid.UUID) (bool, error) {
	return m.onIsUserBanned(userID)
}

func (m *mockLoanRepository) Update(ctx context.Context, loan *domain.Loan) error {
	return m.onUpdate(loan)
}
func (m *mockLoanRepository) BanUser(ctx context.Context, bannedUser *domain.BannedUser) error {
	return m.onBanUser(bannedUser)
}

func (m *mockLoanRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Loan, error) {
	return m.onGetByID(id)
}

func (m *mockLoanRepository) UpdateDeliveryCode(ctx context.Context, id uuid.UUID, code string) error {
	return m.onUpdateDeliveryCode(id , code)
}

func TestCreateLoan(t *testing.T) {
	ownerID := uuid.New()
	userID := uuid.New()
	futureDeadline := time.Now().Add(time.Hour * 24)

	t.Run("success_create_loan", func(t *testing.T) {
		repo := &mockLoanRepository{
			onIsUserBanned: func(u uuid.UUID) (bool, error) {
				return false, nil
			},
			onDropCreate: func(l *domain.Loan) (*domain.Loan, error) {
				if l.ID == uuid.Nil {
					l.ID = uuid.New()
				}
				return l, nil
			},
		}

		svc := service.NewLoanService(repo)
		loanInput := &domain.Loan{
			OwnerID:  ownerID,
			UserID:   userID,
			BookName: "Clean Architecture",
			Deadline: &futureDeadline,
			Status:   "unspecified",
		}

		result, err := svc.CreateLoan(context.Background(), loanInput)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEqual(t, uuid.Nil, result.ID)
		assert.Equal(t, "unspecified", result.Status)
	})

	t.Run("failed_user_is_banned", func(t *testing.T) {
		repo := &mockLoanRepository{
			onIsUserBanned: func(u uuid.UUID) (bool, error) {
				return true, nil
			},
		}

		svc := service.NewLoanService(repo)
		loanInput := &domain.Loan{
			OwnerID:  ownerID,
			UserID:   userID,
			BookName: "Go Programming",
			Deadline: &futureDeadline,
			Status:   "unspecified",
		}

		result, err := svc.CreateLoan(context.Background(), loanInput)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, domain.ErrUserIsOnBannedUsers)
	})
}

func TestUpdateLoan(t *testing.T) {
	t.Run("failed_missing_loan_id", func(t *testing.T) {
		repo := &mockLoanRepository{
			onUpdate: func(l *domain.Loan) error {
				return errors.New("invalid fields to update")
			},
		}

		svc := service.NewLoanService(repo)
		futureDeadline := time.Now().Add(time.Hour * 24)

		loanInput := &domain.Loan{
			ID:       uuid.Nil,
			OwnerID:  uuid.New(),
			UserID:   uuid.New(),
			BookName: "Refactoring",
			Status:   "active",
			Deadline: &futureDeadline,
		}

		err := svc.UpdateLoan(context.Background(), loanInput , loanInput.UserID)

		assert.ErrorContains(t, err, "invalid fields to update")
	})
}
