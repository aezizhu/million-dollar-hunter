package breaker

import (
	"context"

	"github.com/sony/gobreaker"
)

type Breaker struct {
	cb *gobreaker.CircuitBreaker
}

func New(name string) *Breaker {
	st := gobreaker.Settings{
		Name:        name,
		MaxRequests: 5,
		Interval:    0,
		Timeout:     0,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failures := counts.ConsecutiveFailures
			return failures >= 5
		},
	}
	return &Breaker{cb: gobreaker.NewCircuitBreaker(st)}
}

func (b *Breaker) Do(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	return b.cb.Execute(fn)
}
