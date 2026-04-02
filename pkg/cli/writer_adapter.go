package cli

import (
	"github.com/xuxiping/port-scan-mk3/pkg/scanapp"
	"github.com/xuxiping/port-scan-mk3/pkg/writer"
)

// RecordWriterAdapter wraps a writer.CSVWriter to implement scanapp.RecordWriter.
// This lives in CLI layer (not domain) per SOLID - it bridges domain interface to concrete writer.
type RecordWriterAdapter struct {
	w *writer.CSVWriter
}

// NewRecordWriterAdapter creates an adapter from *writer.CSVWriter to scanapp.RecordWriter.
func NewRecordWriterAdapter(csv *writer.CSVWriter) *RecordWriterAdapter {
	return &RecordWriterAdapter{w: csv}
}

func (a *RecordWriterAdapter) Write(record writer.Record) error {
	return a.w.Write(record)
}

func (a *RecordWriterAdapter) WriteHeader() error {
	return a.w.WriteHeader()
}

// Compile-time interface check
var _ scanapp.RecordWriter = (*RecordWriterAdapter)(nil)

// OpenOnlyRecordWriterAdapter wraps a writer.OpenOnlyWriter to implement scanapp.RecordWriter.
// The OpenOnlyWriter filters records to only include "open" status entries.
type OpenOnlyRecordWriterAdapter struct {
	w *writer.OpenOnlyWriter
}

// NewOpenOnlyRecordWriterAdapter creates an adapter from *writer.OpenOnlyWriter to scanapp.RecordWriter.
func NewOpenOnlyRecordWriterAdapter(open *writer.OpenOnlyWriter) *OpenOnlyRecordWriterAdapter {
	return &OpenOnlyRecordWriterAdapter{w: open}
}

func (a *OpenOnlyRecordWriterAdapter) Write(record writer.Record) error {
	return a.w.Write(record)
}

func (a *OpenOnlyRecordWriterAdapter) WriteHeader() error {
	return a.w.WriteHeader()
}

// Compile-time interface check
var _ scanapp.RecordWriter = (*OpenOnlyRecordWriterAdapter)(nil)