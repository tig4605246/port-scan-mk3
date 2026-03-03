package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xuxiping/port-scan-mk3/pkg/input"
)

func TestRichInputParse_WhenMixedRowsProvided_ParsesValidAndKeepsInvalidResults(t *testing.T) {
	path := filepath.Join("testdata", "rich_input", "valid_mixed.csv")
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	rows, err := input.LoadCIDRsWithColumns(f, "ip", "ip_cidr")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("unexpected row result count: %d", len(rows))
	}

	valid := 0
	invalid := 0
	for _, r := range rows {
		if r.IsValid {
			valid++
		} else {
			invalid++
		}
	}
	if valid != 2 || invalid != 1 {
		t.Fatalf("unexpected valid/invalid counts: valid=%d invalid=%d", valid, invalid)
	}
}
