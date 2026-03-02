package ratelimit

import (
	"context"
	"testing"
)

func TestLeakyBucketAcquire_WhenContextCanceled_ReturnsError(t *testing.T) {
	b := NewLeakyBucket(1, 1)
	defer b.Close()

	if err := b.Acquire(context.Background()); err != nil {
		t.Fatalf("unexpected first acquire err: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := b.Acquire(ctx); err == nil {
		t.Fatal("expected context error")
	}
}
