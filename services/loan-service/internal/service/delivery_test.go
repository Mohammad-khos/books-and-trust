package service_test

import (
	"books-and-trust/services/loan-service/internal/domain"
	"books-and-trust/services/loan-service/internal/service"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestLoanService_DeliveryLoan(t *testing.T) {
	
	targetLoanID := uuid.New()
	borrowerID := uuid.New() 
	strangerID := uuid.New()

	tests := []struct {
		name          string
		loanID        uuid.UUID
		updaterID     uuid.UUID
		mockSetup     func(m *mockLoanRepository)
		wantErr       bool
		expectedError error
	}{
		{
			name:      "Success - Borrower requests delivery code",
			loanID:    targetLoanID,
			updaterID: borrowerID, 
			mockSetup: func(m *mockLoanRepository) {
				m.onGetByID = func(id uuid.UUID) (*domain.Loan, error) {
					return &domain.Loan{
						ID:      targetLoanID,
						UserID: borrowerID,
						}, nil
					}
					m.onUpdate = func(loan *domain.Loan) error {
					assert.Len(t, loan.DeliveryCode, 4)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:      "Failure - Unauthorized user requests delivery code",
			loanID:    targetLoanID,
			updaterID: strangerID, 
			mockSetup: func(m *mockLoanRepository) {
				m.onGetByID = func(id uuid.UUID) (*domain.Loan, error) {
					return &domain.Loan{
						ID:      targetLoanID,
						OwnerID: borrowerID,
					}, nil
				}
				m.onUpdateDeliveryCode = func(id uuid.UUID, code string) error {
					t.Fail()
					return nil
				}
			},
			wantErr:       true,
			expectedError: domain.ErrPermissionDenied, 
		},
		{
			name:      "Failure - Loan record not found",
			loanID:    targetLoanID,
			updaterID: borrowerID,
			mockSetup: func(m *mockLoanRepository) {
				m.onGetByID = func(id uuid.UUID) (*domain.Loan, error) {
					return nil, domain.ErrNotFound
				}
			},
			wantErr:       true,
			expectedError: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockLoanRepository{}
			tt.mockSetup(mockRepo)

			service := service.NewLoanService(mockRepo)


			code, err := service.DeliveryLoan(context.Background(), tt.loanID, tt.updaterID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, code)
			}
		})
	}
}
