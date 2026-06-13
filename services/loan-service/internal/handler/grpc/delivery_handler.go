package handler

import (
	"context"
	"errors"

	"books-and-trust/services/loan-service/internal/domain"
	pb "books-and-trust/shared/proto/loan"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (h *gRPCHandler) DeliveryLoan(ctx context.Context, req *pb.DeliveryLoanRequest) (*pb.DeliveryLoanResponse, error) {
	loanID, errID := uuid.Parse(req.GetLoanID())
	if errID != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid loan_id format in request")
	}
	updaterID, errUpdater := uuid.Parse(req.GetUpdaterId())
	if errUpdater != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid updater_id format in request")
	}

	code, err := h.svc.DeliveryLoan(ctx, loanID, updaterID)
	if err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			return nil, status.Error(codes.PermissionDenied, "only the borrower can request a delivery code")
		}
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "loan record not found")
		}
		return nil, status.Error(codes.Internal, "failed to generate delivery code")
	}

	return &pb.DeliveryLoanResponse{
		DeliveryCode: code,
	}, nil
}

func (h *gRPCHandler) ConfirmDelivery(ctx context.Context, req *pb.ConfirmDeliveryRequest) (*emptypb.Empty, error) {
	code := req.GetDeliveryCode()
	if code == "" {
		return nil, status.Error(codes.InvalidArgument, "missing delivery code or loan id")
	}
	loanID, err := uuid.Parse(req.GetLoanID())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid loan id format in request")
	}
	OwnerID, err := uuid.Parse(req.GetOwnerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner id format in request")
	}

	err = h.svc.ConfirmDelivery(ctx, loanID ,OwnerID, code)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPermissionDenied):
			return nil, status.Error(codes.PermissionDenied, "only the borrower can request a delivery code")
		case errors.Is(err, domain.ErrInvalidDeliveryCode):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, domain.ErrLoanAlreadyBeenDelivered):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, domain.ErrNotFound):
			return nil, status.Error(codes.NotFound, "loan not found")
		default:
			return nil, status.Error(codes.Internal, "failed to confirm delivery")
		}
	}
	return &emptypb.Empty{}, nil
}
