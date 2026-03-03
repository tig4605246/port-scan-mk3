package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuxiping/port-scan-mk3/pkg/input"
)

func TestRichInputRejection_WhenAllRowsInvalid_ReturnsNoUsableInputError(t *testing.T) {
	path := filepath.Join("testdata", "rich_input", "invalid_rows.csv")
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	_, err = input.LoadCIDRsWithColumns(f, "ip", "ip_cidr")
	if err == nil || !strings.Contains(err.Error(), "no usable input rows") {
		t.Fatalf("expected no usable input rows error, got %v", err)
	}
}
