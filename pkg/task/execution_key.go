package task

import (
	"fmt"
	"net"
	"strings"
)

// BuildExecutionKey returns the canonical dedup key `dst_ip:port/protocol`.
// It validates IPv4 target, TCP protocol, and port range before generating
// the key used by execution-layer dedup logic.
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
		return "", fmt.Errorf("invalid protocol: %q", protocol)
	}
	return fmt.Sprintf("%s:%d/%s", ip.String(), port, proto), nil
}
