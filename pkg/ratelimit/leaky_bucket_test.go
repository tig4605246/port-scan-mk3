package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestLeakyBucketAcquire_WhenNoTokenAvailable_BlocksUntilRefill(t *testing.T) {
	b := NewLeakyBucket(2, 1)
	defer b.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := b.Acquire(ctx); err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}

	start := time.Now()
	if err := b.Acquire(ctx); err != nil {
		t.Fatalf("second acquire failed: %v", err)
	}
	if time.Since(start) < 450*time.Millisecond {
		t.Fatalf("expected blocking acquire")
	}
}
