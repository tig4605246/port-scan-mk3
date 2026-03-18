package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	speedctrlkit "github.com/xuxiping/port-scan-mk3/internal/testkit/speedcontrol"
)

func TestGenerateSpeedControlReport_WritesMarkdownHTMLAndRawMetrics(t *testing.T) {
	outDir := t.TempDir()
	runs := []speedctrlkit.ScenarioRun{
		{
			Name: "G1_manual_pause",
			Verdict: speedctrlkit.ScenarioVerdict{
				Pass:        true,
				Expected:    "expected_tps=0.00",
				Observed:    "observed_tps=0.00",
				Attribution: "within_expected_range",
				Explanation: "all checks satisfied",
			},
		},
		{
			Name: "C1_single_cidr_steady_rate",
			Verdict: speedctrlkit.ScenarioVerdict{
				Pass:        false,
				Expected:    "expected_tps=20.00",
				Observed:    "observed_tps=3.10",
				Attribution: "throughput_lower_than_expected",
				Explanation: "checks failed",
			},
		},
	}

	if err := GenerateSpeedControlReport(outDir, runs); err != nil {
		t.Fatalf("GenerateSpeedControlReport returned err: %v", err)
	}

	for _, name := range []string{"report.md", "report.html", "raw_metrics.json"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}

	md, err := os.ReadFile(filepath.Join(outDir, "report.md"))
	if err != nil {
		t.Fatal(err)
	}
	mdText := string(md)
	for _, needle := range []string{"Expected", "Observed", "Verdict", "Explanation"} {
		if !strings.Contains(mdText, needle) {
			t.Fatalf("report.md should contain %q, got:\n%s", needle, mdText)
		}
	}
	if !strings.Contains(mdText, "G1_manual_pause") || !strings.Contains(mdText, "C1_single_cidr_steady_rate") {
		t.Fatalf("report.md missing scenario names:\n%s", mdText)
	}

	raw, err := os.ReadFile(filepath.Join(outDir, "raw_metrics.json"))
	if err != nil {
		t.Fatal(err)
	}
	rawText := string(raw)
	if !strings.Contains(rawText, "G1_manual_pause") || !strings.Contains(rawText, "C1_single_cidr_steady_rate") {
		t.Fatalf("raw_metrics.json missing scenario names:\n%s", rawText)
	}
}
