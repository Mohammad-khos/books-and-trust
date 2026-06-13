package middleware

import (
	"books-and-trust/services/api-gateway/util"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type CtxKey string
const UserIDKey CtxKey = "user_id"

type RequestLoggerMiddleware struct {
	logger *zap.SugaredLogger
}

func NewRequestLoggerMiddleware(logger *zap.SugaredLogger) *RequestLoggerMiddleware {
	return &RequestLoggerMiddleware{
		logger: logger,
	}
}

func (m *appMiddlewareHub) RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)
		logger := util.LoggerWithTrace(r.Context(), m.reqLogger.logger)

		var userID string
		if v := r.Context().Value(UserIDKey); v != nil {
			if s, ok := v.(string); ok {
				userID = s
			}
		}

		logger.Infow("http request",
			"request_id", middleware.GetReqID(r.Context()),
			"user_id", userID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"bytes", ww.BytesWritten(),
			"duration", time.Since(start),
			"remote_ip", getClientIP(r),
		)
	})
}
