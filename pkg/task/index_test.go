package task

import "testing"

func TestIndexToTarget_WhenIndexInRange_ReturnsStableMapping(t *testing.T) {
	ips := []string{"10.0.0.1", "10.0.0.2"}
	ports := []int{80, 443}

	ip, port := IndexToTarget(3, ips, ports)
	if ip != "10.0.0.2" || port != 443 {
		t.Fatalf("unexpected mapping: %s:%d", ip, port)
	}
}

func TestChunkRemaining_WhenWithinRange_ReturnsRemainingCount(t *testing.T) {
	c := Chunk{NextIndex: 2, TotalCount: 6}
	if c.Remaining() != 4 {
		t.Fatalf("remaining mismatch: %d", c.Remaining())
	}
}
