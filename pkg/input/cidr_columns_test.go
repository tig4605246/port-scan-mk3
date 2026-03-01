package input

import (
	"strings"
	"testing"
)

func TestLoadCIDRsWithColumns_NamedColumns(t *testing.T) {
	csv := "foo,source_ip,bar,source_cidr,cidr_name,fab_name\n" +
		"x,10.0.0.1,y,10.0.0.0/24,dc,fab-a\n"
	rows, err := LoadCIDRsWithColumns(strings.NewReader(csv), "source_ip", "source_cidr")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("unexpected rows len: %d", len(rows))
	}
	if rows[0].IPRaw != "10.0.0.1" || rows[0].IPCidrRaw != "10.0.0.0/24" {
		t.Fatalf("unexpected row: %#v", rows[0])
	}
}

func TestLoadCIDRsWithColumns_MissingColumns(t *testing.T) {
	csv := "a,b,c\n1,2,3\n"
	_, err := LoadCIDRsWithColumns(strings.NewReader(csv), "ip", "ip_cidr")
	if err == nil {
		t.Fatal("expected error for missing columns")
	}
}
