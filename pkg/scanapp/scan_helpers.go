package scanapp

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
	"syscall"
)

func indexToRuntimeTarget(targets []scanTarget, ports []int, idx int) (scanTarget, int, error) {
	if len(targets) == 0 {
		return scanTarget{}, 0, fmt.Errorf("empty targets")
	}
	if len(ports) == 0 {
		return scanTarget{}, 0, fmt.Errorf("empty ports")
	}
	if idx < 0 {
		return scanTarget{}, 0, fmt.Errorf("negative index")
	}
	targetIdx := idx / len(ports)
	portIdx := idx % len(ports)
	if targetIdx >= len(targets) {
		return scanTarget{}, 0, fmt.Errorf("index out of range")
	}
	return targets[targetIdx], ports[portIdx], nil
}

func ipv4ToUint32(ip string) uint32 {
	parsed := net.ParseIP(ip).To4()
	if parsed == nil {
		return 0
	}
	return binary.BigEndian.Uint32(parsed)
}

func parsePortRows(rows []string) ([]int, error) {
	ports := make([]int, 0, len(rows))
	for _, row := range rows {
		parts := strings.Split(strings.TrimSpace(row), "/")
		if len(parts) != 2 || strings.ToLower(parts[1]) != "tcp" {
			return nil, fmt.Errorf("invalid chunk port row: %s", row)
		}
		n, err := strconv.Atoi(parts[0])
		if err != nil || n < 1 || n > 65535 {
			return nil, fmt.Errorf("invalid chunk port number: %s", row)
		}
		ports = append(ports, n)
	}
	return ports, nil
}

func defaultString(primary, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return fallback
}

func ensureFDLimit(workers int) error {
	var lim syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &lim); err != nil {
		return nil
	}
	minNeed := uint64(1024)
	if workers > 0 {
		workerNeed := uint64(workers * 8)
		if workerNeed > minNeed {
			minNeed = workerNeed
		}
	}
	if lim.Cur < minNeed {
		return fmt.Errorf("file descriptor limit too low: %d (need >= %d)", lim.Cur, minNeed)
	}
	return nil
}
