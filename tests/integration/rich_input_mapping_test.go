package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xuxiping/port-scan-mk3/pkg/input"
)

func TestRichInputMapping_WhenDuplicateRowsExist_ExecutionKeysAreDeduplicated(t *testing.T) {
	path := filepath.Join("testdata", "rich_input", "dedup_context.csv")
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	rows, err := input.LoadCIDRsWithColumns(f, "ip", "ip_cidr")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	keys := map[string]struct{}{}
	validRows := 0
	for _, r := range rows {
		if !r.IsValid {
			continue
		}
		validRows++
		keys[r.ExecutionKey] = struct{}{}
	}
	if validRows != 3 {
		t.Fatalf("unexpected valid rows: %d", validRows)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 unique execution keys, got %d", len(keys))
	}
}
