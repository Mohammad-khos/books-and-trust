package tests

import (
	"context"
	"net"
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
)

func TestUserManagement_GRPC_Integration(t *testing.T) {
	ctx := context.Background()

    db, container, err := setupPostgresContainer(ctx)
    assert.NoError(t, err)
    defer func() { _ = container.Terminate(ctx) }()

    err = db.AutoMigrate(&domain.User{})
    assert.NoError(t, err)
    if err != nil {
        return
    }

	userRepo := repo.NewSQLRepository(db)
	bcryptHasher := crypto.NewBcryptHasher()
	jwtAuth := auth.NewJWTAuthenticator("super_secret_key_1234567890123456", "books-app", "user-service", 15*time.Minute)

	userService := service.NewUserService(userRepo, bcryptHasher, jwtAuth)
	handler := handler.NewGRPCHandler(userService)

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

	conn, err := grpc.NewClient( "passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	assert.NoError(t, err)
	defer func ()  {
		if err := conn.Close(); err != nil {
			t.Errorf("Failed to close gRPC connection %v" , err)
		}
	}()

	client := pb.NewUserServiceClient(conn)

	testUserID := uuid.New()
	plainPassword := "MamadSecure123!"

	testUser := &domain.User{
		ID:       testUserID,
		Name:     "Mohammad Mehdi",
		Username: "mamad_dev",
		Email:    "mamad@example.com",
		Password: domain.Password{
			Text: &plainPassword,
		},
	}

	err = userService.CreateUser(ctx, testUser)
	assert.NoError(t, err)

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

		assert.NoError(t, err)
		assert.NotNil(t, resp)

		checkUser, checkErr := userRepo.GetByID(ctx, testUserID)
		assert.ErrorIs(t, checkErr, domain.ErrResourceNotFound)
		assert.Nil(t, checkUser)
	})

	t.Run("Failed_Delete_User_When_Already_Deleted_Or_Not_Found", func(t *testing.T) {
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


	ptr := func(s string) *string { return &s }

	t.Run("Success_Update_User_Fields_Via_GRPC", func(t *testing.T) {
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

		err := userService.CreateUser(ctx, updateTestUser)
		assert.NoError(t, err)

		updatedName := "Mohammad Mehdi (Updated)"
		updatedUsername := "mamad_dev_pro"
		newPass := "MamadNewPass123!"

		req := &pb.UpdateUserRequest{
			UserId:   updateTestUser.ID.String(),
			Name:     ptr(updatedName),
			Username: ptr(updatedUsername),
			Email:    ptr(updateTestUser.Email),
			Password: ptr(newPass),
		}

		resp, err := client.UpdateUser(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, updatedName, resp.User.Name)
		assert.Equal(t, updatedUsername, resp.User.Username)

		dbUser, dbErr := userRepo.GetByID(ctx, updateTestUser.ID)
		assert.NoError(t, dbErr)
		assert.Equal(t, updatedName, dbUser.Name)
		assert.Equal(t, updatedUsername, dbUser.Username)
	})
	t.Run("Failed_Update_User_With_Invalid_UUID", func(t *testing.T) {
		req := &pb.UpdateUserRequest{
			UserId: "invalid-uuid-string-123",
			Name:   ptr("Mamad"),
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
			Name:     ptr("Mamad"),
			Username: ptr("mamad_unknown"),
			Email:    ptr("unknown@example.com"),
		}

		resp, err := client.UpdateUser(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
	})
}
