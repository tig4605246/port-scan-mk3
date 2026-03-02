package scanapp

import (
	"path/filepath"

	"github.com/xuxiping/port-scan-mk3/pkg/config"
)

func resumePath(cfg config.Config, opts RunOptions) string {
	if opts.ResumeStatePath != "" {
		return opts.ResumeStatePath
	}
	if cfg.Resume != "" {
		return cfg.Resume
	}
	outputDir := filepath.Dir(cfg.Output)
	if outputDir == "." || outputDir == "" {
		return defaultResumeStateFile
	}
	return filepath.Join(outputDir, defaultResumeStateFile)
}
