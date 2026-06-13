package handler

import (
	"books-and-trust/services/loan-service/internal/domain"
	pb "books-and-trust/shared/proto/loan"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type gRPCHandler struct {
	svc domain.LoanService
	pb.UnimplementedLoanServiceServer
}

func NewGRPCHandler(svc domain.LoanService) *gRPCHandler {
	return &gRPCHandler{
		svc: svc,
	}
}

func (h *gRPCHandler) mapDomainToPb(l *domain.Loan) *pb.Loan {
	res := &pb.Loan{
		Id:        l.ID.String(),
		OwnerId:   l.OwnerID.String(),
		UserId:    l.UserID.String(),
		BookName:  l.BookName,
		Status:    pb.LoanStatus(pb.LoanStatus_value[l.Status]),
		CreatedAt: timestamppb.New(l.CreatedAt),
	}
	if l.Deadline != nil {
		res.Deadline = timestamppb.New(*l.Deadline)
	}
	if l.ReturnedAt != nil && !l.ReturnedAt.IsZero() {
		res.ReturnedAt = timestamppb.New(*l.ReturnedAt)
	}

	return res
}
