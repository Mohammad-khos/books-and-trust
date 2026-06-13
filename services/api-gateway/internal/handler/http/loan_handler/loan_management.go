package loanHandler

import (
	"books-and-trust/services/api-gateway/internal/middleware"
	httputil "books-and-trust/services/api-gateway/util"
	pb "books-and-trust/shared/proto/loan"
	"books-and-trust/shared/validation"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sony/gobreaker"
)

// CreateLoanHandler godoc
//
//	@Summary		Create a new loan
//	@Description	Creates a new loan record by calling the internal loan microservice. Only the user themselves can initiate the loan.
//	@Tags			loans
//	@Accept			json
//	@Produce		json
//	@Param			request	body	CreateLoanRequest	true	"Create Loan Request Body"
//	@Security		BearerAuth
//	@Success		201	{object}	contracts.ResponseSuccess{data_message=CreateLoanResponse}	"loan successfully created"
//	@Failure		400	{object}	contracts.ResponseError										"invalid create loan request"
//	@Failure		403	{object}	contracts.ResponseError										"you can not create loan"
//	@Failure		500	{object}	contracts.ResponseError										"internal server error"
//	@Router			/loans [post]
func (h *LoanHandler) CreateLoanHanlder(w http.ResponseWriter, r *http.Request) {
	logger := httputil.LoggerWithTrace(r.Context() , h.Logger)
	var req CreateLoanRequest
	if err := httputil.ReadJSON(w, r, &req); err != nil {
		httputil.BadRequestErr(w, r, logger, "invalid create loan request")
		return
	}
	if err := validation.Validator.Struct(req); err != nil {
		httputil.BadRequestErr(w, r, logger, "invalid create loan request")
		return
	}
	id := h.getUserIDFromContext(w, r)
	if id == "" {
        return 
    }
	if id != req.OwnerID {
		httputil.ForbiddenErr(w, r, logger, "you can not create loan")
		return
	}
	pbResp, err := h.loanCLient.Client.CreateLoan(r.Context(), req.ToProto())
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			httputil.ServiceUnavailableErr(w, r, logger, "")
			return
		}
		httputil.HandleGRPCErr(w, r, logger, err)
		return
	}
	if pbResp == nil || pbResp.GetLoan() == nil {
		httputil.InternalServerErr(w, r, logger, errors.New("empty response from loan service"))
		return
	}
	var deadline, createdAt time.Time
	loan := pbResp.GetLoan()
	if loan.GetDeadline() != nil {
		deadline = loan.GetDeadline().AsTime()
	}
	if loan.GetCreatedAt() != nil {
		createdAt = loan.GetCreatedAt().AsTime()
	}
	resp := CreateLoanResponse{
		ID:        loan.GetId(),
		UserID:    loan.GetUserId(),
		OwnerID:   loan.GetOwnerId(),
		BookName:  loan.GetBookName(),
		Deadline:  deadline,
		CreatedAt: createdAt,
	}
	if err := httputil.WriteJSON(w, resp, http.StatusCreated, "loan successfully created"); err != nil {
		httputil.InternalServerErr(w, r, logger, errors.New("failed to send create loan response"))
		return
	}
}

// UpdateLoanHandler godoc
//
//	@Summary		Update an existing loan
//	@Description	Updates a loan. Only the Lender (UserID) or Borrower (OwnerID) can update it. Note: The borrower cannot extend the deadline.
//	@Tags			loans
//	@Accept			json
//	@Produce		json
//	@Param			request	body	UpdateLoanRequest	true	"Update Loan Request Body"
//	@Security		BearerAuth
//	@Success		204	{object}	contracts.ResponseSuccess	"loan updated successfully"
//	@Failure		400	{object}	contracts.ResponseError		"invalid update loan request"
//	@Failure		403	{object}	contracts.ResponseError		"you can not update this loan / borrower cannot extend the deadline"
//	@Failure		404	{object}	contracts.ResponseError		"loan record not found"
//	@Failure		500	{object}	contracts.ResponseError		"internal server error"
//	@Router			/loans [patch]
func (h *LoanHandler) UpdateLoanHandler(w http.ResponseWriter, r *http.Request) {
	logger := httputil.LoggerWithTrace(r.Context() , h.Logger)
	var req UpdateLoanRequest
	if err := httputil.ReadJSON(w, r, &req); err != nil {
		httputil.BadRequestErr(w, r, logger, "invalid update loan request")
		return
	}
	if err := validation.Validator.Struct(req); err != nil {
		httputil.BadRequestErr(w, r, logger, "invalid update loan request")
		return
	}
	userID := h.getUserIDFromContext(w, r)
	pbLoan, err := h.loanCLient.Client.GetLoanByID(r.Context(), &pb.GetLoanByIDRequest{
		LoanId: req.ID,
	})
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			httputil.ServiceUnavailableErr(w, r, logger, "")
			return
		}
		httputil.HandleGRPCErr(w, r, logger, err)
		return
	}
	if pbLoan == nil || pbLoan.GetLoan() == nil {
		httputil.InternalServerErr(w, r, logger, errors.New("empty response from loan service"))
		return
	}
	loan := pbLoan.GetLoan()
	if loan.GetOwnerId() != userID {
		httputil.ForbiddenErr(w, r, logger, "you can not update this loan")
		return
	}
	_, err = h.loanCLient.Client.UpdateLoan(r.Context(), req.ToProto(userID))
	if err != nil {
		httputil.HandleGRPCErr(w, r, logger, err)
		return
	}
	if err := httputil.WriteJSON(w, nil, http.StatusNoContent, "loan updated successfully"); err != nil {
		httputil.InternalServerErr(w, r, logger, err)
		return
	}

}

