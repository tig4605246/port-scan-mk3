package report

import (
	"fmt"
	"os"
	"path/filepath"
)

type Summary struct {
	Total   int
	Open    int
	Closed  int
	Timeout int
}

func Generate(outDir string, s Summary) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	html := fmt.Sprintf(`<html><head><title>Port Scan E2E Report</title></head><body><h1>Port Scan E2E Report</h1><p>Total=%d Open=%d Closed=%d Timeout=%d</p></body></html>`, s.Total, s.Open, s.Closed, s.Timeout)
	txt := fmt.Sprintf("Port Scan E2E Report\nTotal=%d\nOpen=%d\nClosed=%d\nTimeout=%d\n", s.Total, s.Open, s.Closed, s.Timeout)

	if err := os.WriteFile(filepath.Join(outDir, "report.html"), []byte(html), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "report.txt"), []byte(txt), 0o644)
}
