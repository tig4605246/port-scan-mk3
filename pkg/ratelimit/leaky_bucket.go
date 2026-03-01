package ratelimit

import (
	"context"
	"time"
)

type LeakyBucket struct {
	tokens chan struct{}
	stop   chan struct{}
}

func NewLeakyBucket(rate, capacity int) *LeakyBucket {
	if rate <= 0 {
		rate = 1
	}
	if capacity <= 0 {
		capacity = 1
	}

	b := &LeakyBucket{
		tokens: make(chan struct{}, capacity),
		stop:   make(chan struct{}),
	}
	for i := 0; i < capacity; i++ {
		b.tokens <- struct{}{}
	}

	interval := time.Second / time.Duration(rate)
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				select {
				case b.tokens <- struct{}{}:
				default:
				}
			case <-b.stop:
				return
			}
		}
	}()

	return b
}

func (b *LeakyBucket) Acquire(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-b.tokens:
		return nil
	}
}

func (b *LeakyBucket) Close() {
	select {
	case <-b.stop:
	default:
		close(b.stop)
	}
}
