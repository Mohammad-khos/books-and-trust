package grpc

import (
	"books-and-trust/services/user-service/internal/domain"
	"books-and-trust/services/user-service/internal/service"
	pb "books-and-trust/shared/proto/user"
)

type gRPCHandler struct {
	svc domain.UserService
	pb.UnimplementedUserServiceServer
}

func NewGRPCHandler(svc *service.UserService) *gRPCHandler {
	return &gRPCHandler{
		svc: svc,
	}
}
