package writer

import (
	"encoding/csv"
	"io"
	"strconv"
)

// Record is one scan output row written to CSV.
type Record struct {
	IP         string
	IPCidr     string
	Port       int
	Status     string
	ResponseMS int64
	FabName    string
	CIDR       string
	CIDRName   string
}

// CSVWriter writes scan result rows with the fixed contract header.
type CSVWriter struct {
	w           *csv.Writer
	wroteHeader bool
}

// NewCSVWriter creates a CSV writer for scan result output.
func NewCSVWriter(out io.Writer) *CSVWriter {
	return &CSVWriter{w: csv.NewWriter(out)}
}

// Write appends a single record and writes header first if needed.
func (cw *CSVWriter) Write(r Record) error {
	if err := cw.WriteHeader(); err != nil {
		return err
	}

	cidr := r.IPCidr
	if cidr == "" {
		cidr = r.CIDR
	}
	row := []string{
		r.IP,
		cidr,
		strconv.Itoa(r.Port),
		r.Status,
		strconv.FormatInt(r.ResponseMS, 10),
		r.FabName,
		r.CIDRName,
	}
	if err := cw.w.Write(row); err != nil {
		return err
	}
	cw.w.Flush()
	return cw.w.Error()
}

// WriteHeader writes the fixed result header once.
func (cw *CSVWriter) WriteHeader() error {
	if !cw.wroteHeader {
		if err := cw.w.Write([]string{"ip", "ip_cidr", "port", "status", "response_time_ms", "fab_name", "cidr_name"}); err != nil {
			return err
		}
		cw.wroteHeader = true
		cw.w.Flush()
		return cw.w.Error()
	}
	return nil
}
