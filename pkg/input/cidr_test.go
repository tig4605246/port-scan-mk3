package input

import (
	"strings"
	"testing"
)

func TestLoadCIDRs_WhenCIDRsOverlap_ReturnsNil(t *testing.T) {
	rows := "fab_name,ip,ip_cidr,cidr_name\n" +
		"fab1,10.0.0.1,10.0.0.0/8,a\n" +
		"fab2,10.1.0.1,10.1.0.0/16,b\n"
	_, err := LoadCIDRs(strings.NewReader(rows))
	if err != nil {
		t.Fatalf("expected overlap to be allowed, got %v", err)
	}
}
