package writer

import (
	"errors"
	"testing"
)

type badWriter struct{}

func (badWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestCSVWriter_WriteError(t *testing.T) {
	w := NewCSVWriter(badWriter{})
	err := w.Write(Record{IP: "1.1.1.1", Port: 80, Status: "open"})
	if err == nil {
		t.Fatal("expected error")
	}
}
