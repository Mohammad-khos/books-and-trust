package service

import (
	"books-and-trust/services/user-service/internal/domain"
	"context"
	"books-and-trust/shared/tracing"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (s *UserService) DeleteUserByID(ctx context.Context, userID string) error {
	tracer := tracing.GetTracer("user-service")
	ctx, span := tracer.Start(ctx, "UserService.DeleteUserByID")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID))

	uuid, err := uuid.Parse(userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to parse uuid")
		return err
	}
	
	if err := s.repo.DeleteByID(ctx, uuid); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete user from db")
		return err
	}
	return nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	tracer := tracing.GetTracer("user-service")
	ctx, span := tracer.Start(ctx, "UserService.GetUserByID")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID))

	uuid, err := uuid.Parse(userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to parse uuid")
		return nil, err
	}

	user, err := s.repo.GetByID(ctx, uuid)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch user from db")
		return nil, err
	}
	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	tracer := tracing.GetTracer("user-service")
	ctx, span := tracer.Start(ctx, "UserService.UpdateUser")
	defer span.End()

	if user != nil {
		span.SetAttributes(attribute.String("user.id", user.ID.String()))
	}

	_, err := s.repo.GetByID(ctx, user.ID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to check user existence in db")
		return nil, err
	}

	if user.Password.Text != nil && *user.Password.Text != "" {
		if !user.Password.IsMatched(*user.Password.Text) {
			span.SetStatus(codes.Error, "password regex not matched")
			return nil, domain.ErrRegexNotMatched
		}
		if err := user.Password.GenerateHash(*user.Password.Text, s.hasher); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to generate password hash")
			return nil, err
		}
	}

	if err := s.repo.Update(ctx, user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update user in db")
		return nil, err
	}

	updatedUser, err := s.repo.GetByID(ctx, user.ID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch updated user from db")
		return nil, err
	}
	
	return updatedUser, nil
}