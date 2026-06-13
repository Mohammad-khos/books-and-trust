package prometheus

import "github.com/prometheus/client_golang/prometheus"

type HTTPMetrics struct {
	RequestTotal   *prometheus.CounterVec
	RequestLatency *prometheus.HistogramVec
}

func NewHTTPMetrics(factory *Factory) *HTTPMetrics {
	return &HTTPMetrics{
		RequestTotal: factory.NewCounter(
			"http_requests_total",
			"Total HTTP requests",
			[]string{"method", "path", "status"},
		),

		RequestLatency: factory.NewHistogram(
			"http_request_duration_seconds",
			"HTTP request latency",
			[]string{"method", "path", "status"},
			prometheus.DefBuckets,
		),
	}
}
