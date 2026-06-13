package retry

import (
	"context"
	"time"
)

type Config struct {
	Attempts int           
	Initial  time.Duration 
	MaxDelay time.Duration 
	Factor   float64      
}

func Do(ctx context.Context, cfg Config, operation func() error, isRetriable func(error) bool) error {
	var err error
	delay := cfg.Initial

	for i := 0; i < cfg.Attempts; i++ {
		err = operation()
		if err == nil {
			return nil 
		}

		if i == cfg.Attempts-1 {
			break
		}

		if isRetriable != nil && !isRetriable(err) {
			return err
		}

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}

		delay = time.Duration(float64(delay) * cfg.Factor)
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}
	}

	return err
}