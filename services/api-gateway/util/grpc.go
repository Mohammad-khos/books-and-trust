package util

import (
	"net/http"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HandleGRPCErr(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger, err error) {
	st, ok := status.FromError(err)
	if !ok {
		InternalServerErr(w, r, logger, err)
		return
	}

	switch st.Code() {

	case codes.InvalidArgument:
		BadRequestErr(w, r, logger, st.Message())

	case codes.NotFound:
		NotFoundErr(w, r, logger, st.Message())

	case codes.AlreadyExists:
		BadRequestErr(w, r, logger, st.Message())

	case codes.Unauthenticated:
		UnauthorizedErr(w, r, logger, st.Message())
	case codes.PermissionDenied:
		ForbiddenErr(w, r, logger, st.Message())

	default:
		logger.Errorw("downstream grpc service failure", "code", st.Code(), "error", st.Message())
		InternalServerErr(w, r, logger, err)
	}
}
