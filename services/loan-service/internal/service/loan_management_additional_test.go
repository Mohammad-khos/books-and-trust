package service_test

import (
	"context"
	"testing"

	"books-and-trust/services/loan-service/internal/domain"
	"books-and-trust/services/loan-service/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type enhancedMockLoanRepository struct {
    mockLoanRepository
    onListLoansByOwner func(ownerID uuid.UUID) ([]*domain.Loan, error)
}

func (m *enhancedMockLoanRepository) ListLoansByOwner(ctx context.Context, ownerID uuid.UUID) ([]*domain.Loan, error) {
    return m.onListLoansByOwner(ownerID)
}

func TestUpdateLoanByUser(t *testing.T) {
    ownerID := uuid.New()
    borrowerID := uuid.New()
    loanID := uuid.New()

    t.Run("fails_when_owner_attempts_to_take_own_loan", func(t *testing.T) {
        repo := &mockLoanRepository{
            onGetByID: func(id uuid.UUID) (*domain.Loan, error) {
                return &domain.Loan{ID: loanID, OwnerID: ownerID, UserID: uuid.Nil, Status: "CREATED"}, nil
            },
            onIsUserBanned: func(userID uuid.UUID) (bool, error) {
                return false, nil
            },
        }

        svc := service.NewLoanService(repo)
        err := svc.UpdateLoanByUser(context.Background(), loanID, ownerID)

        assert.ErrorIs(t, err, domain.ErrPermissionDenied)
    })

    t.Run("successfully_assigns_loan_to_user", func(t *testing.T) {
        var updatedLoan *domain.Loan
        repo := &mockLoanRepository{
            onGetByID: func(id uuid.UUID) (*domain.Loan, error) {
                return &domain.Loan{ID: loanID, OwnerID: ownerID, UserID: uuid.Nil, Status: "CREATED"}, nil
            },
            onIsUserBanned: func(userID uuid.UUID) (bool, error) {
                return false, nil
            },
            onUpdate: func(loan *domain.Loan) error {
                updatedLoan = loan
                return nil
            },
        }

        svc := service.NewLoanService(repo)
        err := svc.UpdateLoanByUser(context.Background(), loanID, borrowerID)

        assert.NoError(t, err)
        assert.NotNil(t, updatedLoan)
        assert.Equal(t, borrowerID, updatedLoan.UserID)
        assert.Equal(t, "ACTIVE", updatedLoan.Status)
    })
}

func TestGetLoanByID(t *testing.T) {
    loanID := uuid.New()

    t.Run("returns_not_found_when_repo_fails", func(t *testing.T) {
        repo := &mockLoanRepository{
            onGetByID: func(id uuid.UUID) (*domain.Loan, error) {
                return nil, domain.ErrNotFound
            },
        }

        svc := service.NewLoanService(repo)
        _, err := svc.GetLoanByID(context.Background(), loanID)

        assert.ErrorIs(t, err, domain.ErrNotFound)
    })

    t.Run("returns_existing_loan", func(t *testing.T) {
        expectedLoan := &domain.Loan{ID: loanID, OwnerID: uuid.New(), BookName: "Refactoring"}
        repo := &mockLoanRepository{
            onGetByID: func(id uuid.UUID) (*domain.Loan, error) {
                return expectedLoan, nil
            },
        }

        svc := service.NewLoanService(repo)
        loan, err := svc.GetLoanByID(context.Background(), loanID)

        assert.NoError(t, err)
        assert.Equal(t, expectedLoan, loan)
    })
}

func TestListLoanByOwner(t *testing.T) {
    ownerID := uuid.New()

    repo := &enhancedMockLoanRepository{
        mockLoanRepository: mockLoanRepository{},
        onListLoansByOwner: func(ownerID uuid.UUID) ([]*domain.Loan, error) {
            return []*domain.Loan{
                {ID: uuid.New(), OwnerID: ownerID, BookName: "Go in Action"},
                {ID: uuid.New(), OwnerID: ownerID, BookName: "Database Internals"},
            }, nil
        },
    }

    svc := service.NewLoanService(repo)
    loans, err := svc.ListLoanByOwner(context.Background(), ownerID)

    assert.NoError(t, err)
    assert.Len(t, loans, 2)
    assert.Equal(t, ownerID, loans[0].OwnerID)
}

func TestCreateLoan_InvalidLoan(t *testing.T) {
    svc := service.NewLoanService(nil)
    _, err := svc.CreateLoan(context.Background(), &domain.Loan{OwnerID: uuid.Nil, BookName: ""})

    assert.ErrorIs(t, err, domain.ErrValidation)
}
