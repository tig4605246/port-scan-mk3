package writer

import (
	"encoding/csv"
	"io"
	"strconv"
)

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

type CSVWriter struct {
	w           *csv.Writer
	wroteHeader bool
}

func NewCSVWriter(out io.Writer) *CSVWriter {
	return &CSVWriter{w: csv.NewWriter(out)}
}

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
