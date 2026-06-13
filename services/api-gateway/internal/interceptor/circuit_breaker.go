package interceptor

import (
	circuitBreaker "books-and-trust/services/api-gateway/internal/infra/circuit-breaker"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CircuitBreakerInterceptor(cb circuitBreaker.Breaker) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		
		var grpcErr error

		_, cbErr := cb.Execute(func() (any, error) {
			grpcErr = invoker(ctx, method, req, reply, cc, opts...)
			if grpcErr != nil {
				st, _ := status.FromError(grpcErr)
				if st.Code() == codes.InvalidArgument || st.Code() == codes.NotFound {
					return nil, nil 
				}
				return nil, grpcErr
			}
			return nil, nil
		})

		if cbErr != nil && grpcErr == nil {
			return cbErr 
		}

		return grpcErr
	}
}