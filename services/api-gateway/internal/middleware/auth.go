package middleware

import (
	"books-and-trust/services/api-gateway/internal/client"
	httputil "books-and-trust/services/api-gateway/util"
	pb "books-and-trust/shared/proto/user"
	"context"
	"errors"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type AuthMiddleware struct {
	Logger     *zap.SugaredLogger
	UserClient *client.UserClient
}

func NewAuthMiddleware(
	logger *zap.SugaredLogger,
	userClient *client.UserClient,
) *AuthMiddleware {
	return &AuthMiddleware{
		Logger:     logger,
		UserClient: userClient,
	}
}

func (m *appMiddlewareHub) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := httputil.LoggerWithTrace(r.Context(), m.auth.Logger)
		header := r.Header.Get("Authorization")
		if header == "" {
			httputil.UnauthorizedErr(w, r, logger, "missing authorization header")
			return
		}
		headerSlice := strings.Split(header, " ")
		if len(headerSlice) != 2 || headerSlice[0] != "Bearer" || headerSlice[1] == "" {
			httputil.UnauthorizedErr(w, r, logger, "header malformed")
			return
		}
		token := headerSlice[1]

		pbResp, err := m.auth.UserClient.Client.VerifyToken(r.Context(), &pb.VerifyTokenRequest{
			Token: token,
		})
		if err != nil {
			httputil.HandleGRPCErr(w, r, logger, err)
			return
		}
		if pbResp == nil {
			httputil.InternalServerErr(w, r, logger, errors.New("somthing went wrong"))
			return
		}
		//save user id to request context
		ctx := context.WithValue(r.Context(), UserIDKey, pbResp.GetUserId())
		*r = *r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
