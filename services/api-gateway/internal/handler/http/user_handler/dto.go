package userHandler

import pb "books-and-trust/shared/proto/user"

type RegisterUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Username string `json:"username" validate:"required,min=3,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (req *RegisterUserRequest) ToProto() *pb.RegisterUserRequest {
	return &pb.RegisterUserRequest{
		Name:     req.Name,
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}
}

type RegisterUserResponse struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type LoginUserRequest struct {
	Credential string `json:"credential" validate:"required"`
	Password   string `json:"password" validate:"required"`
}

func (req *LoginUserRequest) ToProto() *pb.LoginUserRequest {
	return &pb.LoginUserRequest{
		UsernameOrEmail: req.Credential,
		Password:        req.Password,
	}
}

type LoginUserResponse struct {
	UserID    string `json:"user_id"`
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
}

type UpdateUserRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

func (req *UpdateUserRequest) ToProto() *pb.UpdateUserRequest {
	protoReq := &pb.UpdateUserRequest{
		UserId: req.UserID,
	}
	if req.Name != "" {
		protoReq.Name = &req.Name
	}
	if req.Email != "" {
		protoReq.Email = &req.Email
	}
	if req.Username != "" {
		protoReq.Username = &req.Username
	}
	if req.Password != "" {
		protoReq.Password = &req.Password
	}

	return protoReq
}

type UpdateUserResponse struct {
	UserID   string `json:"user_id" validate:"required"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty" validate:"email"`
	Username string `json:"username,omitempty"`
}

type GetUserByIDResponse struct {
	UserID   string `json:"user_id" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required"`
}
