package test

import (
	"context"
	"testing"
	"time"

	pb "books-and-trust/shared/proto/loan"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestUserBanFlow_Integration(t *testing.T) {
	cleanDatabase()
	defer cleanDatabase()

	bannedUserID := uuid.New().String()
	expiration := time.Now().Add(time.Hour * 48)

	t.Run("gRPC BanUser_Success_Then_Block_CreateLoan", func(t *testing.T) {
		banReq := &pb.BanUserRequest{
			UserId:    bannedUserID,
			Reason:    "Abuse of service rules",
			ExpiredAt: timestamppb.New(expiration),
		}

		banResp, banErr := loanClient.BanUser(context.Background(), banReq)
		assert.NoError(t, banErr)
		assert.NotNil(t, banResp)

		loanReq := &pb.CreateLoanRequest{
			OwnerId:   bannedUserID, 
			UserId:    uuid.New().String(),
			BookName:  "Designing Data-Intensive Applications",
			Deadline:  timestamppb.New(time.Now().Add(time.Hour * 12)),
		}

		loanResp, loanErr := loanClient.CreateLoan(context.Background(), loanReq)

		assert.Nil(t, loanResp)
		assert.Error(t, loanErr)

		st, ok := status.FromError(loanErr)
		assert.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})
}