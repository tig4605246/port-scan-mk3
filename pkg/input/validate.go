package input

import (
	"fmt"
	"net"

	"github.com/xuxiping/port-scan-mk3/pkg/netutil"
)

// ValidateNoOverlap validates CIDR/IP selector rows and rejects conflicting ranges.
func ValidateNoOverlap(networks []CIDRRecord) error {
	return ValidateIPRows(networks)
}

// ValidateIPRows enforces fail-fast input rules:
// 1) each selector is inside its ip_cidr
// 2) duplicate (src, dst, ip_cidr, port) tuples are rejected.
func ValidateIPRows(rows []CIDRRecord) error {
	for i := range rows {
		if rows[i].Net == nil || rows[i].Selector == nil {
			return fmt.Errorf("row %d is not parsed", i+1)
		}
	}

	seenPair := make(map[string]int, len(rows))
	for i, row := range rows {
		key := duplicateRowKey(row)
		if prev, ok := seenPair[key]; ok {
			src, dst := duplicateTupleSrcDst(row)
			return fmt.Errorf(
				"duplicate (src,dst,ip_cidr,port) found at rows %d and %d: (%s,%s,%s,%d)",
				prev, i+1, src, dst, row.Net.String(), row.Port,
			)
		}
		seenPair[key] = i + 1
	}

	for i := 0; i < len(rows); i++ {
		if !networkContains(rows[i].Net, rows[i].Selector) {
			return fmt.Errorf("ip selector %s is outside ip_cidr %s (row %d)", rows[i].Selector.String(), rows[i].Net.String(), i+1)
		}
	}

	return nil
}

func duplicateRowKey(row CIDRRecord) string {
	src, dst := duplicateTupleSrcDst(row)
	return fmt.Sprintf("%s|%s|%s|%d", row.Net.String(), src, dst, row.Port)
}

func networkContains(outer, inner *net.IPNet) bool {
	if outer == nil || inner == nil {
		return false
	}
	innerStart, innerEnd, ok := netutil.IPRange(inner)
	if !ok {
		return false
	}
	return outer.Contains(innerStart) && outer.Contains(innerEnd)
}

func duplicateTupleSrcDst(row CIDRRecord) (src string, dst string) {
	src = row.SrcIP
	dst = row.DstIP
	if src == "" && row.Selector != nil {
		src = row.Selector.String()
	}
	if dst == "" && row.Selector != nil {
		dst = row.Selector.String()
	}
	return src, dst
}
