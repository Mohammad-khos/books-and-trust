package userHandler

import (
	httputil "books-and-trust/services/api-gateway/util"
	pb "books-and-trust/shared/proto/user"
	"books-and-trust/shared/validation"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// UpdateUser godoc
//
//	@Summary		Update user details
//	@Description	Updates an existing user's information via gRPC user service.
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			request	body	UpdateUserRequest	true	"User update details"
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]interface{}	"{ "message": "user													successfully	updated", "data": { "user_id": "string", "name": "string", "email": "string", "username": "string" } }"
//	@Failure		400	{object}	map[string]interface{}	"{ "code": "BAD_REQUEST", "message": "string", "details": "invalid	user			update	request" }"
//	@Failure		500	{object}	map[string]interface{}	"{ "code": "INTERNAL_SERVER_ERROR", "message": "string", "details": "" }"
//	@Router			/users/update [patch]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	logger := httputil.LoggerWithTrace(r.Context(), h.Logger)
	var req UpdateUserRequest
	if err := httputil.ReadJSON(w, r, &req); err != nil {
		httputil.BadRequestErr(w, r, logger, "invalid user update request")
		return
	}
	if err := validation.Validator.Struct(req); err != nil {
		httputil.BadRequestErr(w, r, logger, "invalid user update request")
		return
	}
	contextUserID := h.getUserIDFromContext(w, r)
	if contextUserID == "" {
		return
	}
	if req.UserID != contextUserID {
		httputil.ForbiddenErr(w, r, logger, "can not update user")
		return
	}
	pbResp, err := h.userCLient.Client.UpdateUser(r.Context(), req.ToProto())
	if err != nil {
		logger.Errorw("gRPC UpdateUser failed", "error", err)
		httputil.HandleGRPCErr(w, r, logger, err)
		return
	}
	if pbResp == nil {
		logger.Errorw("gRPC UpdateUser failed", "error", err)
		httputil.InternalServerErr(w, r, logger, errors.New("empty response from user service"))
		return
	}
	resp := UpdateUserResponse{
		UserID:   pbResp.GetUser().GetId(),
		Name:     pbResp.GetUser().GetName(),
		Email:    pbResp.GetUser().GetEmail(),
		Username: pbResp.GetUser().GetUsername(),
	}
	if err := httputil.WriteJSON(w, resp, http.StatusOK, "user successfully updated"); err != nil {
		logger.Errorw("gRPC UpdateUser failed", "error", err)
		httputil.InternalServerErr(w, r, logger, err)
		return
	}
}

// GetUserByIDHandler godoc
//
//	@Summary		Get user by ID
//	@Description	Fetches user details from the user microservice using their unique ID provided in the URL path.
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"User id"
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]interface{}	"{ "message": "user													successfully	fetched", "data": { "user_id": "string", "name": "string", "email": "string", "username": "string" } }"
//	@Failure		400	{object}	map[string]interface{}	"{ "code": "BAD_REQUEST", "message": "string", "details": "invalid	get				user	request" }"
//	@Failure		500	{object}	map[string]interface{}	"{ "code": "INTERNAL_SERVER_ERROR", "message": "string", "details": "" }"
//	@Router			/users/{id} [get]
func (h *UserHandler) GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	logger := httputil.LoggerWithTrace(r.Context(), h.Logger)
	userID := chi.URLParam(r, "id")
	if userID == "" {
		httputil.BadRequestErr(w, r, logger, "invalid get user request")
		return
	}
	contextUserID := h.getUserIDFromContext(w, r)
	if contextUserID == "" {
		return
	}
	if userID != contextUserID {
		httputil.ForbiddenErr(w, r, logger, "can not update user")
		return
	}
	pbResp, err := h.userCLient.Client.GetUserByID(r.Context(), &pb.GetUserByIDRequest{
		Id: userID,
	})
	if err != nil {
		httputil.HandleGRPCErr(w, r, logger, err)
		return
	}
	if pbResp == nil {
		httputil.InternalServerErr(w, r, logger, errors.New("empty response from user service"))
		return
	}

	resp := GetUserByIDResponse{
		UserID:   pbResp.GetUser().GetId(),
		Name:     pbResp.GetUser().GetName(),
		Email:    pbResp.GetUser().GetEmail(),
		Username: pbResp.GetUser().GetUsername(),
	}
	if err := httputil.WriteJSON(w, resp, http.StatusOK, "user successfully fetched"); err != nil {
		httputil.InternalServerErr(w, r, logger, err)
		return
	}
}

// DeleteUserHandler godoc
//
//	@Summary		Delete user account
//	@Description	Deletes a user account from the system using their unique ID provided in the URL path. Only the authenticated user themselves can delete their own account.
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"User ID to delete"
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]interface{}	"{ "message": "user													successfully	deleted", "data": null }"
//	@Failure		400	{object}	map[string]interface{}	"{ "code": "BAD_REQUEST", "message": "string", "details": "invalid	delete			user	request" }"
//	@Failure		403	{object}	map[string]interface{}	"{ "code": "FORBIDDEN", "message": "string", "details": "you		do				not		have	permission	to	delete	this	user" }"
//	@Failure		500	{object}	map[string]interface{}	"{ "code": "INTERNAL_SERVER_ERROR", "message": "string", "details": "" }"
//	@Router			/users/{id} [delete]
func (h *UserHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	logger := httputil.LoggerWithTrace(r.Context(), h.Logger)
	userID := chi.URLParam(r, "id")
	if userID == "" {
		httputil.BadRequestErr(w, r, logger, "invalid delete user request")
		return
	}
	contextUserID := h.getUserIDFromContext(w, r)
	if contextUserID == "" {
		return
	}
	if userID != contextUserID {
		httputil.ForbiddenErr(w, r, logger, "can not delete user")
		return
	}
	_, err := h.userCLient.Client.DeleteUserByID(r.Context(), &pb.DeleteUserByIDRequest{
		UserId: userID,
	})
	if err != nil {
		httputil.HandleGRPCErr(w, r, logger, err)
		return
	}
	if err := httputil.WriteJSON(w, nil, http.StatusOK, "user successfully deleted"); err != nil {
		httputil.InternalServerErr(w, r, logger, err)
		return
	}
}
