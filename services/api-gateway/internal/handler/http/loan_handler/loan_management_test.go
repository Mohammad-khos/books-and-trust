package loanHandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"books-and-trust/services/api-gateway/internal/client"
	loanHandler "books-and-trust/services/api-gateway/internal/handler/http/loan_handler"
	pb "books-and-trust/shared/proto/loan"
)

func TestCreateLoanHandler(t *testing.T) {
	fixedTime := time.Now().UTC()

	// ۲. تیبل تست دقیقا با امضا و ساختار تست‌های یوزر شما
	tests := []struct {
		name           string
		ctxUserID      any
		requestBody    any // می‌تونه مپ یا استراکت باشه
		mockBehavior   func(ctx context.Context, in *pb.CreateLoanRequest) (*pb.CreateLoanResponse, error)
		expectedStatus int
	}{
		{
			name:      "Success - Valid Create Loan",
			ctxUserID: "user-uuid-123",
			requestBody: map[string]any{
				"user_id":   "user-uuid-123",
				"book_name": "Clean Architecture",
				"deadline":  fixedTime.Add(24 * time.Hour).Format(time.RFC3339),
			},
			mockBehavior: func(ctx context.Context, in *pb.CreateLoanRequest) (*pb.CreateLoanResponse, error) {
				return &pb.CreateLoanResponse{
					Loan: &pb.Loan{
						Id:        "loan-uuid-999",
						UserId:    in.UserId,
						OwnerId:   "owner-uuid-456",
						Deadline:  timestamppb.New(fixedTime.Add(24 * time.Hour)),
						CreatedAt: timestamppb.New(fixedTime),
					},
				}, nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Failure - Invalid JSON Body",
			ctxUserID:      "user-uuid-123",
			requestBody:    `{invalid-json}`,
			mockBehavior:   nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Failure - Validation Required Fields (Missing BookName)",
			ctxUserID: "user-uuid-123",
			requestBody: map[string]any{
				"user_id": "user-uuid-123",
			},
			mockBehavior:   nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Failure - Context User ID Missing",
			ctxUserID: nil,
			requestBody: map[string]any{
				"user_id":   "user-uuid-123",
				"book_name": "Go Blueprints",
			},
			mockBehavior:   nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:      "Failure - Context User ID Mismatch Forbidden",
			ctxUserID: "wrong-user-uuid",
			requestBody: map[string]any{
				"user_id":   "user-uuid-123",
				"book_name": "Go Blueprints",
			},
			mockBehavior:   nil,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:      "Failure - Internal Server Error from Microservice",
			ctxUserID: "user-uuid-123",
			requestBody: map[string]any{
				"user_id":   "user-uuid-123",
				"book_name": "Go Blueprints",
			},
			mockBehavior: func(ctx context.Context, in *pb.CreateLoanRequest) (*pb.CreateLoanResponse, error) {
				return nil, status.Error(codes.Internal, "database error on loan service")
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:      "Failure - Microservice Returns Nil Loan",
			ctxUserID: "user-uuid-123",
			requestBody: map[string]any{
				"user_id":   "user-uuid-123",
				"book_name": "Go Blueprints",
			},
			mockBehavior: func(ctx context.Context, in *pb.CreateLoanRequest) (*pb.CreateLoanResponse, error) {
				return &pb.CreateLoanResponse{Loan: nil}, nil
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ۳. ست‌آپ کلاینت ماک gRPC دقیقا مثل تست یوزر شما
			mockGrpcClient := &mockLoanServiceClient{mockBehavior: tt.mockBehavior}

			// ۴. دور زدن ارور unexported با استفاده از کانستراکتور NewLoanHandler پروژه
			// 🚀 اگر اسم تابعت فرق داره (مثلا New یا NewHandler)، اون رو جایگزین کن
			handler := loanHandler.NewLoanHandler(
				zap.NewNop().Sugar(),
				&client.LoanClient{Client: mockGrpcClient},
			)

			// ۵. سریالایز کردن بادی ریکوئست
			var bodyBytes []byte
			if strBody, ok := tt.requestBody.(string); ok {
				bodyBytes = []byte(strBody)
			} else {
				bodyBytes, _ = json.Marshal(tt.requestBody)
			}

			req, err := http.NewRequest(http.MethodPost, "/api/v1/loans", bytes.NewBuffer(bodyBytes))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			// ۶. تزریق کانتکست
			if tt.ctxUserID != nil {
				ctx := context.WithValue(req.Context(), "user_id", tt.ctxUserID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			// ۷. اجرای هندلر
			handler.CreateLoanHanlder(rr, req)

			// ۸. تایید کد وضعیت خروجی
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestUpdateLoanHandler(t *testing.T) {

	tests := []struct {
		name           string
		ctxUserID      any
		requestBody    any
		mockSetup      func(m *mockLoanServiceClient)
		expectedStatus int
	}{
		{
			name:      "Success - Owner updates loan within permissions",
			ctxUserID: "borrower-uuid-123", // کاربر داخل کانتکست همان OwnerId است
			requestBody: map[string]any{
				"id": "loan-uuid-777",
			},
			mockSetup: func(m *mockLoanServiceClient) {
				m.mockGetLoanByID = func(ctx context.Context, in *pb.GetLoanByIDRequest) (*pb.GetLoanByIDResponse, error) {
					return &pb.GetLoanByIDResponse{
						Loan: &pb.Loan{
							Id:       "loan-uuid-777",
							UserId:   "lender-uuid-456",   // امانت‌دهنده
							OwnerId:  "borrower-uuid-123", // امانت‌گیرنده (کاربر فعلی)
							BookName: "Domain-Driven Design",
						},
					}, nil
				}
				m.mockUpdateLoan = func(ctx context.Context, in *pb.UpdateLoanRequest) (*emptypb.Empty, error) {
					return &emptypb.Empty{}, nil
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		
		{
			name:      "Failure - Unauthorized user tries to update loan",
			ctxUserID: "hacker-uuid-999", // کاربر جاری هیچکاره است ❌
			requestBody: map[string]any{
				"id": "loan-uuid-777",
			},
			mockSetup: func(m *mockLoanServiceClient) {
				m.mockGetLoanByID = func(ctx context.Context, in *pb.GetLoanByIDRequest) (*pb.GetLoanByIDResponse, error) {
					return &pb.GetLoanByIDResponse{
						Loan: &pb.Loan{
							Id:      "loan-uuid-777",
							UserId:  "lender-uuid-456",
							OwnerId: "borrower-uuid-123",
						},
					}, nil
				}
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Failure - Invalid JSON Body",
			ctxUserID:      "user-uuid-123",
			requestBody:    `{invalid-json}`,
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Failure - Microservice returns empty response",
			ctxUserID: "borrower-uuid-123",
			requestBody: map[string]any{
				"id": "loan-uuid-777",
			},
			mockSetup: func(m *mockLoanServiceClient) {
				m.mockGetLoanByID = func(ctx context.Context, in *pb.GetLoanByIDRequest) (*pb.GetLoanByIDResponse, error) {
					return &pb.GetLoanByIDResponse{Loan: nil}, nil // ریسپانس پوچ ❌
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:      "Failure - gRPC Error on GetLoanByID",
			ctxUserID: "borrower-uuid-123",
			requestBody: map[string]any{
				"id": "loan-uuid-777",
			},
			mockSetup: func(m *mockLoanServiceClient) {
				m.mockGetLoanByID = func(ctx context.Context, in *pb.GetLoanByIDRequest) (*pb.GetLoanByIDResponse, error) {
					return nil, status.Error(codes.NotFound, "loan not found in microservice")
				}
			},
			expectedStatus: http.StatusNotFound, // تبدیل خودکار کدهای gRPC به HTTP در هندلر شما
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ۱. ست‌آپ ماک کلاینت gRPC
			mockGrpcClient := &mockLoanServiceClient{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockGrpcClient)
			}

			// ۲. نیو کردن هندلر با ساختار کلاینت ماک
			handler := loanHandler.NewLoanHandler(
				zap.NewNop().Sugar(),
				&client.LoanClient{Client: mockGrpcClient},
			)

			// ۳. سریالایز بادی ریکوئست
			var bodyBytes []byte
			if strBody, ok := tt.requestBody.(string); ok {
				bodyBytes = []byte(strBody)
			} else {
				bodyBytes, _ = json.Marshal(tt.requestBody)
			}

			req, err := http.NewRequest(http.MethodPut, "/api/v1/loans", bytes.NewBuffer(bodyBytes))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			// ۴. تزریق کانتکست یوزر آیدی (اگر پاس داده شده باشد)
			if tt.ctxUserID != nil {
				ctx := context.WithValue(req.Context(), "user_id", tt.ctxUserID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			// ۵. اجرای هندلر تحت تست
			handler.UpdateLoanHandler(rr, req)

			// ۶. بررسی صحت کد وضعیت HTTP بازگشتی
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestGetLoanByIDHandler(t *testing.T) {
	// fixedTime := time.Now().UTC()

	tests := []struct {
		name           string
		ctxUserID      any
		urlLoanID      string
		mockSetup      func(m *mockLoanServiceClient)
		expectedStatus int
	}{
		{
			name:      "Success - Owner (Lender) fetches loan details",
			ctxUserID: "lender-uuid-456", // کاربر جاری مالک کتاب است
			urlLoanID: "loan-uuid-111",
			mockSetup: func(m *mockLoanServiceClient) {
				m.mockGetLoanByID = func(ctx context.Context, in *pb.GetLoanByIDRequest) (*pb.GetLoanByIDResponse, error) {
					return &pb.GetLoanByIDResponse{
						Loan: &pb.Loan{
							Id:       "loan-uuid-111",
							UserId:   "borrower-uuid-123",
							OwnerId:  "lender-uuid-456", // تطابق با کانتکست
							BookName: "The Go Programming Language",
							Status:   pb.LoanStatus_BORROWED,
						},
					}, nil
				}
			},
			expectedStatus: http.StatusOK, // فیکس شده: تبدیل از 204 به 200 برای دریافت بادی
		},
		{
			name:      "Success - Borrower fetches loan details",
			ctxUserID: "borrower-uuid-123", // کاربر جاری امانت‌گیرنده است
			urlLoanID: "loan-uuid-111",
			mockSetup: func(m *mockLoanServiceClient) {
				m.mockGetLoanByID = func(ctx context.Context, in *pb.GetLoanByIDRequest) (*pb.GetLoanByIDResponse, error) {
					return &pb.GetLoanByIDResponse{
						Loan: &pb.Loan{
							Id:       "loan-uuid-111",
							UserId:   "borrower-uuid-123", // تطابق با کانتکست
							OwnerId:  "lender-uuid-456",
							BookName: "The Go Programming Language",
							Status:   pb.LoanStatus_BORROWED,
						},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Failure - Forbidden user tries to access loan",
			ctxUserID: "hacker-uuid-999", // کاربری که نه مالک است و نه امانت‌گیرنده ❌
			urlLoanID: "loan-uuid-111",
			mockSetup: func(m *mockLoanServiceClient) {
				m.mockGetLoanByID = func(ctx context.Context, in *pb.GetLoanByIDRequest) (*pb.GetLoanByIDResponse, error) {
					return &pb.GetLoanByIDResponse{
						Loan: &pb.Loan{
							Id:      "loan-uuid-111",
							UserId:  "borrower-uuid-123",
							OwnerId: "lender-uuid-456",
						},
					}, nil
				}
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Failure - Empty URL ID parameter",
			ctxUserID:      "borrower-uuid-123",
			urlLoanID:      "", // پارامتر آیدی خالی فرستاده شده ❌
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Failure - Microservice returns Nil response (Prevents Panic)",
			ctxUserID: "borrower-uuid-123",
			urlLoanID: "loan-uuid-111",
			mockSetup: func(m *mockLoanServiceClient) {
				m.mockGetLoanByID = func(ctx context.Context, in *pb.GetLoanByIDRequest) (*pb.GetLoanByIDResponse, error) {
					return nil, nil // ریسپانس کاملاً نیل برای تست مچ‌گیری پنیک ❌
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:      "Failure - gRPC Service returns NotFound Error",
			ctxUserID: "borrower-uuid-123",
			urlLoanID: "loan-uuid-111",
			mockSetup: func(m *mockLoanServiceClient) {
				m.mockGetLoanByID = func(ctx context.Context, in *pb.GetLoanByIDRequest) (*pb.GetLoanByIDResponse, error) {
					return nil, status.Error(codes.NotFound, "loan not found")
				}
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ۱. ساخت کلاینت ماک بر اساس استراکت ارتقایافته شما
			mockGrpcClient := &mockLoanServiceClient{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockGrpcClient)
			}

			// ۲. نیو کردن هندلر لایه گیت‌وی
			handler := loanHandler.NewLoanHandler(
				zap.NewNop().Sugar(),
				&client.LoanClient{Client: mockGrpcClient},
			)

			// ۳. شبیه‌سازی ریکوئست GET
			req, err := http.NewRequest(http.MethodGet, "/api/v1/loans/"+tt.urlLoanID, nil)
			if err != nil {
				t.Fatal(err)
			}

			// 🚀 شبیه‌سازی رفتار روتر chi برای خواندن URL Param
			// ابزار chi پارامترها را داخل کانتکست درخواست ذخیره می‌کند
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("id", tt.urlLoanID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			// ۴. تزریق کانتکست یوزر آیدی
			if tt.ctxUserID != nil {
				ctx := context.WithValue(req.Context(), "user_id", tt.ctxUserID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			// ۵. اجرای متد هندلر
			handler.GetLoanByIDHandler(rr, req)

			// ۶. بررسی پاس شدن وضعیت خروجی
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestListLoanByOwner(t *testing.T) {
	fixedTime := time.Now().UTC()

	tests := []struct {
		name           string
		ctxUserID      any
		urlOwnerID     string
		mockSetup      func(m *mockLoanServiceClient)
		expectedStatus int
	}{
		{
    name:       "Success - Owner fetches their own loan list",
    ctxUserID:  "owner-uuid-123", 
    urlOwnerID: "owner-uuid-123", 
    mockSetup: func(m *mockLoanServiceClient) {
        // 🚀 فیکس اصلی: دقیقاً روی شیء m مقداردهی کن، نه روی m.pb.LoanServiceClient
        m.mockListLoansByOwner = func(ctx context.Context, in *pb.ListLoansByOwnerRequest) (*pb.ListLoansByOwnerResponse, error) {
            return &pb.ListLoansByOwnerResponse{
                Loans: []*pb.Loan{
                    {
                        Id:        "loan-1",
                        UserId:    "borrower-1",
                        OwnerId:   "owner-uuid-123",
                        BookName:  "Clean Code",
                        Status:    pb.LoanStatus_BORROWED,
                        Deadline:  timestamppb.New(fixedTime.Add(48 * time.Hour)),
                        CreatedAt: timestamppb.New(fixedTime),
                    },
                    {
                        Id:        "loan-2",
                        UserId:    "borrower-2",
                        OwnerId:   "owner-uuid-123",
                        BookName:  "Refactoring",
                        Status:    pb.LoanStatus_RETURNED,
                        Deadline:  timestamppb.New(fixedTime.Add(96 * time.Hour)),
                        CreatedAt: timestamppb.New(fixedTime),
                    },
                },
            }, nil
        }
    },
    expectedStatus: http.StatusOK,
},
		{
			name:           "Failure - Unauthorized to fetch another user's loans",
			ctxUserID:      "owner-uuid-123",  // کاربر جاری
			urlOwnerID:     "someone-else-id", // تلاش برای دیدن لیست یک نفر دیگه ❌
			mockSetup:      nil,               // اصلاً به لایه gRPC نمی‌رسه
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Failure - Empty ownerID URL parameter",
			ctxUserID:      "owner-uuid-123",
			urlOwnerID:     "", // پارامتر خالی ❌
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "Failure - gRPC service internal error",
			ctxUserID:  "owner-uuid-123",
			urlOwnerID: "owner-uuid-123",
			mockSetup: func(m *mockLoanServiceClient) {
				m.LoanServiceClient = &mockLoanServiceClient{
					mockListLoansByOwner: func(ctx context.Context, in *pb.ListLoansByOwnerRequest) (*pb.ListLoansByOwnerResponse, error) {
						return nil, status.Error(codes.Internal, "database connection failed downstream")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:       "Failure - Empty response from loan microservice",
			ctxUserID:  "owner-uuid-123",
			urlOwnerID: "owner-uuid-123",
			mockSetup: func(m *mockLoanServiceClient) {
				m.LoanServiceClient = &mockLoanServiceClient{
					mockListLoansByOwner: func(ctx context.Context, in *pb.ListLoansByOwnerRequest) (*pb.ListLoansByOwnerResponse, error) {
						return nil, nil // ریسپانس نیل برای جلوگیری از کرش سیستم ❌
					},
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

			// ۳. ساخت ریکوئست گت
			req, err := http.NewRequest(http.MethodGet, "/api/v1/loans/owner/"+tt.urlOwnerID, nil)
			if err != nil {
				t.Fatal(err)
			}

			// 🚀 شبیه‌سازی پارامتر مسیر {ownerID} برای روتر chi
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("ownerID", tt.urlOwnerID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			// ۴. تزریق کانتکست یوزر آیدی
			if tt.ctxUserID != nil {
				ctx := context.WithValue(req.Context(), "user_id", tt.ctxUserID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			// ۵. اجرای هندلر
			handler.ListLoanByOwner(rr, req)

			// ۶. تایید کد وضعیت خروجی
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}