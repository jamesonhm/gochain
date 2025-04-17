package rate

import (
	"context"
	"fmt"
	"time"
)

type Limiter struct {
	ticker *time.Ticker
	ch     chan time.Time
	done   chan struct{}
}

func New(period time.Duration, count int) *Limiter {
	rate := period / time.Duration(count)

	rl := &Limiter{
		ch:     make(chan time.Time),
		done:   make(chan struct{}),
		ticker: time.NewTicker(rate),
	}

	go func() {
		defer rl.ticker.Stop()
		for {
			select {
			case <-rl.done:
				fmt.Println("Limiter goroutine done ch")
				return
			case t := <-rl.ticker.C:
				rl.ch <- t
			}
		}
	}()

	return rl
}

func (rl *Limiter) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		fmt.Println("Wait context cancelled...")
		close(rl.done)
		return ctx.Err()
	case <-rl.ch:
		return nil
	}
}
