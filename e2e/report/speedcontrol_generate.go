package report

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	speedctrlkit "github.com/xuxiping/port-scan-mk3/internal/testkit/speedcontrol"
)

//go:embed speedcontrol_template.html
var speedControlTemplateHTML string

type speedControlTemplateData struct {
	GeneratedAt string
	Summary     SpeedControlSummary
	Scenarios   []SpeedControlScenario
}

func GenerateSpeedControlReport(outDir string, runs []speedctrlkit.ScenarioRun) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	now := time.Now().UTC()
	full := toSpeedControlFullReport(now, runs)

	rawPath := filepath.Join(outDir, "raw_metrics.json")
	rawBytes, err := json.MarshalIndent(runs, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(rawPath, rawBytes, 0o644); err != nil {
		return err
	}

	mdPath := filepath.Join(outDir, "report.md")
	if err := os.WriteFile(mdPath, []byte(renderSpeedControlMarkdown(full)), 0o644); err != nil {
		return err
	}

	tmpl, err := template.New("speedcontrol").Parse(speedControlTemplateHTML)
	if err != nil {
		return err
	}
	var htmlOut bytes.Buffer
	if err := tmpl.Execute(&htmlOut, speedControlTemplateData{
		GeneratedAt: full.Timestamp.Format(time.RFC3339),
		Summary:     full.Summary,
		Scenarios:   full.Scenarios,
	}); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "report.html"), htmlOut.Bytes(), 0o644)
}

func toSpeedControlFullReport(ts time.Time, runs []speedctrlkit.ScenarioRun) SpeedControlFullReport {
	report := SpeedControlFullReport{
		Timestamp: ts,
		Summary: SpeedControlSummary{
			Total: len(runs),
		},
		Scenarios: make([]SpeedControlScenario, 0, len(runs)),
	}
	for _, run := range runs {
		item := SpeedControlScenario{
			Name:        run.Name,
			Pass:        run.Verdict.Pass,
			Expected:    run.Verdict.Expected,
			Observed:    run.Verdict.Observed,
			Attribution: run.Verdict.Attribution,
			Explanation: run.Verdict.Explanation,
		}
		if item.Pass {
			report.Summary.Pass++
		} else {
			report.Summary.Fail++
		}
		report.Scenarios = append(report.Scenarios, item)
	}
	return report
}

func renderSpeedControlMarkdown(report SpeedControlFullReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Speed Control Verification Report\n\n")
	fmt.Fprintf(&b, "- GeneratedAt: %s\n", report.Timestamp.Format(time.RFC3339))
	fmt.Fprintf(&b, "- Total: %d\n", report.Summary.Total)
	fmt.Fprintf(&b, "- Pass: %d\n", report.Summary.Pass)
	fmt.Fprintf(&b, "- Fail: %d\n\n", report.Summary.Fail)

	fmt.Fprintf(&b, "## Scenario Matrix\n\n")
	fmt.Fprintf(&b, "| Scenario | Verdict |\n")
	fmt.Fprintf(&b, "|---|---|\n")
	for _, sc := range report.Scenarios {
		verdict := "FAIL"
		if sc.Pass {
			verdict = "PASS"
		}
		fmt.Fprintf(&b, "| %s | %s |\n", sc.Name, verdict)
	}
	fmt.Fprintf(&b, "\n## Detailed Explanation\n\n")
	for _, sc := range report.Scenarios {
		verdict := "FAIL"
		if sc.Pass {
			verdict = "PASS"
		}
		fmt.Fprintf(&b, "### %s\n\n", sc.Name)
		fmt.Fprintf(&b, "- Verdict: %s\n", verdict)
		fmt.Fprintf(&b, "- Expected: %s\n", sc.Expected)
		fmt.Fprintf(&b, "- Observed: %s\n", sc.Observed)
		fmt.Fprintf(&b, "- Attribution: %s\n", sc.Attribution)
		fmt.Fprintf(&b, "- Explanation: %s\n\n", sc.Explanation)
	}
	return b.String()
}
