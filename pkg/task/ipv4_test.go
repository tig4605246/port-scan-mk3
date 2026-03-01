package task

import (
	"net"
	"testing"
)

func TestIndexToIPv4Target(t *testing.T) {
	_, n, _ := net.ParseCIDR("10.0.0.0/30")
	ip, port, err := IndexToIPv4Target(n, []int{80, 443}, 3)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ip != "10.0.0.1" || port != 443 {
		t.Fatalf("unexpected target: %s:%d", ip, port)
	}
}

func TestCountIPv4Hosts(t *testing.T) {
	_, n, _ := net.ParseCIDR("10.0.0.0/30")
	count, err := CountIPv4Hosts(n)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if count != 4 {
		t.Fatalf("unexpected host count: %d", count)
	}
}
