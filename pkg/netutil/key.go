package netutil

import (
	"fmt"
	"net"
	"strings"
)

// BuildExecutionKey creates a standardized execution key from dst IP, port, and protocol.
func BuildExecutionKey(dstIP string, port int, protocol string) (string, error) {
	ip := net.ParseIP(strings.TrimSpace(dstIP)).To4()
	if ip == nil {
		return "", fmt.Errorf("invalid dst_ip: %q", dstIP)
	}
	if port < 1 || port > 65535 {
		return "", fmt.Errorf("invalid port: %d", port)
	}
	proto := strings.ToLower(strings.TrimSpace(protocol))
	if proto != "tcp" {
		return "", fmt.Errorf("invalid protocol: %q", proto)
	}
	return fmt.Sprintf("%s:%d/%s", ip.String(), port, proto), nil
}