package repo

import (
	"books-and-trust/services/loan-service/internal/domain"
	"books-and-trust/shared/tracing"
	"context"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (r *PostgresRepository) BanUser(ctx context.Context, bannedUser *domain.BannedUser) error {
	tracer := tracing.GetTracer("loan-service")
	ctx, span := tracer.Start(ctx, "postgres.BanUser")
	defer span.End()

	if bannedUser != nil {
		span.SetAttributes(attribute.String("banned.user.id", bannedUser.UserID.String()))
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	err := r.db.WithContext(ctx).Create(bannedUser).Error
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to insert banned user into postgres")
		return err
	}
	return nil
}

func (r *PostgresRepository) IsUserBanned(ctx context.Context, userID uuid.UUID) (bool, error) {
	tracer := tracing.GetTracer("loan-service")
	ctx, span := tracer.Start(ctx, "postgres.IsUserBanned")
	defer span.End()

	span.SetAttributes(attribute.String("check.user.id", userID.String()))

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	var count int64
	err := r.db.WithContext(ctx).Model(&domain.BannedUser{}).
		Where("user_id = ? AND (expired_at IS NULL OR expired_at > ?)", userID, time.Now()).
		Count(&count).Error

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to count banned user records in postgres")
		return false, err
	}

	isBanned := count > 0
	span.SetAttributes(attribute.Bool("user.is_banned", isBanned))

	return isBanned, nil
}
