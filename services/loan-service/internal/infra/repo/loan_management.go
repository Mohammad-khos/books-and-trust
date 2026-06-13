package repo

import (
	"books-and-trust/services/loan-service/internal/domain"
	"books-and-trust/shared/tracing"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"
)

func (r *PostgresRepository) Create(ctx context.Context, loan *domain.Loan) (*domain.Loan, error) {
	tracer := tracing.GetTracer("loan-service")
	ctx, span := tracer.Start(ctx, "postgres.Create")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	err := r.db.WithContext(ctx).Create(loan).Error
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to insert loan into postgres")
		return nil, err
	}
	return loan, nil
}

func (r *PostgresRepository) Update(ctx context.Context, loan *domain.Loan) error {
	tracer := tracing.GetTracer("loan-service")
	ctx, span := tracer.Start(ctx, "postgres.Update")
	defer span.End()

	if loan != nil {
		span.SetAttributes(attribute.String("loan.id", loan.ID.String()))
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	updates := make(map[string]any)
	if loan.UserID != uuid.Nil {
		updates["user_id"] = loan.UserID
	}
	if loan.Status != "" {
		updates["status"] = loan.Status
	}
	if loan.BookName != "" {
		updates["book_name"] = loan.BookName
	}
	if loan.ReturnedAt != nil && !loan.ReturnedAt.IsZero() {
		updates["returned_at"] = loan.ReturnedAt
	}
	if loan.DeliveryCode != "" {
		updates["delivery_code"] = loan.DeliveryCode
	}
	updates["deadline"] = loan.Deadline
	if len(updates) == 0 {
		span.SetStatus(codes.Error, "no dynamic fields to update")
		return domain.ErrNoFieldsToUpdate
	}

	err := r.db.WithContext(ctx).Model(&domain.Loan{}).Where("id = ?", loan.ID).Updates(updates).Error
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update loan in postgres")
		return err
	}

	return nil
}

func (r *PostgresRepository) UpdateStatusAfterDelivery(ctx context.Context, loanID uuid.UUID, status string, returnedAt time.Time) error {
	tracer := tracing.GetTracer("loan-service")
	ctx, span := tracer.Start(ctx, "postgres.UpdateStatusAfterDelivery")
	defer span.End()

	span.SetAttributes(
		attribute.String("loan.id", loanID.String()),
		attribute.String("status.target", status),
	)

	err := r.db.WithContext(ctx).Model(&domain.Loan{}).Where("id = ?", loanID).
		Updates(map[string]any{
			"status":        status,
			"delivery_code": "",
			"returned_at":   returnedAt,
			"updated_at":    returnedAt,
		}).Error

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update status after delivery")
		return err
	}

	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, loanID uuid.UUID) (*domain.Loan, error) {
	tracer := tracing.GetTracer("loan-service")
	ctx, span := tracer.Start(ctx, "postgres.GetByID")
	defer span.End()

	span.SetAttributes(attribute.String("loan.id", loanID.String()))

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	var loan domain.Loan
	err := r.db.WithContext(ctx).Where("id = ?", loanID).First(&loan).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			span.SetStatus(codes.Error, "loan record not found")
			return nil, domain.ErrNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "postgres single fetch error")
		return nil, err
	}
	return &loan, nil
}

func (r *PostgresRepository) ListLoansByOwner(ctx context.Context, ownerID uuid.UUID) ([]*domain.Loan, error) {
	tracer := tracing.GetTracer("loan-service")
	ctx, span := tracer.Start(ctx, "postgres.ListLoansByOwner")
	defer span.End()

	span.SetAttributes(attribute.String("owner.id", ownerID.String()))

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	var loans []*domain.Loan
	err := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).Find(&loans).Error
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "postgres list fetch error")
		return nil, err
	}
	if len(loans) == 0 {
		span.SetStatus(codes.Error, "no loans found for this owner")
		return nil, domain.ErrNotFound
	}

	span.SetAttributes(attribute.Int("loans.fetched.count", len(loans)))
	return loans, nil
}
