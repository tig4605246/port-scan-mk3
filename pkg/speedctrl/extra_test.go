package speedctrl

import (
	"context"
	"testing"
)

func TestGateAndKeyboardLoop(t *testing.T) {
	c := NewController()
	select {
	case <-c.Gate():
	default:
		t.Fatal("expected open gate")
	}
	if err := StartKeyboardLoop(context.Background(), c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
