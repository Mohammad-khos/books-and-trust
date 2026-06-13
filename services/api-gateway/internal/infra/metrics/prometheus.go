package metrics

import (
	p "books-and-trust/shared/metrics/prometheus"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	RequestTotal   *prometheus.CounterVec
	RequestLatency *prometheus.HistogramVec
	InFlight       prometheus.Gauge
}

func NewMetrics(factory *p.Factory) *Metrics {
	return &Metrics{
		RequestTotal: factory.NewCounter(
			"http_requests_total",
			"Total HTTP requests",
			[]string{"method", "route", "status"},
		),

		RequestLatency: factory.NewHistogram(
			"http_request_duration_seconds",
			"Request latency",
			[]string{"method", "route", "status"},
			prometheus.DefBuckets,
		),

		InFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Current in-flight requests",
			},
		),
	}
}
