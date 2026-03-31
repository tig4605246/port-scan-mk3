package task

import (
	"encoding/binary"
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/xuxiping/port-scan-mk3/pkg/netutil"
)

func ExpandIPSelectors(selectors []string) ([]string, error) {
	uniq := make(map[uint32]struct{})
	for _, raw := range selectors {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return nil, fmt.Errorf("empty selector")
		}
		if ip := net.ParseIP(raw); ip != nil {
			v4 := ip.To4()
			if v4 == nil {
				return nil, fmt.Errorf("only ipv4 is supported: %s", raw)
			}
			uniq[binary.BigEndian.Uint32(v4)] = struct{}{}
			continue
		}
		_, n, err := net.ParseCIDR(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid selector %q: %w", raw, err)
		}
		start, end, ok := netutil.IPRange(n)
		if !ok {
			return nil, fmt.Errorf("only ipv4 is supported: %s", raw)
		}
		startN := binary.BigEndian.Uint32(start.To4())
		endN := binary.BigEndian.Uint32(end.To4())
		for curr := startN; curr <= endN; curr++ {
			uniq[curr] = struct{}{}
			if curr == ^uint32(0) {
				break
			}
		}
	}

	keys := make([]uint32, 0, len(uniq))
	for n := range uniq {
		keys = append(keys, n)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	out := make([]string, 0, len(keys))
	for _, n := range keys {
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, n)
		out = append(out, ip.String())
	}
	return out, nil
}
