package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type LoanRepository interface {
	Create(context.Context, *Loan) (*Loan, error)
	Update(context.Context, *Loan) error
	UpdateStatusAfterDelivery(context.Context, uuid.UUID, string, time.Time) error
	GetByID(context.Context, uuid.UUID) (*Loan, error)
	ListLoansByOwner(context.Context, uuid.UUID) ([]*Loan, error)
	BanUser(context.Context, *BannedUser) error
	IsUserBanned(context.Context, uuid.UUID) (bool, error)
}
