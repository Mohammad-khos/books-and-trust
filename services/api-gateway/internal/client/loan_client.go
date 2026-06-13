package client

import (
	circuitBreaker "books-and-trust/services/api-gateway/internal/infra/circuit-breaker"
	"books-and-trust/services/api-gateway/internal/interceptor"
	pb "books-and-trust/shared/proto/loan"
	"books-and-trust/shared/tracing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type LoanClient struct {
	Client pb.LoanServiceClient
	conn   *grpc.ClientConn
}

func NewLoanClient(clientAddr string, cb circuitBreaker.Breaker) (*LoanClient, error) {
	conn, err := grpc.NewClient(clientAddr, append(tracing.DialOptionsWithTracing(), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithChainUnaryInterceptor(
		interceptor.RetryInterceptor(),
		interceptor.CircuitBreakerInterceptor(cb)))...,
	)
	if err != nil {
		return nil, err
	}
	client := pb.NewLoanServiceClient(conn)
	return &LoanClient{
		Client: client,
		conn:   conn,
	}, nil
}

func (u *LoanClient) Close() {
	if u.conn != nil {
		if err := u.conn.Close(); err != nil {
			return
		}
	}
}