// GetLoanByIDHandler godoc
//
//	@Summary		Get loan details by ID
//	@Description	Fetches the details of a specific loan using its UUID from the URL parameter. Only the Lender (UserID) or Borrower (OwnerID) involved in the loan can access it.
//	@Tags			loans
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"Loan UUID"
//	@Security		BearerAuth
//	@Success		200	{object}	contracts.ResponseSuccess{data_message=GetLoanByIDResponse}	"loan fetched successfully"
//	@Failure		400	{object}	contracts.ResponseError										"invalid url param"
//	@Failure		403	{object}	contracts.ResponseError										"you can not get loan"
//	@Failure		404	{object}	contracts.ResponseError										"loan record not found"
//	@Failure		500	{object}	contracts.ResponseError										"internal server error"
//	@Router			/loans/{id} [get]
func (h *LoanHandler) GetLoanByIDHandler(w http.ResponseWriter, r *http.Request) {
	logger := httputil.LoggerWithTrace(r.Context() , h.Logger)
	loanID := chi.URLParam(r, "id")
	if loanID == "" {
		httputil.BadRequestErr(w, r, logger, "invalid url param")
		return
	}
	userID := h.getUserIDFromContext(w, r)
	if userID == "" {
        return 
    }
	pbResp, err := h.loanCLient.Client.GetLoanByID(r.Context(), &pb.GetLoanByIDRequest{
		LoanId: loanID,
	})
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			httputil.ServiceUnavailableErr(w, r, logger, "")
			return
		}
		httputil.HandleGRPCErr(w, r, logger, err)
		return
	}
	if pbResp == nil || pbResp.GetLoan() == nil {
		httputil.InternalServerErr(w, r, logger, errors.New("empty response from loan service"))
		return
	}
	loan := pbResp.GetLoan()
	if loan.OwnerId != userID && loan.UserId != userID {
		httputil.ForbiddenErr(w, r, logger, "you cannot access this loan")
		return
	}
	var deadline, createdAt time.Time
	if loan.GetDeadline() != nil {
		deadline = loan.GetDeadline().AsTime()
	}
	if loan.GetCreatedAt() != nil {
		createdAt = loan.GetCreatedAt().AsTime()
	}
	resp := GetLoanByIDResponse{
		Loan: Loan{
			ID:        loan.GetId(),
			UserID:    loan.GetUserId(),
			OwnerID:   loan.GetOwnerId(),
			BookName:  loan.GetBookName(),
			Status:    loan.GetStatus().String(),
			Deadline:  deadline,
			CreatedAt: createdAt,
		},
	}
	if err := httputil.WriteJSON(w, resp, http.StatusOK, "loan fetched successfully"); err != nil {
		httputil.InternalServerErr(w, r, logger, err)
		return
	}
}

