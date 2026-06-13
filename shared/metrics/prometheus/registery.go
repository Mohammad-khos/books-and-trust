package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Registry struct {
	Prom *prometheus.Registry
}

func NewRegistry() *Registry {
	return &Registry{
		Prom: prometheus.NewRegistry(),
	}
}