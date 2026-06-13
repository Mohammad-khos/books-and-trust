package loanHandler

import (
	httputil "books-and-trust/services/api-gateway/util"
	"books-and-trust/shared/validation"
	"net/http"
)

// BanUserHandler godoc
//
//	@Summary		Ban a user from the platform
//	@Description	Bans a user by their UUID, preventing them from creating new loans or borrowing books. This action is strictly restricted to Administrators.
//	@Tags			admin,loans
//	@Accept			json
//	@Produce		json
//	@Param			request	body	BanUserRequest	true	"Ban User Request Body"
//	@Security		BearerAuth
//	@Success		200	{object}	contracts.ResponseSuccess	"user banned successfully"
//	@Failure		400	{object}	contracts.ResponseError		"invalid ban user request"
//	@Failure		401	{object}	contracts.ResponseError		"unauthorized"
//	@Failure		403	{object}	contracts.ResponseError		"forbidden - admin access required"
//	@Failure		500	{object}	contracts.ResponseError		"internal server error"
//	@Router			/admin/users/ban [post]
func (h *LoanHandler) BanUserHandler(w http.ResponseWriter, r *http.Request) {
	logger := httputil.LoggerWithTrace(r.Context(), h.Logger)
	var req BanUserRequest
	if err := httputil.ReadJSON(w, r, &req); err != nil {
		httputil.BadRequestErr(w, r, logger, "invalid ban user request")
		return
	}
	if err := validation.Validator.Struct(req); err != nil {
		httputil.BadRequestErr(w, r, logger, "invalid ban user request")
		return
	}
	_, err := h.loanCLient.Client.BanUser(r.Context(), req.ToProto())
	if err != nil {
		httputil.HandleGRPCErr(w, r, logger, err)
		return
	}
	if err := httputil.WriteJSON(w, nil, http.StatusOK, "user banned successfully"); err != nil {
		httputil.InternalServerErr(w, r, logger, err)
		return
	}
}
