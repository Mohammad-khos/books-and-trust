package domain

import (
	"context"

	"github.com/google/uuid"
)

type LoanService interface {
	CreateLoan(context.Context, *Loan) (*Loan, error)
	UpdateLoan(ctx context.Context, loan *Loan, updaterID uuid.UUID) error
	UpdateLoanByUser(ctx context.Context, loanID uuid.UUID, updaterID uuid.UUID) error
	GetLoanByID(ctx context.Context, loanID uuid.UUID) (*Loan, error)
	BanUser(ctx context.Context, bannedUser *BannedUser) error
	ListLoanByOwner(ctx context.Context, ownerID uuid.UUID) ([]*Loan, error)
	DeliveryLoan(ctx context.Context, loanID uuid.UUID, updaterID uuid.UUID) (string, error)
	ConfirmDelivery(ctx context.Context, loanID uuid.UUID,ownerID uuid.UUID , code string) error
}
