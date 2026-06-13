package service

import (
	"books-and-trust/services/loan-service/internal/domain"
	"books-and-trust/shared/tracing"
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (s *LoanService) DeliveryLoan(ctx context.Context, loanID uuid.UUID, updaterID uuid.UUID) (string, error) {
	tracer := tracing.GetTracer("loan-service")
	ctx, span := tracer.Start(ctx, "LoanService.DeliveryLoan")
	defer span.End()

	span.SetAttributes(
		attribute.String("loan.id", loanID.String()),
		attribute.String("updater.id", updaterID.String()),
	)

	loan, err := s.repo.GetByID(ctx, loanID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch loan record for delivery")
		return "", err
	}

	if loan.UserID != updaterID {
		span.SetStatus(codes.Error, "permission denied: updater is not the borrower")
		return "", domain.ErrPermissionDenied
	}

	generatedCode := fmt.Sprintf("%04d", rand.Intn(10000))
	loan.DeliveryCode = generatedCode

	err = s.repo.Update(ctx, loan)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update loan record with delivery code")
		return "", err
	}

	return generatedCode, nil
}

func (s *LoanService) ConfirmDelivery(ctx context.Context, loanID uuid.UUID, ownerID uuid.UUID, code string) error {
	tracer := tracing.GetTracer("loan-service")
	ctx, span := tracer.Start(ctx, "LoanService.ConfirmDelivery")
	defer span.End()

	span.SetAttributes(
		attribute.String("loan.id", loanID.String()),
		attribute.String("owner.id", ownerID.String()),
	)

	loan, err := s.repo.GetByID(ctx, loanID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch loan for delivery confirmation")
		return err
	}

	if ownerID != loan.OwnerID {
		span.SetStatus(codes.Error, "permission denied: only owner can confirm delivery")
		return domain.ErrPermissionDenied
	}

	if loan.DeliveryCode != code {
		span.SetStatus(codes.Error, "invalid delivery code provided")
		return domain.ErrInvalidDeliveryCode
	}

	if loan.ReturnedAt != nil {
		span.SetStatus(codes.Error, "loan has already been delivered and completed")
		return domain.ErrLoanAlreadyBeenDelivered
	}

	err = s.repo.UpdateStatusAfterDelivery(ctx, loanID, "RETURNED", time.Now())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update loan status to returned in database")
		return err
	}

	return nil
}
