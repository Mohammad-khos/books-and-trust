package service

import (
	"books-and-trust/services/loan-service/internal/domain"
	"books-and-trust/shared/tracing"
	"books-and-trust/shared/validation"
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (s *LoanService) BanUser(ctx context.Context, bannedUser *domain.BannedUser) error {
	tracer := tracing.GetTracer("loan-service")
	ctx, span := tracer.Start(ctx, "LoanService.BanUser")
	defer span.End()

	if bannedUser != nil {
		span.SetAttributes(attribute.String("banned.user.id", bannedUser.UserID.String()))
	}

	if err := validation.Validator.Struct(bannedUser); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "banned user validation failed")
		return err
	}

	err := s.repo.BanUser(ctx, bannedUser)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to ban user in database")
		return err
	}

	return nil
}

