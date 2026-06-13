package userHandler

import (
	httputil "books-and-trust/services/api-gateway/util"
	"books-and-trust/shared/validation"
	"errors"
	"net/http"
	"time"
)

// RegisterUserHandler godoc
//
//	@Summary		Register a new user
//	@Description	Accepts user registration details and registers a new account.
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RegisterUserRequest		true				"User registration details"
//	@Success		201		{object}	map[string]interface{}	"{ "message": "user	successfully	registered", "data": { "user_id": "string", "email": "string", "created_at": "string" } }"
//	@Failure		400		{object}	map[string]interface{}	"{ "code": "BAD_REQUEST", "message": "string", "details": "string" }"
//	@Failure		500		{object}	map[string]interface{}	"{ "code": "INTERNAL_SERVER_ERROR", "message": "string", "details": "" }"
//	@Router			/users/register [post]
func (h *UserHandler) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	logger := httputil.LoggerWithTrace(r.Context(), h.Logger)
	var req RegisterUserRequest
	if err := httputil.ReadJSON(w, r, &req); err != nil {
		httputil.BadRequestErr(w, r, logger, "invalid user register request")
		return
	}
	if err := validation.Validator.Struct(req); err != nil {
		httputil.BadRequestErr(w, r, logger, "invalid user register request")
		return
	}
	pbResp, err := h.userCLient.Client.RegisterUser(r.Context(), req.ToProto())
	if err != nil {
		httputil.HandleGRPCErr(w, r, logger, err)
		return
	}
	if pbResp == nil {
		httputil.InternalServerErr(w, r, logger, errors.New("empty response from user service"))
		return
	}
	resp := RegisterUserResponse{
		UserID:    pbResp.GetUser().Id,
		Email:     pbResp.GetUser().Email,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	if err := httputil.WriteJSON(w, resp, http.StatusCreated, "user successfully registered"); err != nil {
		httputil.InternalServerErr(w, r, logger, err)
		return
	}
}

// LoginUserHandler godoc
//
//	@Summary		Authenticate and login user
//	@Description	Accepts user credentials, validates them against the user service via gRPC, and returns a JWT access token upon successful authentication.
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		LoginUserRequest		true																	"User login credentials"
//	@Success		200		{object}	map[string]interface{}	"{ "message": "user														successfully	logged	in", "data": { "user_id": "string", "token": "string", "token_type": "string" } }"
//	@Failure		400		{object}	map[string]interface{}	"{ "code": "BAD_REQUEST", "message": "string", "details": "invalid		user			login	request" }"
//	@Failure		401		{object}	map[string]interface{}	"{ "code": "UNAUTHENTICATED", "message": "string", "details": "invalid	username		or		password" }"
//	@Failure		500		{object}	map[string]interface{}	"{ "code": "INTERNAL_SERVER_ERROR", "message": "string", "details": "" }"
//	@Router			/users/login [post]
func (h *UserHandler) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	logger := httputil.LoggerWithTrace(r.Context(), h.Logger)
	var req LoginUserRequest
	if err := httputil.ReadJSON(w, r, &req); err != nil {
		httputil.BadRequestErr(w, r, logger, "invalid user login request")
		return
	}
	if err := validation.Validator.Struct(req); err != nil {
		httputil.BadRequestErr(w, r, logger, "invalid user login request")
		return
	}
	pbRes, err := h.userCLient.Client.LoginUser(r.Context(), req.ToProto())
	if err != nil {
		httputil.HandleGRPCErr(w, r, logger, err)
		return
	}
	if pbRes == nil {
		httputil.InternalServerErr(w, r, logger, errors.New("empty response from user service"))
		return
	}
	resp := LoginUserResponse{
		UserID:    pbRes.GetUser().Id,
		Token:     pbRes.AccessToken,
		TokenType: pbRes.TokenType,
	}
	if err := httputil.WriteJSON(w, resp, http.StatusOK, "user successfully logged in"); err != nil {
		httputil.InternalServerErr(w, r, logger, err)
		return
	}
}
