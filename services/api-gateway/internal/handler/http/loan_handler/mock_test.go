package loanHandler_test

import (
	pb "books-and-trust/shared/proto/loan"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type mockLoanServiceClient struct {
	pb.LoanServiceClient
	mockBehavior         func(ctx context.Context, in *pb.CreateLoanRequest) (*pb.CreateLoanResponse, error)
	mockGetLoanByID      func(ctx context.Context, in *pb.GetLoanByIDRequest) (*pb.GetLoanByIDResponse, error)
	mockUpdateLoan       func(ctx context.Context, in *pb.UpdateLoanRequest) (*emptypb.Empty, error)
	mockListLoansByOwner func(ctx context.Context, in *pb.ListLoansByOwnerRequest) (*pb.ListLoansByOwnerResponse, error)
	mockBanUser          func(ctx context.Context, in *pb.BanUserRequest) (*emptypb.Empty, error)
	mockDeliveryLoan     func(ctx context.Context, in *pb.DeliveryLoanRequest) (*pb.DeliveryLoanResponse, error)
}

func (m *mockLoanServiceClient) CreateLoan(ctx context.Context, in *pb.CreateLoanRequest, opts ...grpc.CallOption) (*pb.CreateLoanResponse, error) {
	if m.mockBehavior == nil {
		return nil, nil
	}
	return m.mockBehavior(ctx, in)
}

func (m *mockLoanServiceClient) GetLoanByID(ctx context.Context, in *pb.GetLoanByIDRequest, opts ...grpc.CallOption) (*pb.GetLoanByIDResponse, error) {
	return m.mockGetLoanByID(ctx, in)
}

func (m *mockLoanServiceClient) UpdateLoan(ctx context.Context, in *pb.UpdateLoanRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.mockUpdateLoan(ctx, in)
}

func (m *mockLoanServiceClient) ListLoansByOwner(ctx context.Context, in *pb.ListLoansByOwnerRequest, opts ...grpc.CallOption) (*pb.ListLoansByOwnerResponse, error) {
	if m.mockListLoansByOwner == nil {
		return nil, nil
	}
	return m.mockListLoansByOwner(ctx, in)
}

func (m *mockLoanServiceClient) BanUser(ctx context.Context, in *pb.BanUserRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if m.mockBanUser == nil {
		return nil, nil
	}
	return m.mockBanUser(ctx, in)
}

func (m *mockLoanServiceClient) DeliveryLoan(ctx context.Context, in *pb.DeliveryLoanRequest, opts ...grpc.CallOption) (*pb.DeliveryLoanResponse, error) {
	if m.mockDeliveryLoan == nil {
		return nil, nil
	}
	return m.mockDeliveryLoan(ctx, in)
}