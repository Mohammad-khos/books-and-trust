package userHandler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"books-and-trust/services/api-gateway/internal/client"
	pb "books-and-trust/shared/proto/user"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRegisterUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]string
		mockBehavior   func(ctx context.Context, in *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error)
		expectedStatus int
	}{
		{
			name: "Success - Valid Registration",
			requestBody: map[string]string{
				"name":     "mamad",
				"username": "mms121",
				"email":    "mammad@test.com",
				"password": "StrongPassword123!",
			},
			mockBehavior: func(ctx context.Context, in *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
				return &pb.RegisterUserResponse{
					User: &pb.User{Id: "usr_123"},
				}, nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Failure - Password Not Matching Regex",
			requestBody: map[string]string{
				"name":     "mamad",
				"username": "mms121",
				"email":    "mammad@test.com",
				"password": "123",
			},
			mockBehavior: func(ctx context.Context, in *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
				return nil, status.Error(codes.InvalidArgument, "invalid password or email")
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Failure - Internal Server Error from Microservice",
			requestBody: map[string]string{
				"name":     "mamad",
				"username": "mms121",
				"email":    "mammad@test.com",
				"password": "StrongPassword123!",
			},
			mockBehavior: func(ctx context.Context, in *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
				return nil, status.Error(codes.Internal, "internal grpc error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGrpcClient := &mockUserServiceClient{registerFunc: tt.mockBehavior}
			handler := &UserHandler{
				userCLient: &client.UserClient{Client: mockGrpcClient},
				Logger:     zap.NewNop().Sugar(),
			}

			jsonBody, _ := json.Marshal(tt.requestBody)
			req, err := http.NewRequest(http.MethodPost, "/api/v1/users/register", bytes.NewBuffer(jsonBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.RegisterUserHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %d want %d. Body: %s", rr.Code, tt.expectedStatus, rr.Body.String())
			}
		})
	}
}

func TestLoginUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]string 
		mockBehavior   func(ctx context.Context, in *pb.LoginUserRequest) (*pb.LoginUserResponse, error)
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name: "Success - Valid Login",
			requestBody: map[string]string{
				"credential": "mammad@test.com",
				"password":   "StrongPassword123!",
			},
			mockBehavior: func(ctx context.Context, in *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
				return &pb.LoginUserResponse{
					User:        &pb.User{Id: "usr_123"},
					AccessToken: "mocked_jwt_token",
					TokenType:   "Bearer",
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name: "Failure - Invalid Email Format",
			requestBody: map[string]string{
				"credential": "invalid-email",
				"password":   "StrongPassword123!",
			},
			mockBehavior: func(ctx context.Context, in *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
				return nil, status.Error(codes.InvalidArgument, "invalid password or email")
			}, expectedStatus: http.StatusBadRequest,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGrpcClient := &mockUserServiceClient{loginFunc: tt.mockBehavior}
			handler := &UserHandler{
				userCLient: &client.UserClient{Client: mockGrpcClient},
				Logger:     zap.NewNop().Sugar(),
			}

			jsonBody, _ := json.Marshal(tt.requestBody)
			req, err := http.NewRequest(http.MethodPost, "/api/v1/users/login", bytes.NewBuffer(jsonBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.LoginUserHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.expectedStatus, rr.Code, rr.Body.String())
			}

			if tt.expectSuccess {
				var respMap map[string]interface{}
				_ = json.Unmarshal(rr.Body.Bytes(), &respMap)
				if respMap["status"] != "success" {
					t.Errorf("expected status 'success', got %v", respMap["status"])
				}
			}
		})
	}
}
