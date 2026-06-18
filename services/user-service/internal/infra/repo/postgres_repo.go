package repo

import (
	"books-and-trust/services/user-service/internal/domain"
	"books-and-trust/shared/tracing"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"
)

type SQLRepository struct {
	db *gorm.DB
}

func NewSQLRepository(db *gorm.DB) *SQLRepository {
	return &SQLRepository{
		db: db,
	}
}

func (r *SQLRepository) Create(ctx context.Context, user *domain.User) error {
	tracer := tracing.GetTracer("user-service")
	ctx, span := tracer.Start(ctx, "SQLRepository.Create")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := r.db.WithContext(ctx).Create(user).Error
	if err == gorm.ErrDuplicatedKey {
		span.RecordError(err)
		span.SetStatus(codes.Error, "duplicate key violation")
		return domain.ErrResourceAleadyExists
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to insert user record")
		return err
	}
	return nil
}

func (r *SQLRepository) GetByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	tracer := tracing.GetTracer("user-service")
	ctx, span := tracer.Start(ctx, "SQLRepository.GetByID")
	defer span.End()

	span.SetAttributes(attribute.String("db.user.id", userID.String()))

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user *domain.User
	err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error

	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			span.SetStatus(codes.Error, "user record not found")
			return nil, domain.ErrResourceNotFound
		default:
			span.SetStatus(codes.Error, "database execution query failed")
			return nil, err
		}
	}

	return user, nil
}

func (r *SQLRepository) DeleteByID(ctx context.Context, userID uuid.UUID) error {
	tracer := tracing.GetTracer("user-service")
	ctx, span := tracer.Start(ctx, "SQLRepository.DeleteByID")
	defer span.End()

	span.SetAttributes(attribute.String("db.user.id", userID.String()))

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result := r.db.WithContext(ctx).
		Delete(&domain.User{}, "id = ?", userID)

	if result.Error != nil {
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, "database execution query failed")
		return result.Error
	}

	if result.RowsAffected == 0 {
		span.SetStatus(codes.Error, "user record not found for deletion")
		return domain.ErrResourceNotFound
	}

	return nil
}

func (r *SQLRepository) GetByEmailOrUsername(ctx context.Context, credential string) (*domain.User, error) {
	tracer := tracing.GetTracer("user-service")
	ctx, span := tracer.Start(ctx, "SQLRepository.GetByEmailOrUsername")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user *domain.User
	err := r.db.WithContext(ctx).Where("email = ? OR username = ?", credential, credential).First(&user).Error

	if err != nil {
		span.RecordError(err)
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			span.SetStatus(codes.Error, "user record not found by credential")
			return nil, domain.ErrResourceNotFound
		default:
			span.SetStatus(codes.Error, "database execution query failed")
			return nil, err
		}
	}
	return user, nil
}

func (r *SQLRepository) Update(ctx context.Context, user *domain.User) error {
	tracer := tracing.GetTracer("user-service")
	ctx, span := tracer.Start(ctx, "SQLRepository.Update")
	defer span.End()

	if user != nil {
		span.SetAttributes(attribute.String("db.user.id", user.ID.String()))
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	updates := make(map[string]any)

	if user.Name != "" {
		updates["name"] = user.Name
	}
	if user.Username != "" {
		updates["username"] = user.Username
	}
	if user.Email != "" {
		updates["email"] = user.Email
	}
	if len(user.Password.Hash) > 0 {
		updates["password"] = user.Password.Hash
	}
	if len(updates) == 0 {
		span.SetStatus(codes.Error, "no update fields specified")
		return domain.ErrNoFieldsToUpdate
	}

	err := r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", user.ID).Updates(updates).Error
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update user record in db")
		return err
	}

	return nil
}
