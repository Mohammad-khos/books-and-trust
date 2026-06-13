package userHandler

import (
	"books-and-trust/services/api-gateway/internal/client"
	"books-and-trust/services/api-gateway/internal/middleware"
	httputil "books-and-trust/services/api-gateway/util"
	"net/http"

	"go.uber.org/zap"
)


type UserHandler struct {
	Logger     *zap.SugaredLogger
	userCLient *client.UserClient
}

func NewUserHandler(logger *zap.SugaredLogger, client *client.UserClient) *UserHandler {
	return &UserHandler{
		Logger:     logger,
		userCLient: client,
	}
}

func (h *UserHandler) getUserIDFromContext(w http.ResponseWriter, r *http.Request) string {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		httputil.UnauthorizedErr(w, r, h.Logger, "unauthorized access")
		return ""
	}
	return userID
}

