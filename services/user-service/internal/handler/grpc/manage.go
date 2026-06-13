package grpc

import (
	"books-and-trust/services/user-service/internal/domain"
	pb "books-and-trust/shared/proto/user"
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (h *gRPCHandler) GetUserByID(ctx context.Context, req *pb.GetUserByIDRequest) (*pb.GetUserByIDResponse, error) {
	userID := req.GetId()
	user, err := h.svc.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrResourceNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to get user")
	}
	resp := &pb.GetUserByIDResponse{
		User: &pb.User{
			Id:       user.ID.String(),
			Name:     user.Name,
			Email:    user.Email,
			Username: user.Username,
		},
	}
	return resp, nil
}

func (h *gRPCHandler) DeleteUserByID(ctx context.Context, req *pb.DeleteUserByIDRequest) (*emptypb.Empty, error) {
	userID := req.GetUserId()
	if err := h.svc.DeleteUserByID(ctx, userID); err != nil {
		if errors.Is(err, domain.ErrResourceNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete user")
	}
	return &emptypb.Empty{}, nil
}

func (h *gRPCHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	userID := req.GetUserId()
	uuid, err := uuid.Parse(userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id format")
	}
	name := req.GetName()
	email := req.GetEmail()
	username := req.GetUsername()
	password := req.GetPassword()
	user := &domain.User{
		ID:       uuid,
		Name:     name,
		Username: username,
		Email:    email,
		Password: domain.Password{
			Text: &password,
		},
	}
	updatedUser, err := h.svc.UpdateUser(ctx, user)
    if err != nil {
        if errors.Is(err, domain.ErrResourceNotFound) {
            return nil, status.Error(codes.NotFound, "user not found")
        }
        if errors.Is(err, domain.ErrRegexNotMatched) {
            return nil, status.Error(codes.InvalidArgument, err.Error())
        }
        return nil, status.Error(codes.Internal,err.Error())
    }
	resp := &pb.UpdateUserResponse{
		User: &pb.User{
			Id:       updatedUser.ID.String(),
			Name:     updatedUser.Name,
			Email:    updatedUser.Email,
			Username: updatedUser.Username,
		},
	}
	return resp, nil
}
