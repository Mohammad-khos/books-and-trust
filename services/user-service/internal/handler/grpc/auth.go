package grpc

import (
	"books-and-trust/services/user-service/internal/domain"
	pb "books-and-trust/shared/proto/user"
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *gRPCHandler) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	password := req.GetPassword()

	user := &domain.User{
		Name:     req.GetName(),
		Username: req.GetUsername(),
		Email:    req.GetEmail(),
		Password: domain.Password{
			Text: &password,
		},
	}
	err := h.svc.CreateUser(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrRegexNotMatched):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, domain.ErrValidationFailed):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, domain.ErrResourceAleadyExists):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to create user")
		}
	}

	resp := &pb.RegisterUserResponse{
		User: &pb.User{
			Id:       user.ID.String(),
			Name:     user.Name,
			Username: user.Username,
			Email:    user.Email,
		},
	}
	return resp, nil
}
func (h *gRPCHandler) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	credential := req.GetUsernameOrEmail()
	password := req.GetPassword()
	user, token, err := h.svc.LoginUser(ctx, credential, password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrResourceNotFound), errors.Is(err, domain.ErrInvalidCredential):
			return nil, status.Error(codes.Unauthenticated, "invalid username or password")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}
	resp := &pb.LoginUserResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		User: &pb.User{
			Id:       user.ID.String(),
			Name:     user.Name,
			Email:    user.Email,
			Username: user.Username,
		},
	}
	return resp, nil
}

func (h *gRPCHandler) VerifyToken(ctx context.Context, req *pb.VerifyTokenRequest) (*pb.VerifyTokenResponse, error) {
	token := req.GetToken()
	userID, err := h.svc.VerifyToken(ctx , token)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidToken):
			return nil, status.Error(codes.InvalidArgument, "invalid jwt token")
		default:
			status.Error(codes.Internal, "failed to verify jwt token")
		}
	}
	return &pb.VerifyTokenResponse{
		UserId: userID,
	}, nil
}
