package tests

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"books-and-trust/services/user-service/internal/domain"
	handler "books-and-trust/services/user-service/internal/handler/grpc"
	"books-and-trust/services/user-service/internal/infra/crypto"
	auth "books-and-trust/services/user-service/internal/infra/jwt"
	"books-and-trust/services/user-service/internal/infra/repo"
	"books-and-trust/services/user-service/internal/service"
	pb "books-and-trust/shared/proto/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUserManagement_GRPC_Integration(t *testing.T) {
	ctx := context.Background()

	// 🚀 ۱. ایجاد یک فایل دیتابیس فیزیکی موقت به جای In-Memory
	testDBPath := "user_test.db"
	db, err := gorm.Open(sqlite.Open(testDBPath), &gorm.Config{})
	assert.NoError(t, err)

	// پس از پایان یافتن کل تست‌ها، فایل دیتابیس موقت را پاک کن تا سیستم تمیز بماند
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
		_ = os.Remove(testDBPath)
	}()

	// اجرای Migration
	err = db.AutoMigrate(&domain.User{})
	assert.NoError(t, err)

	// ۲. سیم‌کشی و نمونه‌سازی از لایه‌های واقعی سیستم
	userRepo := repo.NewSQLRepository(db)
	bcryptHasher := crypto.NewBcryptHasher()
	jwtAuth := auth.NewJWTAuthenticator("super_secret_key_1234567890123456", "books-app", "user-service", 15*time.Minute)

	userService := service.NewUserService(userRepo, bcryptHasher, jwtAuth)
	handler := handler.NewGRPCHandler(userService)

	// ۳. راه‌اندازی سرور gRPC روی بافر شبکه در رم (bufconn)
	buffer := 1024 * 1024
	lis := bufconn.Listen(buffer)

	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, handler)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	defer grpcServer.GracefulStop()

	// ۴. ساخت کلاینت gRPC برای شلیک درخواست‌ها
	conn, err := grpc.DialContext(ctx, "passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)

	// ۵. داده‌آمایی: ایجاد یک کاربر نمونه در دیتابیس برای تست گرفتن و حذف کردن
	testUserID := uuid.New()
	plainPassword := "MamadSecure123!"

	testUser := &domain.User{
		ID:       testUserID,
		Name:     "Mohammad Mehdi",
		Username: "mamad_dev",
		Email:    "mamad@example.com",
		Password: domain.Password{
			Text: &plainPassword, // دادن پسورد خام برای اینکه ولیدیشن سرویس پاس شود
		},
	}
	// err = testUser.Password.GenerateHash("MamadSecure123!", bcryptHasher)
	// assert.NoError(t, err)

	err = userService.CreateUser(ctx, testUser)
	assert.NoError(t, err)

	// ==========================================
	// بخش اول: تست‌های یکپارچگی GetUserByID
	// ==========================================
	t.Run("Success_Get_User_By_ID", func(t *testing.T) {
		req := &pb.GetUserByIDRequest{
			Id: testUser.ID.String(),
		}

		resp, err := client.GetUserByID(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, testUser.ID.String(), resp.User.Id)
		assert.Equal(t, testUser.Username, resp.User.Username)
		assert.Equal(t, testUser.Email, resp.User.Email)
	})
	t.Run("Success_Delete_User_By_ID", func(t *testing.T) {
		req := &pb.DeleteUserByIDRequest{
			UserId: testUser.ID.String(),
		}

		resp, err := client.DeleteUserByID(ctx, req)

		// نباید اروری وجود داشته باشد و پاسخ تهی (Empty) gRPC با موفقیت برگردد
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// تایید نهایی: مطمئن می‌شویم کاربر واقعاً از دیتابیس حذف شده و دیگر پیدا نمی‌شود
		checkUser, checkErr := userRepo.GetByID(ctx, testUserID)
		assert.ErrorIs(t, checkErr, domain.ErrResourceNotFound)
		assert.Nil(t, checkUser)
	})

	t.Run("Failed_Delete_User_When_Already_Deleted_Or_Not_Found", func(t *testing.T) {
		// چون در تست قبلی کاربر حذف شد، درخواست مجدد برای همان ID باید ارور NotFound بدهد
		req := &pb.DeleteUserByIDRequest{
			UserId: testUser.ID.String(),
		}

		resp, err := client.DeleteUserByID(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
	})

// ==========================================
	// بخش سوم: تست‌های یکپارچگی UpdateUser
	// ==========================================
	
	// 🎯 یک تابع کمکی کوچک برای تبدیل راحت string به string* در بدنه تست
	ptr := func(s string) *string { return &s }

	t.Run("Success_Update_User_Fields_Via_GRPC", func(t *testing.T) {
		// 🚀 ۱. داده‌آمایی اختصاصی برای تست آپدیت (چون کاربر قبلی حذف شده است)
		updateUserUUID := uuid.New()
		updateUserPassword := "MamadSecure123!"
		
		updateTestUser := &domain.User{
			ID:       updateUserUUID,
			Name:     "Mamad For Update",
			Username: "mamad_update_test",
			Email:    "update_test@example.com",
			Password: domain.Password{
				Text: &updateUserPassword,
			},
		}

		// ساخت کاربر جدید در دیتابیس مخصوص این تست
		err := userService.CreateUser(ctx, updateTestUser)
		assert.NoError(t, err)

		// ۲. آماده‌سازی فیلدهای جدید برای آپدیت
		updatedName := "Mohammad Mehdi (Updated)"
		updatedUsername := "mamad_dev_pro"
		newPass := "MamadNewPass123!"

		req := &pb.UpdateUserRequest{
			UserId:   updateTestUser.ID.String(), // استفاده از ID کاربر جدید و زنده
			Name:     ptr(updatedName),
			Username: ptr(updatedUsername),
			Email:    ptr(updateTestUser.Email),
			Password: ptr(newPass),
		}

		// ۳. شلیک درخواست به سرور gRPC
		resp, err := client.UpdateUser(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, updatedName, resp.User.Name)
		assert.Equal(t, updatedUsername, resp.User.Username)

		// ۴. تأیید نهایی از دیتابیس واقعی
		dbUser, dbErr := userRepo.GetByID(ctx, updateTestUser.ID)
		assert.NoError(t, dbErr)
		assert.Equal(t, updatedName, dbUser.Name)
		assert.Equal(t, updatedUsername, dbUser.Username)
})
	t.Run("Failed_Update_User_With_Invalid_UUID", func(t *testing.T) {
		req := &pb.UpdateUserRequest{
			UserId: "invalid-uuid-string-123",
			Name:   ptr("Mamad"), // 🎯 اصلاح با ptr
		}

		resp, err := client.UpdateUser(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("Failed_Update_User_When_Resource_Not_Found", func(t *testing.T) {
		randomUUID := uuid.New().String()
		req := &pb.UpdateUserRequest{
			UserId:   randomUUID,
			Name:     ptr("Mamad"),          // 🎯 اصلاح با ptr
			Username: ptr("mamad_unknown"),   // 🎯 اصلاح با ptr
			Email:    ptr("unknown@example.com"), // 🎯 اصلاح با ptr
		}

		resp, err := client.UpdateUser(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
	})
}
