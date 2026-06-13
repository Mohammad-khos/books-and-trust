package loanHandler

import (
	pb "books-and-trust/shared/proto/loan"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type CreateLoanRequest struct {
	OwnerID  string    `json:"owner_id" validate:"required"`
	BookName string    `json:"book_name" validate:"required"`
	Deadline time.Time `json:"deadline"`
}

func (req *CreateLoanRequest) ToProto() *pb.CreateLoanRequest {
	var dl *timestamppb.Timestamp
	if !req.Deadline.IsZero() {
		dl = timestamppb.New(req.Deadline)
	}
	return &pb.CreateLoanRequest{
		OwnerId:  req.OwnerID,
		UserId:   "",
		BookName: req.BookName,
		Deadline: dl,
	}
}

type CreateLoanResponse struct {
	ID        string    `json:"id" validate:"required"`
	UserID    string    `json:"user_id" validate:"required"`
	OwnerID   string    `json:"owner_id,omitempty"`
	BookName  string    `json:"book_name" validate:"required"`
	Deadline  time.Time `json:"deadline"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdateLoanRequest struct {
	ID       string    `json:"id" validate:"required"`
	Status   string    `json:"status,omitempty"`
	BookName string    `json:"book_name,omitempty"`
	Deadline time.Time `json:"deadline"`
}

func (req *UpdateLoanRequest) ToProto(updaterID string) *pb.UpdateLoanRequest {
	return &pb.UpdateLoanRequest{
		Id:        req.ID,
		BookName:  req.BookName,
		Status:    pb.LoanStatus(pb.LoanStatus_value[req.Status]),
		Deadline:  timestamppb.New(req.Deadline),
		UpdaterId: updaterID,
	}
}

type Loan struct {
	ID        string    `json:"id" validate:"required" example:"loan-uuid-111"`
	UserID    string    `json:"user_id" validate:"required" example:"user-uuid-123"`
	OwnerID   string    `json:"owner_id,omitempty" example:"owner-uuid-456"`
	BookName  string    `json:"book_name" validate:"required" example:"The Go Programming Language"`
	Status    string    `json:"status,omitempty" example:"BORROWED"`
	Deadline  time.Time `json:"deadline" example:"2026-06-16T18:30:00Z"`
	CreatedAt time.Time `json:"created_at" example:"2026-06-03T12:00:00Z"`
}

type GetLoanByIDResponse struct {
	Loan Loan `json:"loan" validate:"required"`
}

type ListLoanByOwnerResponse []Loan

type BanUserRequest struct {
	UserID    string    `json:"user_id" validate:"required"`
	Reason    string    `json:"reason,omitempty"`
	ExpiredAT time.Time `json:"expired_at,omitempty"`
}

func (req *BanUserRequest) ToProto() *pb.BanUserRequest {
	return &pb.BanUserRequest{
		UserId:    req.UserID,
		Reason:    req.Reason,
		ExpiredAt: timestamppb.New(req.ExpiredAT),
	}
}

type DeliveryLoanResponse struct {
	DeliveryCode string `json:"delivery_code" example:"DEL-98765"`
}

type ConfirmDeliveryRequest struct {
	DeliveryCode string `json:"delivery_code" validate:"required,len=4"`
}
