package client

import (
	pb "books-and-trust/shared/proto/user"
	"books-and-trust/shared/tracing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserClient struct {
	Client pb.UserServiceClient
	conn   *grpc.ClientConn
}

func NewUserClient(clientAddr string) (*UserClient, error) {
	conn, err := grpc.NewClient(
		clientAddr,
		append(tracing.DialOptionsWithTracing(), grpc.WithTransportCredentials(insecure.NewCredentials()))...,
	)
	if err != nil {
		return nil, err
	}
	client := pb.NewUserServiceClient(conn)
	return &UserClient{
		Client: client,
		conn:   conn,
	}, nil
}

func (u *UserClient) Close() {
	if u.conn != nil {
		if err := u.conn.Close(); err != nil {
			return
		}
	}
}
