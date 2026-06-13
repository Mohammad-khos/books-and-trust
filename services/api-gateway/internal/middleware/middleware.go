package middleware

import "net/http"

type GatewayMiddleware interface {
	CspMiddleware(next http.Handler) http.Handler
	AuthMiddleware(next http.Handler) http.Handler
	AdminsMiddleware(next http.Handler) http.Handler
	GlobalRateLimitMiddleware(next http.Handler) http.Handler
	IPRateLimitMiddleware(next http.Handler) http.Handler
	UserRateLimitMiddleware(next http.Handler) http.Handler
	SecurityHeadersMiddleware(next http.Handler) http.Handler
	RequestLoggerMiddleware(next http.Handler) http.Handler
	MetricsMiddleware(next http.Handler) http.Handler
}

type appMiddlewareHub struct {
	auth      *AuthMiddleware
	csp       *CSPMiddleware
	admin     *AdminMiddleware
	limiter   *RateLimiterMiddleware
	security  *SecurityHeadersMiddleware
	reqLogger *RequestLoggerMiddleware
	metrics   *MetricsMiddleware
}

func NewGatewayMiddleware(
	auth *AuthMiddleware,
	csp *CSPMiddleware,
	admin *AdminMiddleware,
	ratelimiter *RateLimiterMiddleware,
	security *SecurityHeadersMiddleware,
	reqLogger *RequestLoggerMiddleware,
	metrics *MetricsMiddleware) GatewayMiddleware {
	return &appMiddlewareHub{
		auth:      auth,
		csp:       csp,
		admin:     admin,
		limiter:   ratelimiter,
		security:  security,
		reqLogger: reqLogger,
		metrics:   metrics,
	}
}
