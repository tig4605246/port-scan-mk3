package input

import (
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"strings"
)

func LoadCIDRs(r io.Reader) ([]CIDRRecord, error) {
	cr := csv.NewReader(r)
	rows, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("cidr csv must include header and at least one row")
	}
	header := rows[0]
	if len(header) < 3 ||
		strings.TrimSpace(header[0]) != "fab_name" ||
		strings.TrimSpace(header[1]) != "cidr" ||
		strings.TrimSpace(header[2]) != "cidr_name" {
		return nil, fmt.Errorf("invalid cidr csv header")
	}

	out := make([]CIDRRecord, 0, len(rows)-1)
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 3 {
			return nil, fmt.Errorf("invalid cidr row %d", i+1)
		}
		cidr := strings.TrimSpace(row[1])
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, fmt.Errorf("invalid cidr %q: %w", cidr, err)
		}
		out = append(out, CIDRRecord{
			FabName:  strings.TrimSpace(row[0]),
			CIDR:     cidr,
			CIDRName: strings.TrimSpace(row[2]),
			Net:      ipNet,
		})
	}
	if err := ValidateNoOverlap(out); err != nil {
		return nil, err
	}
	return out, nil
}
