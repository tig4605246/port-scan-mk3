package input

import (
	"strings"
	"testing"
)

func TestLoadCIDRs_DetectOverlap(t *testing.T) {
	rows := "fab_name,cidr,cidr_name\nfab1,10.0.0.0/8,a\nfab2,10.1.0.0/16,b\n"
	_, err := LoadCIDRs(strings.NewReader(rows))
	if err == nil {
		t.Fatal("expected overlap error")
	}
}
