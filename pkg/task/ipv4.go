package task

import (
	"encoding/binary"
	"fmt"
	"net"
)

func CountIPv4Hosts(ipNet *net.IPNet) (int, error) {
	if ipNet == nil {
		return 0, fmt.Errorf("nil network")
	}
	v4 := ipNet.IP.To4()
	if v4 == nil {
		return 0, fmt.Errorf("only ipv4 is supported")
	}
	ones, bits := ipNet.Mask.Size()
	if bits != 32 {
		return 0, fmt.Errorf("unexpected mask bits: %d", bits)
	}
	count := 1 << (bits - ones)
	return count, nil
}

func IndexToIPv4Target(ipNet *net.IPNet, ports []int, idx int) (string, int, error) {
	if ipNet == nil {
		return "", 0, fmt.Errorf("nil network")
	}
	if len(ports) == 0 {
		return "", 0, fmt.Errorf("empty ports")
	}
	if idx < 0 {
		return "", 0, fmt.Errorf("negative index")
	}
	totalHosts, err := CountIPv4Hosts(ipNet)
	if err != nil {
		return "", 0, err
	}
	ipOffset := idx / len(ports)
	portIdx := idx % len(ports)
	if ipOffset >= totalHosts {
		return "", 0, fmt.Errorf("index out of range")
	}

	base := binary.BigEndian.Uint32(ipNet.IP.To4())
	curr := base + uint32(ipOffset)
	outIP := make(net.IP, 4)
	binary.BigEndian.PutUint32(outIP, curr)
	return outIP.String(), ports[portIdx], nil
}
