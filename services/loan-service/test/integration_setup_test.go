package test

import (
	"context"
	"log"
	"net"
	"os"
	"testing"

	"books-and-trust/services/loan-service/internal/domain"
	grpc_handler "books-and-trust/services/loan-service/internal/handler/grpc"
	"books-and-trust/services/loan-service/internal/infra/repo"
	"books-and-trust/services/loan-service/internal/service"
	pb "books-and-trust/shared/proto/loan"

	"github.com/testcontainers/testcontainers-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"gorm.io/gorm"
)

const bufSize = 1024 * 1024

var (
	lis        *bufconn.Listener
	testDB     *gorm.DB
	loanClient pb.LoanServiceClient
	container  testcontainers.Container
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	var err error
	testDB, container, err = setupPostgresContainer(ctx)
	if err != nil {
		log.Fatalf("❌ failed to start Postgres container: %v", err)
	}

	err = testDB.AutoMigrate(&domain.Loan{}, &domain.BannedUser{})
	if err != nil {
		_ = container.Terminate(ctx)
		log.Fatalf("❌ failed to migrate Postgres database: %v", err)
	}

	lis = bufconn.Listen(bufSize)
	baseServer := grpc.NewServer()

	repo := repo.NewPostgresRepository(testDB)
	svc := service.NewLoanService(repo)
	handler := grpc_handler.NewGRPCHandler(svc)

	pb.RegisterLoanServiceServer(baseServer, handler)

	go func() {
		if err := baseServer.Serve(lis); err != nil {
			log.Fatalf("❌ Server exited with error: %v", err)
		}
	}()

	conn, err := grpc.NewClient( "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		_ = container.Terminate(ctx)
		log.Fatalf("❌ Failed to dial bufnet: %v", err)
	}

	loanClient = pb.NewLoanServiceClient(conn)

	code := m.Run()

	baseServer.GracefulStop()
	_ = conn.Close()
	_ = container.Terminate(ctx)
	os.Exit(code)
}

func cleanDatabase() {
	testDB.Exec("DELETE FROM loans;")
	testDB.Exec("DELETE FROM banned_users;")
}
