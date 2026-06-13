package loanHandler_test

import (
	"books-and-trust/services/api-gateway/internal/client"
	loanHandler "books-and-trust/services/api-gateway/internal/handler/http/loan_handler"
	"books-and-trust/services/api-gateway/internal/middleware"
	pb "books-and-trust/shared/proto/loan"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestBanUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		body           any
		mockSetup      func(m *mockLoanServiceClient)
		expectedStatus int
	}{
		{
			name: "Success - Valid ban user request",
			body: map[string]string{"user_id": "target-user-uuid-999"},
			mockSetup: func(m *mockLoanServiceClient) {
				m.mockBanUser = func(ctx context.Context, in *pb.BanUserRequest) (*emptypb.Empty, error) {
					return &emptypb.Empty{}, nil // میکروسرویس با موفقیت کاربر رو بن میکنه
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Failure - Invalid JSON body",
			body:           "invalid-json-string",
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Failure - Microservice returns internal error",
			body: map[string]string{"user_id": "target-user-uuid-999"},
			mockSetup: func(m *mockLoanServiceClient) {
				m.mockBanUser = func(ctx context.Context, in *pb.BanUserRequest) (*emptypb.Empty, error) {
					return nil, status.Error(codes.Internal, "failed to update user status in db")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGrpcClient := &mockLoanServiceClient{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockGrpcClient)
			}

			handler := loanHandler.NewLoanHandler(
				zap.NewNop().Sugar(),
				&client.LoanClient{Client: mockGrpcClient},
			)

			// تبدیل بادی به جیسون
			var jsonBody []byte
			if str, ok := tt.body.(string); ok {
				jsonBody = []byte(str)
			} else {
				jsonBody, _ = json.Marshal(tt.body)
			}

			req, err := http.NewRequest(http.MethodPost, "/api/v1/admin/users/ban", bytes.NewBuffer(jsonBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.BanUserHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestAdminsMiddleware(t *testing.T) {
	// ساخت یک مپ فیک از ادمین‌ها در حافظه برای تست میدل‌ور
	mockAdminConfig := &middleware.AdminMiddleware{
		Admins: map[string]bool{
			"admin-uuid-111": true, // فقط این آیدی ادمین است
		},
		Logger: zap.NewNop().Sugar(),
	}

	// فرض میکنیم استراکت میدل‌ور هاب شما فیلد admin رو داره
	middlewareHub := middleware.NewGatewayMiddleware(nil , nil , mockAdminConfig , nil , nil , nil , nil)

	// یک هندلر نهایی فیک که اگر کاربر ادمین بود و از میدل‌ور رد شد، وضعیت 200 بده
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		ctxUserID      any // آیدی که توی کانتکست تست تزریق میشه
		expectedStatus int
	}{
		{
			name:           "Success - User is a verified admin in memory map",
			ctxUserID:      "admin-uuid-111",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Failure - User is authenticated but NOT an admin",
			ctxUserID:      "normal-user-uuid-222",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Failure - User ID is missing from context entirely",
			ctxUserID:      nil,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/api/v1/admin/users/ban", nil)
			if err != nil {
				t.Fatal(err)
			}

			// تزریق یوزر آیدی به کانتکست (شبیه‌سازی میدل‌ور Auth قبلی)
			if tt.ctxUserID != nil {
				ctx := context.WithValue(req.Context(), "user_id", tt.ctxUserID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			// اجرای میدل‌ور به همراه هندلر فیک بعدی
			middlewareToTest := middlewareHub.AdminsMiddleware(nextHandler)
			middlewareToTest.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
