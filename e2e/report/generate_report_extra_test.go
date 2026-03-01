package report

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateReport_InvalidOutDir(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "not-a-dir")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Generate(file, Summary{}); err == nil {
		t.Fatal("expected error")
	}
}
