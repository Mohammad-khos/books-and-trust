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
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestLoginUser_GRPC_Integration(t *testing.T) {
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

	jwtAuth := auth.NewJWTAuthenticator("my_super_secret_key_1234567890123456", "books-app", "user-service", 15*time.Minute)

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

	conn, err := grpc.NewClient("bufnet",
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

	plainPassword := "MamadSecure123!"
	testUser := &domain.User{
		ID:       uuid.New(),
		Name:     "Mohammad",
		Username: "mamad_dev",
		Email:    "mamad@example.com",
	}
	err = testUser.Password.GenerateHash(plainPassword, bcryptHasher)
	assert.NoError(t, err)
	err = userRepo.Create(ctx, testUser)
	assert.NoError(t, err)


	t.Run("Success_Integration_Login", func(t *testing.T) {
		req := &pb.LoginUserRequest{
			UsernameOrEmail: "mamad_dev",
			Password:        plainPassword,
		}

		resp, err := client.LoginUser(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.AccessToken)
		assert.Equal(t, "Bearer", resp.TokenType)
		assert.Equal(t, testUser.Username, resp.User.Username)
		assert.Equal(t, testUser.Email, resp.User.Email)
	})

	t.Run("Failed_Integration_Login_Wrong_Password", func(t *testing.T) {
		req := &pb.LoginUserRequest{
			UsernameOrEmail: "mamad_dev",
			Password:        "WrongPassword!!!",
		}

		resp, err := client.LoginUser(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)

		assert.Contains(t, err.Error(), "invalid username or password")
	})
}
