package middleware

import (
	"books-and-trust/services/api-gateway/internal/ratelimiter"
	"books-and-trust/services/api-gateway/util"
	"net"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type RateLimiterMiddleware struct {
	ratelimiter ratelimiter.RateLimiter
	logger      *zap.SugaredLogger
	MaxTokens   int
	RefillRate  int
}

func NewRateLimiterMiddleware(
	rt ratelimiter.RateLimiter,
	logger *zap.SugaredLogger,
	MaxTokens int,
	RefillRate int) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		ratelimiter: rt,
		logger:      logger,
		MaxTokens:1000 ,
		RefillRate: 10,
	}
}

func (m *appMiddlewareHub) GlobalRateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := "rl:global"
		maxTokens := 1000
		refillRate := 100

		allowed, err := m.limiter.ratelimiter.Allow(r.Context(), key, maxTokens, refillRate)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		if !allowed {
			w.Header().Set("Retry-After", "1")
			util.TooManyRequestsErr(w, r, m.limiter.logger, "Server is busy. Please try again later")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *appMiddlewareHub) IPRateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		key := "rl:ip:" + ip

		allowed, err := m.limiter.ratelimiter.Allow(r.Context(), key, m.limiter.MaxTokens, m.limiter.RefillRate)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		if !allowed {
			w.Header().Set("Retry-After", "5")
			util.TooManyRequestsErr(w, r, m.limiter.logger, "Server is busy. Please try again later")
			return
		}

		next.ServeHTTP(w, r)
	})
}
func (m *appMiddlewareHub) UserRateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(UserIDKey).(string)
		if !ok || userID == "" {
			util.ForbiddenErr(w, r, m.limiter.logger, "unauthorized - user id missing for rate limiting")
			return
		}

		key := "rl:user:" + userID
		maxTokens := 10
		refillRate := 1

		allowed, err := m.limiter.ratelimiter.Allow(r.Context(), key, maxTokens, refillRate)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		if !allowed {
			w.Header().Set("Retry-After", "1")
			util.TooManyRequestsErr(w, r, m.limiter.logger, "Server is busy. Please try again later")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
