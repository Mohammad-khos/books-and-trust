package prometheus

import "github.com/prometheus/client_golang/prometheus"

type Factory struct {
	namespace string
	subsystem string
}

func NewFactory(namespace, subsystem string) *Factory {
	return &Factory{
		namespace: namespace,
		subsystem: subsystem,
	}
}

func (f *Factory) NewCounter(name, help string, labels []string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: f.namespace,
			Subsystem: f.subsystem,
			Name:      name,
			Help:      help,
		},
		labels,
	)
}

func (f *Factory) NewGauge(name, help string, labels []string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: f.namespace,
			Subsystem: f.subsystem,
			Name:      name,
			Help:      help,
		},
		labels,
	)
}

func (f *Factory) NewHistogram(name, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: f.namespace,
			Subsystem: f.subsystem,
			Name:      name,
			Help:      help,
			Buckets:   buckets,
		},
		labels,
	)
}