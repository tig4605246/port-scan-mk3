package scanapp

import (
	"path/filepath"
	"testing"

	"github.com/xuxiping/port-scan-mk3/pkg/config"
)

func TestResumePath_WhenConfigVariantsProvided_ResolvesExpectedPath(t *testing.T) {
	if got := resumePath(config.Config{Resume: "cfg.json"}, RunOptions{ResumeStatePath: "opt.json"}); got != "opt.json" {
		t.Fatalf("unexpected resume path with option override: %s", got)
	}
	if got := resumePath(config.Config{Resume: "cfg.json"}, RunOptions{}); got != "cfg.json" {
		t.Fatalf("unexpected resume path from cfg: %s", got)
	}
	cfg := config.Config{Output: filepath.Join("/tmp", "scan_results.csv")}
	if got := resumePath(cfg, RunOptions{}); got != filepath.Join("/tmp", defaultResumeStateFile) {
		t.Fatalf("unexpected fallback resume path: %s", got)
	}
}
