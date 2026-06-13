package service

import (
	"books-and-trust/services/loan-service/internal/domain"
)

type LoanService struct {
	repo domain.LoanRepository
}

func NewLoanService(repo domain.LoanRepository) *LoanService {
	return &LoanService{
		repo: repo,
	}
}
