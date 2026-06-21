package tests

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"books-and-trust/services/user-service/internal/domain"
	h "books-and-trust/services/user-service/internal/handler/grpc"
	"books-and-trust/services/user-service/internal/infra/crypto"
	auth "books-and-trust/services/user-service/internal/infra/jwt"
	"books-and-trust/services/user-service/internal/infra/repo"
	"books-and-trust/services/user-service/internal/service"
	pb "books-and-trust/shared/proto/user"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func TestRegisterUser_GRPC_Integration(t *testing.T) {
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
	jwt := auth.NewJWTAuthenticator("wQeuJsdksMcnOWnkdwe", "test", "test", time.Hour)

	userService := service.NewUserService(userRepo, bcryptHasher, jwt)
	grpcHandler := h.NewGRPCHandler(userService)

	lis := bufconn.Listen(bufSize)
	baseServer := grpc.NewServer()
	pb.RegisterUserServiceServer(baseServer, grpcHandler)

	go func() {
		if err := baseServer.Serve(lis); err != nil {
			log.Printf("Server exited with error: %v", err)
		}
	}()
	defer baseServer.Stop()

	dialCtx := context.Background()
	conn, err := grpc.DialContext(dialCtx, "bufnet",
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

	req := &pb.RegisterUserRequest{
		Name:     "Mamad GRPC",
		Username: "grpc_test_user",
		Email:    "grpc@test.com",
		Password: "SecurePassword123!",
	}

	resp, err := client.RegisterUser(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Mamad GRPC", resp.User.GetName())

	var savedUser domain.User
	err = db.Where("username = ?", "grpc_test_user").First(&savedUser).Error
	assert.NoError(t, err)
	assert.Equal(t, "grpc@test.com", savedUser.Email)

	db.Exec("DELETE FROM users WHERE username = ?", "grpc_test_user")
}
