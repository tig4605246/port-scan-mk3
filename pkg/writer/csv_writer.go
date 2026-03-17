package writer

import (
	"encoding/csv"
	"io"
	"strconv"
)

// Record is one scan output row written to CSV.
type Record struct {
	IP                string
	IPCidr            string
	Port              int
	Status            string
	ResponseMS        int64
	FabName           string
	CIDR              string
	CIDRName          string
	ServiceLabel      string
	Decision          string
	PolicyID          string
	Reason            string
	ExecutionKey      string
	SrcIP             string
	SrcNetworkSegment string
}

// ColumnDef maps a header name to a Record field extractor.
type ColumnDef struct {
	Name    string
	Extract func(Record) string
}

// Columns defines the CSV output contract as a single source of truth.
var Columns = []ColumnDef{
	{"ip", func(r Record) string { return r.IP }},
	{"ip_cidr", func(r Record) string {
		if r.IPCidr != "" {
			return r.IPCidr
		}
		return r.CIDR
	}},
	{"port", func(r Record) string { return strconv.Itoa(r.Port) }},
	{"status", func(r Record) string { return r.Status }},
	{"response_time_ms", func(r Record) string { return strconv.FormatInt(r.ResponseMS, 10) }},
	{"fab_name", func(r Record) string { return r.FabName }},
	{"cidr_name", func(r Record) string { return r.CIDRName }},
	{"service_label", func(r Record) string { return r.ServiceLabel }},
	{"decision", func(r Record) string { return r.Decision }},
	{"policy_id", func(r Record) string { return r.PolicyID }},
	{"reason", func(r Record) string { return r.Reason }},
	{"execution_key", func(r Record) string { return r.ExecutionKey }},
	{"src_ip", func(r Record) string { return r.SrcIP }},
	{"src_network_segment", func(r Record) string { return r.SrcNetworkSegment }},
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
	row := make([]string, len(Columns))
	for i, col := range Columns {
		row[i] = col.Extract(r)
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
		header := make([]string, len(Columns))
		for i, col := range Columns {
			header[i] = col.Name
		}
		if err := cw.w.Write(header); err != nil {
			return err
		}
		cw.wroteHeader = true
		cw.w.Flush()
		return cw.w.Error()
	}
	return nil
}
