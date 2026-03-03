package task

import "testing"

func TestBuildExecutionKey_WhenInputsValid_ReturnsCanonicalKey(t *testing.T) {
	got, err := BuildExecutionKey(" 10.0.0.8 ", 443, "TCP")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "10.0.0.8:443/tcp" {
		t.Fatalf("unexpected key: %s", got)
	}
}

func TestBuildExecutionKey_WhenInputsInvalid_ReturnsError(t *testing.T) {
	if _, err := BuildExecutionKey("bad-ip", 80, "tcp"); err == nil {
		t.Fatal("expected invalid ip error")
	}
	if _, err := BuildExecutionKey("10.0.0.1", 0, "tcp"); err == nil {
		t.Fatal("expected invalid port error")
	}
	if _, err := BuildExecutionKey("10.0.0.1", 80, "udp"); err == nil {
		t.Fatal("expected invalid protocol error")
	}
}
