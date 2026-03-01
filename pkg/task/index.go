package task

func IndexToTarget(idx int, ips []string, ports []int) (string, int) {
	if idx < 0 || len(ips) == 0 || len(ports) == 0 {
		return "", 0
	}
	ipIdx := idx / len(ports)
	portIdx := idx % len(ports)
	if ipIdx >= len(ips) {
		return "", 0
	}
	return ips[ipIdx], ports[portIdx]
}
