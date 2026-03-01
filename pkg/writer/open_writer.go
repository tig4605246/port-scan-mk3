package writer

type OpenOnlyWriter struct {
	inner *CSVWriter
}

func NewOpenOnlyWriter(inner *CSVWriter) *OpenOnlyWriter {
	return &OpenOnlyWriter{inner: inner}
}

func (w *OpenOnlyWriter) Write(r Record) error {
	if w == nil || w.inner == nil {
		return nil
	}
	if r.Status != "open" {
		return nil
	}
	return w.inner.Write(r)
}

func (w *OpenOnlyWriter) WriteHeader() error {
	if w == nil || w.inner == nil {
		return nil
	}
	return w.inner.WriteHeader()
}
