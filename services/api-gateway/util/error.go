package util

import (
	"net/http"

	"go.uber.org/zap"
)

// logAndWrite is a helper that logs the error and writes the HTTP response.
func logAndWrite(
	w http.ResponseWriter,
	r *http.Request,
	logger *zap.SugaredLogger,
	status int,
	logMsg string,
	code string,
	message string,
	details string,
) {
	logger.Warnw(logMsg,
		"path", r.URL.Path,
		"method", r.Method,
		"status", status,
	)
	WriteError(w, status, code, message, details)

}

// BadRequestErr responds with 400 Bad Request.
// Use when the client sends malformed or invalid input.
func BadRequestErr(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger, details string) {
	logAndWrite(w, r, logger, http.StatusBadRequest,
		"bad request",
		"BAD_REQUEST",
		"The request could not be understood or was missing required parameters.",
		details,
	)
}

// UnauthorizedErr responds with 401 Unauthorized.
// Use when the request lacks valid authentication credentials.
func UnauthorizedErr(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger, details string) {
	logAndWrite(w, r, logger, http.StatusUnauthorized,
		"unauthorized request",
		"UNAUTHORIZED",
		"Authentication is required to access this resource.",
		details,
	)
}

// ForbiddenErr responds with 403 Forbidden.
// Use when the client is authenticated but does not have permission.
func ForbiddenErr(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger, details string) {
	logAndWrite(w, r, logger, http.StatusForbidden,
		"forbidden request",
		"FORBIDDEN",
		"You do not have permission to access this resource.",
		details,
	)
}

// NotFoundErr responds with 404 Not Found.
// Use when the requested resource does not exist.
func NotFoundErr(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger, details string) {
	logAndWrite(w, r, logger, http.StatusNotFound,
		"resource not found",
		"NOT_FOUND",
		"The requested resource could not be found.",
		details,
	)
}

// MethodNotAllowedErr responds with 405 Method Not Allowed.
// Use when the HTTP method is not supported for the given route.
func MethodNotAllowedErr(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger, details string) {
	logAndWrite(w, r, logger, http.StatusMethodNotAllowed,
		"method not allowed",
		"METHOD_NOT_ALLOWED",
		"The HTTP method is not allowed for this endpoint.",
		details,
	)
}

// ConflictErr responds with 409 Conflict.
// Use when the request conflicts with the current state of the resource (e.g. duplicate entry).
func ConflictErr(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger, details string) {
	logAndWrite(w, r, logger, http.StatusConflict,
		"conflict",
		"CONFLICT",
		"The request could not be completed due to a conflict with the current state of the resource.",
		details,
	)
}

// UnprocessableEntityErr responds with 422 Unprocessable Entity.
// Use when input is syntactically valid but semantically incorrect (e.g. failed validation).
func UnprocessableEntityErr(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger, details string) {
	logAndWrite(w, r, logger, http.StatusUnprocessableEntity,
		"unprocessable entity",
		"UNPROCESSABLE_ENTITY",
		"The request was well-formed but contains semantic errors.",
		details,
	)
}

// TooManyRequestsErr responds with 429 Too Many Requests.
// Use when the client has exceeded the rate limit.
func TooManyRequestsErr(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger, details string) {
	logAndWrite(w, r, logger, http.StatusTooManyRequests,
		"rate limit exceeded",
		"TOO_MANY_REQUESTS",
		"You have exceeded the allowed number of requests. Please try again later.",
		details,
	)
}

// InternalServerErr responds with 500 Internal Server Error.
// Use for unexpected server-side failures. Logs at Error level.
func InternalServerErr(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger, err error) {
	logger.Errorw("internal server error",
		"path", r.URL.Path,
		"method", r.Method,
		"status", http.StatusInternalServerError,
		"error", err,
	)
	WriteError(w, http.StatusInternalServerError,
		"INTERNAL_SERVER_ERROR",
		"An unexpected error occurred on our end. Please try again later.",
		"",
	)
}

// ServiceUnavailableErr responds with 503 Service Unavailable.
// Use when the server is temporarily unable to handle the request (e.g. dependency down).
func ServiceUnavailableErr(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger, details string) {
	logger.Errorw("service unavailable",
		"path", r.URL.Path,
		"method", r.Method,
		"status", http.StatusServiceUnavailable,
	)
	WriteError(w, http.StatusServiceUnavailable,
		"SERVICE_UNAVAILABLE",
		"The service is temporarily unavailable. Please try again later.",
		details,
	)
}
