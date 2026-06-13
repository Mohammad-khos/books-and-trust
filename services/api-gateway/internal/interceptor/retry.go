package interceptor

import (
	"books-and-trust/shared/retry"
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var cfg = retry.Config{
	Attempts: 3,
	Initial: time.Millisecond * 100,
	MaxDelay: time.Second * 2,
	Factor: 2.0,
}

func RetryInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		return retry.Do(ctx, cfg, func() error {
			return invoker(ctx, method, req, reply, cc, opts...)
		}, func(err error) bool {
			st, ok := status.FromError(err)
			if !ok {
				return true 
			}
			
			switch st.Code() {
			case codes.Unavailable,       
				 codes.DeadlineExceeded,  
				 codes.Internal,          
				 codes.ResourceExhausted:
				return true
			}
			return false
		})
	}
}