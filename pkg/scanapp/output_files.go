package scanapp

import (
	"fmt"
	"os"

	"github.com/xuxiping/port-scan-mk3/pkg/writer"
)

type batchOutputs struct {
	scanFile       *os.File
	openOnlyFile   *os.File
	scanWriter     *writer.CSVWriter
	openOnlyWriter *writer.OpenOnlyWriter
	scanFinalPath  string
	openFinalPath  string
}

func openBatchOutputs(scanPath, openPath string) (*batchOutputs, error) {
	scanTmpPath := scanPath + ".tmp"
	scanFile, err := os.Create(scanTmpPath)
	if err != nil {
		return nil, err
	}

	scanWriter := writer.NewCSVWriter(scanFile)
	if err := scanWriter.WriteHeader(); err != nil {
		if closeErr := scanFile.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "failed to close scan file: %v\n", closeErr)
		}
		return nil, err
	}

	openTmpPath := openPath + ".tmp"
	openOnlyFile, err := os.Create(openTmpPath)
	if err != nil {
		if closeErr := scanFile.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "failed to close scan file: %v\n", closeErr)
		}
		return nil, err
	}

	openOnlyWriter := writer.NewOpenOnlyWriter(writer.NewCSVWriter(openOnlyFile))
	if err := openOnlyWriter.WriteHeader(); err != nil {
		if closeErr := openOnlyFile.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "failed to close open-only file: %v\n", closeErr)
		}
		if closeErr := scanFile.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "failed to close scan file: %v\n", closeErr)
		}
		return nil, err
	}

	return &batchOutputs{
		scanFile:       scanFile,
		openOnlyFile:   openOnlyFile,
		scanWriter:     scanWriter,
		openOnlyWriter: openOnlyWriter,
		scanFinalPath:  scanPath,
		openFinalPath:  openPath,
	}, nil
}

func (b *batchOutputs) Finalize(success bool) error {
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
	if firstErr != nil {
		return firstErr
	}

	if success {
		if err := os.Rename(b.scanFinalPath+".tmp", b.scanFinalPath); err != nil {
			return err
		}
		if err := os.Rename(b.openFinalPath+".tmp", b.openFinalPath); err != nil {
			return err
		}
	}
	return nil
}
