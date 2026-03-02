package writer

// OpenOnlyWriter writes only rows whose status is `open`.
type OpenOnlyWriter struct {
	inner *CSVWriter
}

// NewOpenOnlyWriter creates a filter writer for open-only output.
func NewOpenOnlyWriter(inner *CSVWriter) *OpenOnlyWriter {
	return &OpenOnlyWriter{inner: inner}
}

// Write forwards only `open` status records.
func (w *OpenOnlyWriter) Write(r Record) error {
	if w == nil || w.inner == nil {
		return nil
	}
	if r.Status != "open" {
		return nil
	}
	return w.inner.Write(r)
}

// WriteHeader writes the output header through the inner writer.
func (w *OpenOnlyWriter) WriteHeader() error {
	if w == nil || w.inner == nil {
		return nil
	}
	return w.inner.WriteHeader()
}
