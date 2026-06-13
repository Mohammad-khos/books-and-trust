package middleware

import (
	"books-and-trust/services/api-gateway/internal/infra/metrics"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type MetricsMiddleware struct {
	metrics *metrics.Metrics
}

func NewMetricsMiddleware(m *metrics.Metrics) *MetricsMiddleware {
	return &MetricsMiddleware{
		metrics: m,
	}
}

func (m *appMiddlewareHub) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ignore metrics endpoint and options method
		if r.URL.Path == "/metrics" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()

		m.metrics.metrics.InFlight.Inc()
		defer m.metrics.metrics.InFlight.Dec()
		
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		
		next.ServeHTTP(ww, r)
		route := chi.RouteContext(r.Context()).RoutePattern()

		status := strconv.Itoa(ww.Status())
		duration := time.Since(start).Seconds()

		// 1. request count
		m.metrics.metrics.RequestTotal.
			WithLabelValues(r.Method, route, status).
			Inc()

		// 2. latency
		m.metrics.metrics.RequestLatency.
			WithLabelValues(r.Method, route, status).
			Observe(duration)
	})

}
