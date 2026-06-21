package loanHandler_test

import (
	"books-and-trust/services/api-gateway/internal/client"
	loanHandler "books-and-trust/services/api-gateway/internal/handler/http/loan_handler"
	"books-and-trust/services/api-gateway/internal/middleware"
	pb "books-and-trust/shared/proto/loan"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDeliveryLoanHandler(t *testing.T) {
	tests := []struct {
		name           string
		ctxUserID      any
		urlLoanID      string
		mockSetup      func(m *mockLoanServiceClient)
		expectedStatus int
	}{
		{
			name:      "Success - Loan marked as delivered and returns code",
			ctxUserID: "lender-uuid-123",
			urlLoanID: "loan-uuid-555",
			mockSetup: func(m *mockLoanServiceClient) {
				m.mockDeliveryLoan = func(ctx context.Context, in *pb.DeliveryLoanRequest) (*pb.DeliveryLoanResponse, error) {
					return &pb.DeliveryLoanResponse{
						DeliveryCode: "DEL-123456",
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Failure - Empty URL ID parameter",
			ctxUserID:      "lender-uuid-123",
			urlLoanID:      "",
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Failure - gRPC service returns internal error",
			ctxUserID: "lender-uuid-123",
			urlLoanID: "loan-uuid-555",
			mockSetup: func(m *mockLoanServiceClient) {
				m.mockDeliveryLoan = func(ctx context.Context, in *pb.DeliveryLoanRequest) (*pb.DeliveryLoanResponse, error) {
					return nil, status.Error(codes.Internal, "failed to update status to delivered in database")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:      "Failure - Microservice returns Nil response (Prevents Panic)",
			ctxUserID: "lender-uuid-123",
			urlLoanID: "loan-uuid-555",
			mockSetup: func(m *mockLoanServiceClient) {
				m.mockDeliveryLoan = func(ctx context.Context, in *pb.DeliveryLoanRequest) (*pb.DeliveryLoanResponse, error) {
					return nil, nil // بازگرداندن مقدار نیل برای مچ‌گیری پنیک ❌
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:      "Failure - Microservice returns empty delivery code",
			ctxUserID: "lender-uuid-123",
			urlLoanID: "loan-uuid-555",
			mockSetup: func(m *mockLoanServiceClient) {
				m.mockDeliveryLoan = func(ctx context.Context, in *pb.DeliveryLoanRequest) (*pb.DeliveryLoanResponse, error) {
					return &pb.DeliveryLoanResponse{
						DeliveryCode: "", // کد خالی فرستاده شده ❌
					}, nil
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ۱. ست‌آپ ماک کلاینت gRPC
			mockGrpcClient := &mockLoanServiceClient{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockGrpcClient)
			}

			// ۲. نیو کردن هندلر با کلاینت ماک
			handler := loanHandler.NewLoanHandler(
				zap.NewNop().Sugar(),
				&client.LoanClient{Client: mockGrpcClient},
			)

			// ۳. ساخت ریکوئست POST
			req, err := http.NewRequest(http.MethodPost, "/api/v1/loans/"+tt.urlLoanID+"/delivery", nil)
			if err != nil {
				t.Fatal(err)
			}

			// 🚀 شبیه‌سازی پارامتر مسیر {id} برای روتر chi
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("id", tt.urlLoanID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			// ۴. تزریق کانتکست یوزر آیدی (فرد آپدیت کننده)
			if tt.ctxUserID != nil {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.ctxUserID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			// ۵. اجرای متد هندلر
			handler.DeliveryLoanHandler(rr, req)

			// ۶. تایید کد وضعیت خروجی
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
