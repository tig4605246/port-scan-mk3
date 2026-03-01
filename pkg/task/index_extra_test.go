package task

import "testing"

func TestIndexToTarget_OutOfRange(t *testing.T) {
	ip, port := IndexToTarget(99, []string{"1.1.1.1"}, []int{80})
	if ip != "" || port != 0 {
		t.Fatalf("expected empty target, got %s:%d", ip, port)
	}
}

func TestRemaining_NoNegative(t *testing.T) {
	c := Chunk{NextIndex: 10, TotalCount: 6}
	if c.Remaining() != 0 {
		t.Fatalf("expected 0, got %d", c.Remaining())
	}
}
