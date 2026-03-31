package netutil

import "net"

// IPRange computes the start and end IPv4 addresses for a CIDR network.
// Returns nil, nil, false if the input is nil or not an IPv4 network.
func IPRange(n *net.IPNet) (start net.IP, end net.IP, ok bool) {
	if n == nil {
		return nil, nil, false
	}
	base := n.IP.To4()
	if base == nil {
		return nil, nil, false
	}
	mask := n.Mask
	if len(mask) != net.IPv4len {
		return nil, nil, false
	}
	start = make(net.IP, net.IPv4len)
	for i := 0; i < net.IPv4len; i++ {
		start[i] = base[i] & mask[i]
	}
	end = make(net.IP, net.IPv4len)
	for i := 0; i < net.IPv4len; i++ {
		end[i] = start[i] | ^mask[i]
	}
	return start, end, true
}