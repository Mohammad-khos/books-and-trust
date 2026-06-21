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

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"gorm.io/driver/sqlite" 
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const bufSize = 1024 * 1024

var (
	lis        *bufconn.Listener
	testDB     *gorm.DB
	loanClient pb.LoanServiceClient
)

func TestMain(m *testing.M) {
	var err error
	
	testDB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("❌ failed to connect to in-memory SQLite: %v", err)
	}

	err = testDB.AutoMigrate(&domain.Loan{}, &domain.BannedUser{})
	if err != nil {
		log.Fatalf("❌ failed to migrate SQLite database: %v", err)
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

	conn, err := grpc.NewClient("bufnet", 
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), 
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("❌ Failed to dial bufnet: %v", err)
	}
	defer func ()  {
		if err := conn.Close(); err != nil {
			log.Fatalf("Failed to close gRPC connection %v" , err)
		}
	}()

	loanClient = pb.NewLoanServiceClient(conn)

	code := m.Run()

	baseServer.GracefulStop()
	os.Exit(code)
}

func cleanDatabase() {
	testDB.Exec("DELETE FROM loans;")
	testDB.Exec("DELETE FROM banned_users;")
	testDB.Exec("DELETE FROM sqlite_sequence WHERE name IN ('loans', 'banned_users');")
}