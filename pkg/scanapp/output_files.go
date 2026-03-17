package scanapp

import (
	"os"

	"github.com/xuxiping/port-scan-mk3/pkg/writer"
)

type batchOutputs struct {
	scanFile       *os.File
	openOnlyFile   *os.File
	scanWriter     *writer.CSVWriter
	openOnlyWriter *writer.OpenOnlyWriter
}

func openBatchOutputs(scanPath, openPath string) (*batchOutputs, error) {
	scanFile, err := os.Create(scanPath)
	if err != nil {
		return nil, err
	}

	scanWriter := writer.NewCSVWriter(scanFile)
	if err := scanWriter.WriteHeader(); err != nil {
		_ = scanFile.Close()
		return nil, err
	}

	openOnlyFile, err := os.Create(openPath)
	if err != nil {
		_ = scanFile.Close()
		return nil, err
	}

	openOnlyWriter := writer.NewOpenOnlyWriter(writer.NewCSVWriter(openOnlyFile))
	if err := openOnlyWriter.WriteHeader(); err != nil {
		_ = openOnlyFile.Close()
		_ = scanFile.Close()
		return nil, err
	}

	return &batchOutputs{
		scanFile:       scanFile,
		openOnlyFile:   openOnlyFile,
		scanWriter:     scanWriter,
		openOnlyWriter: openOnlyWriter,
	}, nil
}

func (b *batchOutputs) Close() error {
	if b == nil {
		return nil
	}
	var firstErr error
	if b.openOnlyFile != nil {
		if err := b.openOnlyFile.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if b.scanFile != nil {
		if err := b.scanFile.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
