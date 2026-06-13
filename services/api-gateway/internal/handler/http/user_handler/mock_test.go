package userHandler

import (
	"context"
	pb "books-and-trust/shared/proto/user"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type mockUserServiceClient struct {
	pb.UserServiceClient
	registerFunc   func(ctx context.Context, in *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error)
	loginFunc      func(ctx context.Context, in *pb.LoginUserRequest) (*pb.LoginUserResponse, error)
	updateFunc     func(ctx context.Context, in *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error)
	getUserFunc    func(ctx context.Context, in *pb.GetUserByIDRequest) (*pb.GetUserByIDResponse, error)
	deleteUserFunc func(ctx context.Context, in *pb.DeleteUserByIDRequest) (*emptypb.Empty, error)
}

func (m *mockUserServiceClient) RegisterUser(ctx context.Context, in *pb.RegisterUserRequest, opts ...grpc.CallOption) (*pb.RegisterUserResponse, error) {
	if m.registerFunc == nil {
		return nil, nil
	}
	return m.registerFunc(ctx, in)
}

func (m *mockUserServiceClient) LoginUser(ctx context.Context, in *pb.LoginUserRequest, opts ...grpc.CallOption) (*pb.LoginUserResponse, error) {
	if m.loginFunc == nil {
		return nil, nil
	}
	return m.loginFunc(ctx, in)
}

func (m *mockUserServiceClient) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest, opts ...grpc.CallOption) (*pb.UpdateUserResponse, error) {
	if m.updateFunc == nil {
		return nil, nil
	}
	return m.updateFunc(ctx, in)
}

func (m *mockUserServiceClient) GetUserByID(ctx context.Context, in *pb.GetUserByIDRequest, opts ...grpc.CallOption) (*pb.GetUserByIDResponse, error) {
	if m.getUserFunc == nil {
		return nil, nil
	}
	return m.getUserFunc(ctx, in)
}

func (m *mockUserServiceClient) DeleteUserByID(ctx context.Context, in *pb.DeleteUserByIDRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if m.deleteUserFunc == nil {
		return nil, nil
	}
	return m.deleteUserFunc(ctx, in)
}
