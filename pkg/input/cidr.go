package input

import (
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"strings"
)

// LoadCIDRs loads CIDR records using default required column names
// `ip` and `ip_cidr`.
func LoadCIDRs(r io.Reader) ([]CIDRRecord, error) {
	return LoadCIDRsWithColumns(r, "ip", "ip_cidr")
}

// LoadCIDRsWithColumns loads CIDR records from CSV with caller-provided
// case-sensitive column names for selector (`ipCol`) and boundary CIDR (`ipCidrCol`).
func LoadCIDRsWithColumns(r io.Reader, ipCol, ipCidrCol string) ([]CIDRRecord, error) {
	cr := csv.NewReader(r)
	rows, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("cidr csv must include header and at least one row")
	}

	ipCol = strings.TrimSpace(ipCol)
	ipCidrCol = strings.TrimSpace(ipCidrCol)
	if ipCol == "" || ipCidrCol == "" {
		return nil, fmt.Errorf("ip and ip_cidr column names must be non-empty")
	}

	header := normalizeHeader(rows[0])
	ipIdx := headerIndex(header, ipCol)
	if ipIdx < 0 {
		return nil, fmt.Errorf("cidr csv missing required ip column %q", ipCol)
	}
	ipCidrIdx := headerIndex(header, ipCidrCol)
	if ipCidrIdx < 0 {
		return nil, fmt.Errorf("cidr csv missing required ip_cidr column %q", ipCidrCol)
	}
	fabIdx := headerIndex(header, "fab_name")
	cidrNameIdx := headerIndex(header, "cidr_name")

	out := make([]CIDRRecord, 0, len(rows)-1)
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) <= max(ipIdx, ipCidrIdx) {
			return nil, fmt.Errorf("invalid cidr row %d", i+1)
		}

		rec := CIDRRecord{
			IPRaw:     strings.TrimSpace(row[ipIdx]),
			IPCidrRaw: strings.TrimSpace(row[ipCidrIdx]),
			RowNumber: i + 1,
			IPColName: ipCol,
			IPCidrCol: ipCidrCol,
		}
		if fabIdx >= 0 && fabIdx < len(row) {
			rec.FabName = strings.TrimSpace(row[fabIdx])
		}
		if cidrNameIdx >= 0 && cidrNameIdx < len(row) {
			rec.CIDRName = strings.TrimSpace(row[cidrNameIdx])
		}
		if err := rec.Parse(); err != nil {
			return nil, fmt.Errorf("invalid cidr row %d: %w", i+1, err)
		}
		out = append(out, rec)
	}
	if err := ValidateIPRows(out); err != nil {
		return nil, err
	}
	return out, nil
}

// Parse validates and normalizes a CIDRRecord into parsed selector/network forms.
func (r *CIDRRecord) Parse() error {
	if r == nil {
		return fmt.Errorf("nil cidr record")
	}
	if strings.TrimSpace(r.IPRaw) == "" {
		return fmt.Errorf("empty ip")
	}
	if strings.TrimSpace(r.IPCidrRaw) == "" {
		return fmt.Errorf("empty ip_cidr")
	}

	_, ipCidrNet, err := net.ParseCIDR(strings.TrimSpace(r.IPCidrRaw))
	if err != nil {
		return fmt.Errorf("invalid ip_cidr %q: %w", r.IPCidrRaw, err)
	}
	if ipCidrNet.IP.To4() == nil {
		return fmt.Errorf("only ipv4 ip_cidr is supported: %q", r.IPCidrRaw)
	}
	r.Net = ipCidrNet
	r.CIDR = ipCidrNet.String()
	r.IPCidrRaw = strings.TrimSpace(r.IPCidrRaw)

	selector, err := parseSelector(strings.TrimSpace(r.IPRaw))
	if err != nil {
		return fmt.Errorf("invalid ip %q: %w", r.IPRaw, err)
	}
	r.Selector = selector
	r.IPRaw = strings.TrimSpace(r.IPRaw)
	return nil
}

func parseSelector(raw string) (*net.IPNet, error) {
	if ip := net.ParseIP(raw); ip != nil {
		v4 := ip.To4()
		if v4 == nil {
			return nil, fmt.Errorf("only ipv4 is supported")
		}
		return &net.IPNet{
			IP:   v4,
			Mask: net.CIDRMask(32, 32),
		}, nil
	}
	_, sel, err := net.ParseCIDR(raw)
	if err != nil {
		return nil, err
	}
	if sel.IP.To4() == nil {
		return nil, fmt.Errorf("only ipv4 is supported")
	}
	return sel, nil
}

func normalizeHeader(header []string) []string {
	out := make([]string, len(header))
	for i, h := range header {
		out[i] = strings.TrimSpace(h)
	}
	return out
}

func headerIndex(header []string, name string) int {
	for i, h := range header {
		if h == name {
			return i
		}
	}
	return -1
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
