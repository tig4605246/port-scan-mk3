package input

import (
	"encoding/binary"
	"fmt"
	"net"
)

// ValidateNoOverlap validates CIDR/IP selector rows and rejects conflicting ranges.
func ValidateNoOverlap(networks []CIDRRecord) error {
	return ValidateIPRows(networks)
}

// ValidateIPRows enforces fail-fast input rules:
// 1) each selector is inside its ip_cidr
// 2) duplicate (ip, ip_cidr) pairs are rejected
// 3) ip_cidr groups cannot overlap globally
// 4) selectors inside same ip_cidr cannot overlap.
func ValidateIPRows(rows []CIDRRecord) error {
	for i := range rows {
		if rows[i].Net == nil || rows[i].Selector == nil {
			return fmt.Errorf("row %d is not parsed", i+1)
		}
	}

	seenPair := make(map[string]int, len(rows))
	for i, row := range rows {
		key := row.Selector.String() + "|" + row.Net.String()
		if prev, ok := seenPair[key]; ok {
			return fmt.Errorf("duplicate (ip,ip_cidr) found at rows %d and %d: (%s,%s)", prev, i+1, row.Selector.String(), row.Net.String())
		}
		seenPair[key] = i + 1
	}

	for i := 0; i < len(rows); i++ {
		if !networkContains(rows[i].Net, rows[i].Selector) {
			return fmt.Errorf("ip selector %s is outside ip_cidr %s (row %d)", rows[i].Selector.String(), rows[i].Net.String(), i+1)
		}
	}

	uniqueIPCidrs := make([]*net.IPNet, 0, len(rows))
	for i := range rows {
		cur := rows[i].Net
		found := false
		for _, n := range uniqueIPCidrs {
			if n.String() == cur.String() {
				found = true
				break
			}
		}
		if !found {
			uniqueIPCidrs = append(uniqueIPCidrs, cur)
		}
	}
	for i := 0; i < len(uniqueIPCidrs); i++ {
		for j := i + 1; j < len(uniqueIPCidrs); j++ {
			a := uniqueIPCidrs[i]
			b := uniqueIPCidrs[j]
			if networksOverlap(a, b) {
				return fmt.Errorf("ip_cidr overlap detected: %s <-> %s", a.String(), b.String())
			}
		}
	}

	selectorGroups := make(map[string][]selectorWithRow)
	for i, row := range rows {
		key := row.Net.String()
		selectorGroups[key] = append(selectorGroups[key], selectorWithRow{
			selector: row.Selector,
			row:      i + 1,
		})
	}
	for ipCidr, selectors := range selectorGroups {
		for i := 0; i < len(selectors); i++ {
			for j := i + 1; j < len(selectors); j++ {
				if networksOverlap(selectors[i].selector, selectors[j].selector) {
					return fmt.Errorf("ip selector overlap detected in same ip_cidr %s: row %d (%s) <-> row %d (%s)",
						ipCidr, selectors[i].row, selectors[i].selector.String(), selectors[j].row, selectors[j].selector.String())
				}
			}
		}
	}

	return nil
}

type selectorWithRow struct {
	selector *net.IPNet
	row      int
}

func networkContains(outer, inner *net.IPNet) bool {
	if outer == nil || inner == nil {
		return false
	}
	innerStart, innerEnd, ok := ipv4Range(inner)
	if !ok {
		return false
	}
	return outer.Contains(innerStart) && outer.Contains(innerEnd)
}

func networksOverlap(a, b *net.IPNet) bool {
	if a == nil || b == nil {
		return false
	}
	aStart, aEnd, okA := ipv4Range(a)
	bStart, bEnd, okB := ipv4Range(b)
	if !okA || !okB {
		return false
	}
	aStartN := binary.BigEndian.Uint32(aStart.To4())
	aEndN := binary.BigEndian.Uint32(aEnd.To4())
	bStartN := binary.BigEndian.Uint32(bStart.To4())
	bEndN := binary.BigEndian.Uint32(bEnd.To4())
	return aStartN <= bEndN && bStartN <= aEndN
}

func ipv4Range(n *net.IPNet) (start net.IP, end net.IP, ok bool) {
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