// ListLoanByOwner godoc
//
//	@Summary		List loans by owner ID
//	@Description	Retrieves a list of all loans shared or owned by a specific user. Users can only fetch their own loan lists (ownerID must match the authenticated user ID).
//	@Tags			loans
//	@Accept			json
//	@Produce		json
//	@Param			ownerID	path	string	true	"Owner User UUID"
//	@Security		BearerAuth
//	@Success		200	{object}	contracts.ResponseSuccess{data_message=ListLoanByOwnerResponse}	"loans fetched successfully"
//	@Failure		400	{object}	contracts.ResponseError											"invalid url param"
//	@Failure		403	{object}	contracts.ResponseError											"you cannot access these loans"
//	@Failure		500	{object}	contracts.ResponseError											"internal server error"
//	@Router			/loans/owner/{ownerID} [get]
func (h *LoanHandler) ListLoanByOwner(w http.ResponseWriter, r *http.Request) {
	logger := httputil.LoggerWithTrace(r.Context() , h.Logger)
	ownerID := chi.URLParam(r, "ownerID")
	if ownerID == "" {
		httputil.BadRequestErr(w, r, logger, "invalid url param")
		return
	}
	userID := h.getUserIDFromContext(w, r)
	if userID == "" {
        return 
    }
	if ownerID != userID {
		httputil.ForbiddenErr(w, r, logger, "you cannot access these loans")
		return
	}
	pbResp, err := h.loanCLient.Client.ListLoansByOwner(r.Context(), &pb.ListLoansByOwnerRequest{
		OwnerId: ownerID,
	})
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			httputil.ServiceUnavailableErr(w, r, logger, "")
			return
		}
		httputil.HandleGRPCErr(w, r, logger, err)
		return
	}
	if pbResp == nil || pbResp.GetLoans() == nil {
		httputil.InternalServerErr(w, r, logger, errors.New("empty response from loan service"))
		return
	}
	pbLoans := pbResp.GetLoans()
	resp := make(ListLoanByOwnerResponse, 0, len(pbLoans))
	for _, pbLoan := range pbLoans {
		var deadline, createdAt time.Time
		if pbLoan.GetDeadline() != nil {
			deadline = pbLoan.GetDeadline().AsTime()
		}
		if pbLoan.GetCreatedAt() != nil {
			createdAt = pbLoan.GetCreatedAt().AsTime()
		}
		loanData := Loan{
			ID:        pbLoan.GetId(),
			UserID:    pbLoan.GetUserId(),
			OwnerID:   pbLoan.GetOwnerId(),
			BookName:  pbLoan.GetBookName(),
			Status:    pbLoan.GetStatus().String(),
			Deadline:  deadline,
			CreatedAt: createdAt,
		}
		resp = append(resp, loanData)
	}

	if err := httputil.WriteJSON(w, resp, http.StatusOK, "loans fetched successfully"); err != nil {
		httputil.InternalServerErr(w, r, logger, err)
		return
	}
}

// ClaimLoanHandler godoc
//
//	@Summary		Claim an available loan position
//	@Description	Allows an authenticated user to claim/borrow an available book loan by its ID. The user cannot be the owner of the loan.
//	@Tags			loans
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"Loan UUID"
//	@Security		BearerAuth
//	@Success		200	{object}	contracts.ResponseSuccess	"user banned successfully"
//	@Failure		400	{object}	contracts.ResponseError		"invalid url param"
//	@Failure		401	{object}	contracts.ResponseError		"unauthorized"
//	@Failure		403	{object}	contracts.ResponseError		"forbidden - cannot claim own loan or user is banned"
//	@Failure		404	{object}	contracts.ResponseError		"loan record not found"
//	@Failure		409	{object}	contracts.ResponseError		"conflict - loan already claimed by another user"
//	@Failure		500	{object}	contracts.ResponseError		"internal server error"
//	@Router			/loans/{id}/claim [post]
func (h *LoanHandler) ClaimLoanHandler(w http.ResponseWriter, r *http.Request) {
	logger := httputil.LoggerWithTrace(r.Context() , h.Logger)
	updaterID := h.getUserIDFromContext(w, r)
	if updaterID == "" {
        return 
    }
	loanID := chi.URLParam(r, "id")
	if loanID == "" {
		httputil.BadRequestErr(w, r, logger, "invalid url param: loan id is required")
		return
	}

	_, err := h.loanCLient.Client.ClaimLoan(r.Context(), &pb.ClaimLoanRequest{
		LoanId: loanID,
		UserId: updaterID,
	})

	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			httputil.ServiceUnavailableErr(w, r, logger, "")
			return
		}
		httputil.HandleGRPCErr(w, r, logger, err)
		return
	}

	if err := httputil.WriteJSON(w, nil, http.StatusOK, "loan claimed successfully"); err != nil {
		httputil.InternalServerErr(w, r, logger, errors.New("failed to write claim loan response"))
		return
	}
}

func (h *LoanHandler) getUserIDFromContext(w http.ResponseWriter, r *http.Request) string {
	logger := httputil.LoggerWithTrace(r.Context() , h.Logger)
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		httputil.UnauthorizedErr(w, r, logger, "unauthorized access")
		return ""
	}
	return userID
}
