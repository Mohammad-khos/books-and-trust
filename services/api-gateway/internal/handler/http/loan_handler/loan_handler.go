package loanHandler

import (
	"books-and-trust/services/api-gateway/internal/client"

	"go.uber.org/zap"
)

type LoanHandler struct {
	Logger     *zap.SugaredLogger
	loanCLient *client.LoanClient
}

func NewLoanHandler(logger *zap.SugaredLogger, client *client.LoanClient) *LoanHandler {
	return &LoanHandler{
		Logger:     logger,
		loanCLient: client,
	}
}
