package userHandler

import (
	"books-and-trust/services/api-gateway/internal/client"
	pb "books-and-trust/shared/proto/user"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestUpdateUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]string
		mockBehavior   func(ctx context.Context, in *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error)
		expectedStatus int
	}{
		{
			name: "Success - Valid Update",
			requestBody: map[string]string{
				"user_id":  "usr_123",
				"name":     "Mammad Dev",
				"email":    "mammad_new@test.com",
				"username": "mammad_updated",
			},
			mockBehavior: func(ctx context.Context, in *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
				return &pb.UpdateUserResponse{
					User: &pb.User{
						Id:       in.UserId,
						Name:     *in.Name,
						Email:    *in.Email,
						Username: *in.Username,
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Failure - Invalid Email Format (From Microservice)",
			requestBody: map[string]string{
				"user_id":  "usr_123",
				"name":     "Mammad Dev",
				"email":    "invalid-email-format@test.com", // ارسال دیتای معتبر به ولیدیشن داخلی گیت‌وی
				"username": "mammad_updated",
			},
			mockBehavior: func(ctx context.Context, in *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
				// شلیک خطای ۴۰۰ از سمت میکرو سرویس برای جلوگیری از پنیک پکیج util گیت‌وی
				return nil, status.Error(codes.InvalidArgument, "invalid email format")
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Failure - Internal Server Error from Microservice",
			requestBody: map[string]string{
				"user_id":  "usr_123",
				"name":     "Mammad Dev",
				"email":    "mammad@test.com",
				"username": "mammad_updated",
			},
			mockBehavior: func(ctx context.Context, in *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
				return nil, status.Error(codes.Internal, "database error on user service")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGrpcClient := &mockUserServiceClient{updateFunc: tt.mockBehavior}
			handler := &UserHandler{
				userCLient: &client.UserClient{Client: mockGrpcClient},
				Logger:     zap.NewNop().Sugar(),
			}

			jsonBody, _ := json.Marshal(tt.requestBody)
			req, err := http.NewRequest(http.MethodPost, "/api/v1/users/update", bytes.NewBuffer(jsonBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.UpdateUser(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %d want %d. Body: %s", rr.Code, tt.expectedStatus, rr.Body.String())
			}
		})
	}
}

func TestGetUserByIDHandler(t *testing.T) {
	tests := []struct {
		name           string
		userIDParam    string
		mockBehavior   func(ctx context.Context, in *pb.GetUserByIDRequest) (*pb.GetUserByIDResponse, error)
		expectedStatus int
	}{
		{
			name:        "Success - Fetch User",
			userIDParam: "usr_786",
			mockBehavior: func(ctx context.Context, in *pb.GetUserByIDRequest) (*pb.GetUserByIDResponse, error) {
				return &pb.GetUserByIDResponse{
					User: &pb.User{
						Id:       in.Id,
						Name:     "Mammad Test",
						Email:    "mammad@test.com",
						Username: "mammad_test",
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Failure - User Not Found (404 from Microservice)",
			userIDParam: "usr_not_exist",
			mockBehavior: func(ctx context.Context, in *pb.GetUserByIDRequest) (*pb.GetUserByIDResponse, error) {
				return nil, status.Error(codes.NotFound, "user not found")
			},
			expectedStatus: http.StatusNotFound, // متد HandleGRPCErr تو این را به ۴۰۴ مپ می‌کند
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGrpcClient := &mockUserServiceClient{getUserFunc: tt.mockBehavior}
			handler := &UserHandler{
				userCLient: &client.UserClient{Client: mockGrpcClient},
				Logger:     zap.NewNop().Sugar(),
			}

			req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/"+tt.userIDParam, nil)

			// 🚀 شبیه‌سازی مکانیزم Path Parameter پکیج chi در محیط تست
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.userIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.GetUserByIDHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %d want %d. Body: %s", rr.Code, tt.expectedStatus, rr.Body.String())
			}
		})
	}
}

func TestDeleteUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		userIDParam    string
		contextUserID  string // آی‌دی که از میدل‌ور فیک توی کانتکست می‌نشیند
		mockBehavior   func(ctx context.Context, in *pb.DeleteUserByIDRequest) (*emptypb.Empty, error)
		expectedStatus int
	}{
		{
			name:          "Success - Authorized Delete",
			userIDParam:   "usr_mammad",
			contextUserID: "usr_mammad", // تطابق کامل برای عبور از چک امنیتی تو
			mockBehavior: func(ctx context.Context, in *pb.DeleteUserByIDRequest) (*emptypb.Empty, error) {
				return &emptypb.Empty{}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Failure - Forbidden (403 ID Mismatch)",
			userIDParam:    "usr_target",
			contextUserID:  "usr_hacker", // عدم تطابق آی‌دی‌ها برای تست بلاک ۴۰۳
			mockBehavior:   nil,          // اصلاً نباید به لایه gRPC برسد
			expectedStatus: http.StatusForbidden,
		},
		{
			name:          "Failure - Internal Server Error from Microservice",
			userIDParam:   "usr_mammad",
			contextUserID: "usr_mammad",
			mockBehavior: func(ctx context.Context, in *pb.DeleteUserByIDRequest) (*emptypb.Empty, error) {
				return nil, status.Error(codes.Internal, "failed to delete from db")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// فرض بر این است که متد مپ شده در ماک اسمش deleteUserFunc است
			mockGrpcClient := &mockUserServiceClient{deleteUserFunc: tt.mockBehavior}
			handler := &UserHandler{
				userCLient: &client.UserClient{Client: mockGrpcClient},
				Logger:     zap.NewNop().Sugar(),
			}

			req, _ := http.NewRequest(http.MethodDelete, "/api/v1/users/"+tt.userIDParam, nil)

			// ۱. تزریق آی‌دی به کانتکست (شبیه‌سازی میدل‌ور احراز هویت تو)
			ctx := context.WithValue(req.Context(), "user_id", tt.contextUserID)

			// ۲. تزریق پارامتر id به مسیر روتینگ chi
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.userIDParam)
			ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.DeleteUserHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %d want %d. Body: %s", rr.Code, tt.expectedStatus, rr.Body.String())
			}
		})
	}
}
