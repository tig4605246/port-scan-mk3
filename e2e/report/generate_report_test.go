package report

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateReport_WritesHTMLAndText(t *testing.T) {
	outDir := t.TempDir()
	in := Summary{Total: 4, Open: 2, Closed: 1, Timeout: 1}

	if err := Generate(outDir, in); err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "report.html")); err != nil {
		t.Fatalf("missing html report: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "report.txt")); err != nil {
		t.Fatalf("missing text report: %v", err)
	}
}
