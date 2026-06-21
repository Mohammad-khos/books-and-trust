package test

import (
	"context"
	"testing"
	"time"

	pb "books-and-trust/shared/proto/loan"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestLoanFlow_Integration(t *testing.T) {
	cleanDatabase()
	defer cleanDatabase()

	ownerID := uuid.New().String()
	userID := uuid.New().String()
	deadline := time.Now().Add(time.Hour * 24)

	t.Run("gRPC CreateLoan_Success_To_Database", func(t *testing.T) {
		req := &pb.CreateLoanRequest{
			OwnerId:  ownerID,
			UserId:   userID,
			BookName: "Software Architecture patterns",
			Deadline: timestamppb.New(deadline),
		}

		resp, err := loanClient.CreateLoan(context.Background(), req)

		
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		assert.NotEmpty(t, resp.Loan.Id)
		assert.Equal(t, pb.LoanStatus_LOAN_STATUS_UNSPECIFIED, resp.Loan.Status)
	})
}
