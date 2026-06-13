package loanHandler

import (
	httputil "books-and-trust/services/api-gateway/util"
	pb "books-and-trust/shared/proto/loan"
	"books-and-trust/shared/validation"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// DeliveryLoanHandler godoc
//
//	@Summary		Mark a loan as delivered
//	@Description	Updates the status of a loan to indicate that the book has been physically delivered. Returns a unique delivery confirmation code.
//	@Tags			loans
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"Loan UUID"
//	@Security		BearerAuth
//	@Success		200	{object}	contracts.ResponseSuccess{data_message=DeliveryLoanResponse}	"loan delivered successfully"
//	@Failure		400	{object}	contracts.ResponseError											"invalid url param"
//	@Failure		401	{object}	contracts.ResponseError											"unauthorized"
//	@Failure		403	{object}	contracts.ResponseError											"forbidden - you do not have permission to update this loan"
//	@Failure		404	{object}	contracts.ResponseError											"loan not found"
//	@Failure		500	{object}	contracts.ResponseError											"internal server error"
//	@Router			/loans/{id}/delivery [post]
func (h *LoanHandler) DeliveryLoanHandler(w http.ResponseWriter, r *http.Request) {
	logger := httputil.LoggerWithTrace(r.Context(), h.Logger)
	updaterID := h.getUserIDFromContext(w, r)
	if updaterID == "" {
        return 
    }
	loanID := chi.URLParam(r, "id")
	if loanID == "" {
		httputil.BadRequestErr(w, r, logger, "invalid url param")
		return
	}
	pbResp, err := h.loanCLient.Client.DeliveryLoan(r.Context(), &pb.DeliveryLoanRequest{
		LoanID:    loanID,
		UpdaterId: updaterID,
	})
	if err != nil {
		httputil.HandleGRPCErr(w, r, logger, err)
		return
	}
	if pbResp == nil || pbResp.GetDeliveryCode() == "" {
		httputil.InternalServerErr(w, r, logger, errors.New("empty response from loan service"))
		return
	}
	resp := DeliveryLoanResponse{
		DeliveryCode: pbResp.GetDeliveryCode(),
	}
	if err := httputil.WriteJSON(w, resp, http.StatusOK, "loan delivered successfully"); err != nil {
		httputil.InternalServerErr(w, r, logger, errors.New("failed to create delivery loan response"))
		return
	}
}

// ConfirmDeliveryHandler godoc
//
//	@Summary		Confirm loan delivery using OTP code
//	@Description	Allows the loan owner to input the 6-digit verification code received from the borrower to finalize and complete the loan.
//	@Tags			loans
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string					true	"Loan UUID"
//	@Param			body	body	ConfirmDeliveryRequest	true	"Delivery Code Body"
//	@Security		BearerAuth
//	@Success		200	{object}	contracts.ResponseSuccess	"loan delivery confirmed successfully"
//	@Failure		400	{object}	contracts.ResponseError		"invalid loan id or missing code"
//	@Failure		401	{object}	contracts.ResponseError		"unauthorized"
//	@Failure		404	{object}	contracts.ResponseError		"loan not found"
//	@Failure		500	{object}	contracts.ResponseError		"internal server error"
//	@Router			/loans/{id}/confirm-delivery [post]
func (h *LoanHandler) ConfirmDeliveryHandler(w http.ResponseWriter, r *http.Request) {
	logger := httputil.LoggerWithTrace(r.Context(), h.Logger)
	ownerID := h.getUserIDFromContext(w, r)
	if ownerID == "" {
		httputil.UnauthorizedErr(w, r, logger, "unauthorized - user id not found in context")
		return
	}
	loanID := chi.URLParam(r, "id")
	if loanID == "" {
		httputil.BadRequestErr(w, r, logger, "invalid url param: loan id is required")
		return
	}
	var req ConfirmDeliveryRequest
	if err := httputil.ReadJSON(w, r, &req); err != nil {
		httputil.BadRequestErr(w, r, logger, "invalid request body format")
		return
	}

	if err := validation.Validator.Struct(req); err != nil {
		httputil.BadRequestErr(w, r, logger, "delivery code must be a valid 6-digit string")
		return
	}
	_, err := h.loanCLient.Client.ConfirmDelivery(r.Context(), &pb.ConfirmDeliveryRequest{
		LoanID:       loanID,
		DeliveryCode: req.DeliveryCode,
		OwnerId: ownerID,
	})

	if err != nil {
		httputil.HandleGRPCErr(w, r, logger, err)
		return
	}
	if err := httputil.WriteJSON(w, nil, http.StatusOK, "loan delivery confirmed successfully"); err != nil {
		httputil.InternalServerErr(w, r, logger, errors.New("failed to send confirm delivery response"))
		return
	}
}
