package service

import (
	"books-and-trust/services/loan-service/internal/domain"
	"books-and-trust/shared/tracing"
	"books-and-trust/shared/validation"
	"context"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (s *LoanService) CreateLoan(ctx context.Context, loan *domain.Loan) (*domain.Loan, error) {
	tracer := tracing.GetTracer("loan-service")
	ctx, span := tracer.Start(ctx, "LoanService.CreateLoan")
	defer span.End()

	if err := validation.Validator.Struct(loan); err != nil {
		span.RecordError(domain.ErrValidation)
		span.SetStatus(codes.Error, "loan validation failed")
		return nil, domain.ErrValidation
	}

	isBann, err := s.repo.IsUserBanned(ctx, loan.OwnerID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to check if owner is banned")
		return nil, err
	}
	if isBann {
		span.SetStatus(codes.Error, "owner is banned from system")
		return nil, domain.ErrUserIsOnBannedUsers
	}

	if loan.UserID == loan.OwnerID {
		span.SetStatus(codes.Error, "self loan operation is forbidden")
		return nil, domain.ErrSelfOperationNotAllowed
	}

	return s.repo.Create(ctx, loan)
}

func (s *LoanService) UpdateLoan(ctx context.Context, loan *domain.Loan, updaterID uuid.UUID) error {
	tracer := tracing.GetTracer("loan-service")
	ctx, span := tracer.Start(ctx, "LoanService.UpdateLoan")
	defer span.End()

	if loan != nil {
		span.SetAttributes(attribute.String("loan.id", loan.ID.String()))
	}
	span.SetAttributes(attribute.String("updater.id", updaterID.String()))

	if loan.ID == uuid.Nil || updaterID == uuid.Nil {
		span.SetStatus(codes.Error, "missing required fields for update")
		return domain.ErrInvalidFieldsToUpdate
	}

	existingLoan, err := s.repo.GetByID(ctx, loan.ID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch existing loan")
		return err
	}

	if existingLoan.OwnerID != updaterID {
		span.SetStatus(codes.Error, "permission denied for updating loan")
		return domain.ErrPermissionDenied
	}

	if existingLoan.UserID != uuid.Nil {
		span.SetStatus(codes.Error, "cannot update an already taken loan")
		return domain.ErrLoanAlreadyTaken
	}

	existingLoan.BookName = loan.BookName
	existingLoan.Deadline = loan.Deadline
	existingLoan.UpdatedAt = time.Now()

	return s.repo.Update(ctx, existingLoan)
}

func (s *LoanService) UpdateLoanByUser(ctx context.Context, loanID uuid.UUID, updaterID uuid.UUID) error {
	tracer := tracing.GetTracer("loan-service")
	ctx, span := tracer.Start(ctx, "LoanService.UpdateLoanByUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("loan.id", loanID.String()),
		attribute.String("updater.id", updaterID.String()),
	)

	if loanID == uuid.Nil || updaterID == uuid.Nil {
		span.SetStatus(codes.Error, "invalid fields provided")
		return domain.ErrInvalidFieldsToUpdate
	}

	existingLoan, err := s.repo.GetByID(ctx, loanID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch loan by id")
		return err
	}

	if existingLoan.OwnerID == updaterID {
		span.SetStatus(codes.Error, "owner cannot take their own loan")
		return domain.ErrPermissionDenied
	}

	if existingLoan.UserID != uuid.Nil {
		span.SetStatus(codes.Error, "loan is already assigned to another user")
		return domain.ErrLoanAlreadyTaken
	}

	isBanned, err := s.repo.IsUserBanned(ctx, updaterID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to check user ban status")
		return err
	}
	if isBanned {
		span.SetStatus(codes.Error, "banned user attempted to take a loan")
		return domain.ErrUserIsOnBannedUsers
	}

	if existingLoan.Status == "RETURNED" {
		span.SetStatus(codes.Error, "cannot active a returned loan")
		return domain.ErrLoanReturned
	}

	existingLoan.UserID = updaterID
	existingLoan.Status = "ACTIVE"
	existingLoan.UpdatedAt = time.Now()

	return s.repo.Update(ctx, existingLoan)
}

func (s *LoanService) GetLoanByID(ctx context.Context, loanID uuid.UUID) (*domain.Loan, error) {
	tracer := tracing.GetTracer("loan-service")
	ctx, span := tracer.Start(ctx, "LoanService.GetLoanByID")
	defer span.End()

	span.SetAttributes(attribute.String("loan.id", loanID.String()))

	loan, err := s.repo.GetByID(ctx, loanID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "loan record not found or db error")
		return nil, err
	}

	return loan, nil
}

func (s *LoanService) ListLoanByOwner(ctx context.Context, ownerID uuid.UUID) ([]*domain.Loan, error) {
	tracer := tracing.GetTracer("loan-service")
	ctx, span := tracer.Start(ctx, "LoanService.ListLoanByOwner")
	defer span.End()

	span.SetAttributes(attribute.String("owner.id", ownerID.String()))

	loans, err := s.repo.ListLoansByOwner(ctx, ownerID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch loans list from database")
		return nil, err
	}

	span.SetAttributes(attribute.Int("loans.count", len(loans)))

	return loans, nil
}
