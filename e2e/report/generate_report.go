package report

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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

func SummarizeCSV(path string) (Summary, error) {
	f, err := os.Open(path)
	if err != nil {
		return Summary{}, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil {
		return Summary{}, err
	}
	if len(rows) < 2 {
		return Summary{}, fmt.Errorf("scan result csv has no data rows")
	}

	s := Summary{}
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 3 {
			return Summary{}, io.ErrUnexpectedEOF
		}
		status := strings.TrimSpace(row[2])
		s.Total++
		switch status {
		case "open":
			s.Open++
		case "close":
			s.Closed++
		case "close(timeout)":
			s.Timeout++
		default:
			return Summary{}, fmt.Errorf("unsupported status in result csv: %s", status)
		}
	}
	return s, nil
}
