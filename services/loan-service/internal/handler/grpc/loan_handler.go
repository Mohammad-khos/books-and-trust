package handler

import (
	"books-and-trust/services/loan-service/internal/domain"
	pb "books-and-trust/shared/proto/loan"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (h *gRPCHandler) CreateLoan(ctx context.Context, req *pb.CreateLoanRequest) (*pb.CreateLoanResponse, error) {
	var ownerID uuid.UUID
	if req.GetOwnerId() != "" {
		id, err := uuid.Parse(req.GetOwnerId())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid owner id format")
		}
		ownerID = id
	}
	var deadlinePtr *time.Time
	if req.GetDeadline() != nil {
		t := req.GetDeadline().AsTime()
		deadlinePtr = &t
	}

	loan := &domain.Loan{
		OwnerID:  ownerID,
		UserID:   uuid.Nil,
		BookName: req.GetBookName(),
		Deadline: deadlinePtr,
		Status:   "unspecified",
	}

	createdLoan, err := h.svc.CreateLoan(ctx, loan)
	if err != nil {
		if errors.Is(err, domain.ErrUserIsOnBannedUsers) {
			return nil, status.Error(codes.PermissionDenied, "user is banned")
		}
		if errors.Is(err, domain.ErrSelfOperationNotAllowed) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, domain.ErrValidation) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateLoanResponse{
		Loan: h.mapDomainToPb(createdLoan),
	}, nil
}

func (h *gRPCHandler) ClaimLoan(ctx context.Context, in *pb.ClaimLoanRequest) (*emptypb.Empty, error) {
	loanID, err := uuid.Parse(in.GetLoanId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid loan_id format")
	}

	updaterID, err := uuid.Parse(in.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	err = h.svc.UpdateLoanByUser(ctx, loanID, updaterID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			return nil, status.Error(codes.NotFound, "loan record not found")

		case errors.Is(err, domain.ErrPermissionDenied):
			return nil, status.Error(codes.PermissionDenied, "you cannot claim your own loan position")

		case errors.Is(err, domain.ErrLoanAlreadyTaken):
			return nil, status.Error(codes.AlreadyExists, "this loan has already been claimed by another user")

		case errors.Is(err, domain.ErrUserIsOnBannedUsers):
			return nil, status.Error(codes.PermissionDenied, "user is banned and cannot claim loans")

		case errors.Is(err, domain.ErrInvalidFieldsToUpdate):
			return nil, status.Error(codes.InvalidArgument, "invalid fields provided for update")
		case errors.Is(err, domain.ErrLoanReturned):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	return &emptypb.Empty{}, nil
}
func (h *gRPCHandler) UpdateLoan(ctx context.Context, req *pb.UpdateLoanRequest) (*emptypb.Empty, error) {
	loanID, errID := uuid.Parse(req.GetId())
	updaterID, errUpdater := uuid.Parse(req.GetUpdaterId())
	if errID != nil || errUpdater != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id or updater_id format in request")
	}

	var deadlinePtr *time.Time
	if req.GetDeadline() != nil {
		t := req.GetDeadline().AsTime()
		deadlinePtr = &t
	}

	err := h.svc.UpdateLoan(ctx, &domain.Loan{
		ID:       loanID,
		BookName: req.GetBookName(),
		Status:   req.GetStatus().String(),
		Deadline: deadlinePtr,
	}, updaterID)

	if err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			return nil, status.Error(codes.PermissionDenied, "you do not have permission to update this loan")
		}
		if errors.Is(err, domain.ErrUserIsOnBannedUsers) {
			return nil, status.Error(codes.PermissionDenied, "this user already banned")
		}
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "loan record not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
func (h *gRPCHandler) GetLoanByID(ctx context.Context, req *pb.GetLoanByIDRequest) (*pb.GetLoanByIDResponse, error) {
	loanID, err := uuid.Parse(req.GetLoanId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid loan id format")
	}

	loan, err := h.svc.GetLoanByID(ctx, loanID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "loan not found")
		}
		return nil, status.Error(codes.Internal, "failed to get loan")
	}

	return &pb.GetLoanByIDResponse{
		Loan: h.mapDomainToPb(loan),
	}, nil
}

func (h *gRPCHandler) ListLoansByOwner(ctx context.Context, req *pb.ListLoansByOwnerRequest) (*pb.ListLoansByOwnerResponse, error) {
	ownerID, err := uuid.Parse(req.GetOwnerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner format")
	}

	loans, err := h.svc.ListLoanByOwner(ctx, ownerID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "loans not found")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}

	var pbLoans []*pb.Loan
	for _, loan := range loans {
		pbLoans = append(pbLoans, h.mapDomainToPb(loan))
	}

	return &pb.ListLoansByOwnerResponse{
		Loans: pbLoans,
	}, nil
}
