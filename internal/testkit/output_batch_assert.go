package testkit

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
)

var batchOutputNameRE = regexp.MustCompile(`^(scan_results|opened_results)-([0-9]{8}T[0-9]{6}Z)(?:-([1-9][0-9]*))?\.csv$`)

// BatchOutputName is the parsed structure of a scan batch output filename.
type BatchOutputName struct {
	Prefix    string
	Timestamp string
	Sequence  int
}

// ParseBatchOutputName parses a timestamped output filename.
func ParseBatchOutputName(path string) (BatchOutputName, error) {
	name := filepath.Base(path)
	m := batchOutputNameRE.FindStringSubmatch(name)
	if m == nil {
		return BatchOutputName{}, fmt.Errorf("invalid batch output filename: %s", name)
	}

	out := BatchOutputName{Prefix: m[1], Timestamp: m[2]}
	if m[3] == "" {
		return out, nil
	}
	seq, err := strconv.Atoi(m[3])
	if err != nil || seq <= 0 {
		return BatchOutputName{}, fmt.Errorf("invalid batch sequence in filename: %s", name)
	}
	out.Sequence = seq
	return out, nil
}

// AssertBatchPair ensures main and open-only outputs belong to the same batch.
func AssertBatchPair(scanResultsPath, openedResultsPath string) error {
	scan, err := ParseBatchOutputName(scanResultsPath)
	if err != nil {
		return err
	}
	open, err := ParseBatchOutputName(openedResultsPath)
	if err != nil {
		return err
	}
	if scan.Prefix != "scan_results" {
		return fmt.Errorf("expected scan_results prefix, got %s", scan.Prefix)
	}
	if open.Prefix != "opened_results" {
		return fmt.Errorf("expected opened_results prefix, got %s", open.Prefix)
	}
	if scan.Timestamp != open.Timestamp {
		return fmt.Errorf("batch timestamp mismatch: %s != %s", scan.Timestamp, open.Timestamp)
	}
	if scan.Sequence != open.Sequence {
		return fmt.Errorf("batch sequence mismatch: %d != %d", scan.Sequence, open.Sequence)
	}
	return nil
}
