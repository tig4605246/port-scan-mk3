package writer

import (
	"encoding/csv"
	"io"
	"strconv"
)

type Record struct {
	IP         string
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
	if !cw.wroteHeader {
		if err := cw.w.Write([]string{"ip", "port", "status", "response_time_ms", "fab_name", "cidr", "cidr_name"}); err != nil {
			return err
		}
		cw.wroteHeader = true
	}

	row := []string{
		r.IP,
		strconv.Itoa(r.Port),
		r.Status,
		strconv.FormatInt(r.ResponseMS, 10),
		r.FabName,
		r.CIDR,
		r.CIDRName,
	}
	if err := cw.w.Write(row); err != nil {
		return err
	}
	cw.w.Flush()
	return cw.w.Error()
}
