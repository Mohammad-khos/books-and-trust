package circuitbreaker

type Breaker interface {
	Execute(req func() (any, error)) (any, error)
}