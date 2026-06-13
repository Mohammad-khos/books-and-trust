package circuitbreaker

import (
	"time"

	"github.com/sony/gobreaker"
)

type SonyBreakerWrapper struct {
	instance *gobreaker.CircuitBreaker
}


func (w *SonyBreakerWrapper) Execute(req func() (any, error)) (any, error) {
	return w.instance.Execute(req)
}

func NewLoanServiceBreaker() Breaker {
	settings := gobreaker.Settings{
		Name:        "loan-service",
		MaxRequests: 5,              
		Interval:    10 * time.Second, 
		Timeout:     20 * time.Second, 

		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			println("⚡ Circuit Breaker State Changed | Name:", name, " | From:", from.String(), " | To:", to.String())
		},
	}

	sb := gobreaker.NewCircuitBreaker(settings)
	return &SonyBreakerWrapper{instance: sb}
}
