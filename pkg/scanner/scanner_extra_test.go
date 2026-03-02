package scanner

import (
	"errors"
	"net"
	"testing"
	"time"
)

type timeoutErr struct{}

func (timeoutErr) Error() string   { return "timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return false }

func TestScanTCP_WhenDialTimeout_ReturnsCloseTimeoutStatus(t *testing.T) {
	dial := func(string, string, time.Duration) (net.Conn, error) {
		return nil, timeoutErr{}
	}
	res := ScanTCP(dial, "127.0.0.1", 80, 10*time.Millisecond)
	if res.Status != "close(timeout)" {
		t.Fatalf("expected close(timeout), got %s", res.Status)
	}
}

func TestScanTCP_WhenConnectionRefused_ReturnsCloseStatus(t *testing.T) {
	dial := func(string, string, time.Duration) (net.Conn, error) {
		return nil, errors.New("connection refused")
	}
	res := ScanTCP(dial, "127.0.0.1", 80, 10*time.Millisecond)
	if res.Status != "close" {
		t.Fatalf("expected close, got %s", res.Status)
	}
}
