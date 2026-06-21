package service

import (
	"books-and-trust/services/user-service/internal/domain"
	"books-and-trust/shared/tracing"
	"books-and-trust/shared/validation"
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (s *UserService) CreateUser(ctx context.Context, user *domain.User) error {
	tracer := tracing.GetTracer("user-service")
	ctx, span := tracer.Start(ctx, "UserService.CreateUser")
	defer span.End()

	//set span attributes
	span.SetAttributes(attribute.String("user.id", user.ID.String()))

	user.Password = domain.Password{
		Text: user.Password.Text,
	}
	if err := validation.Validator.Struct(user); err != nil {
		span.RecordError(domain.ErrValidationFailed)
		span.SetStatus(codes.Error, "user validation failed")
		return domain.ErrValidationFailed
	}
	//check regex
	isMatched := user.Password.IsMatched(*user.Password.Text)
	if !isMatched {
		span.RecordError(domain.ErrRegexNotMatched)
		span.SetStatus(codes.Error, "password regex not matched")
		return domain.ErrRegexNotMatched
	}

	if err := user.Password.GenerateHash(*user.Password.Text, s.hasher); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to generate password hash")
		return err
	}

	err := s.repo.Create(ctx, user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create user in db")
		return err
	}
	return nil
}

func (s *UserService) LoginUser(ctx context.Context, credential string, password string) (*domain.User, string, error) {
	tracer := tracing.GetTracer("user-service")
	ctx, span := tracer.Start(ctx, "UserService.LoginUser")
	defer span.End()

	if credential == "" {
		span.SetStatus(codes.Error, "empty credential provided")
		return nil, "", domain.ErrInvalidCredential
	}

	user, err := s.repo.GetByEmailOrUsername(ctx, credential)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "authentication failed")
		return nil, "", err
	}

	span.SetAttributes(attribute.String("user.id", user.ID.String()))

	if !user.Password.IsCorrect(password, s.hasher) {
		span.SetStatus(codes.Error, "authentication failed")
		return nil, "", domain.ErrInvalidCredential
	}

	span.AddEvent("user_authenticated_successfully")

	token, err := s.authenticator.GenerateToken(user.ID.String())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to generate authentication token")
		return nil, "", err
	}

	return user, token, nil
}

func (s *UserService) VerifyToken(ctx context.Context, token string) (string, error) {
	tracer := tracing.GetTracer("user-service")
	_, span := tracer.Start(ctx, "UserService.VerifyToken")
	defer span.End()

	userID, err := s.authenticator.VerifyToken(token)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "token verification failed")
		return "", err
	}

	span.SetAttributes(attribute.String("user.id", userID))

	return userID, nil
}
