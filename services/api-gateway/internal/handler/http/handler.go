package handler

import (
	loanHandler "books-and-trust/services/api-gateway/internal/handler/http/loan_handler"
	userHandler "books-and-trust/services/api-gateway/internal/handler/http/user_handler"

	"go.uber.org/zap"
)

type HTTPHandler struct {
	Logger      *zap.SugaredLogger
	UserHandler *userHandler.UserHandler
	LoanHandler *loanHandler.LoanHandler
}

func NewHTTPHandler(logger *zap.SugaredLogger, userHandler *userHandler.UserHandler, loanHandler *loanHandler.LoanHandler) *HTTPHandler {
	return &HTTPHandler{
		Logger:      logger,
		UserHandler: userHandler,
		LoanHandler: loanHandler,
	}
}
