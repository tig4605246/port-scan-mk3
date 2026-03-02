package scanner

import (
	"net"
	"strconv"
	"testing"
	"time"
)

func TestScanTCP_WhenPortIsOpen_ReturnsOpenStatus(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err == nil {
			_ = conn.Close()
		}
	}()

	host, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(portStr)

	res := ScanTCP(net.DialTimeout, host, port, 200*time.Millisecond)
	if res.Status != "open" {
		t.Fatalf("expected open, got %s", res.Status)
	}
	<-done
}
