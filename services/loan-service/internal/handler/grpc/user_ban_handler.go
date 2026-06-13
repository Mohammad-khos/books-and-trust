package handler

import (
	"books-and-trust/services/loan-service/internal/domain"
	pb "books-and-trust/shared/proto/loan"
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (h *gRPCHandler) BanUser(ctx context.Context, req *pb.BanUserRequest) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id format")
	}
	var expPtr *time.Time
	if req.GetExpiredAt() != nil {
		t := req.GetExpiredAt().AsTime()
		expPtr = &t
	}

	bannedUser := &domain.BannedUser{
		UserID:    userID,
		Reason:    req.GetReason(),
		ExpiredAt: expPtr,
	}

	if err := h.svc.BanUser(ctx, bannedUser); err != nil {
		return nil, status.Error(codes.Internal, "failed to ban user")
	}

	return &emptypb.Empty{}, nil
}
