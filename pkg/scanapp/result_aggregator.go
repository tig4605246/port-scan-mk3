package scanapp

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/xuxiping/port-scan-mk3/pkg/speedctrl"
	"github.com/xuxiping/port-scan-mk3/pkg/writer"
)

type resultSummary struct {
	written      int
	openCount    int
	closeCount   int
	timeoutCount int
}

func writeScanRecord(csvWriter *writer.CSVWriter, openOnlyWriter *writer.OpenOnlyWriter, record writer.Record) error {
	if err := csvWriter.Write(record); err != nil {
		return err
	}
	return openOnlyWriter.Write(record)
}

func applyScanResult(runtimes []*chunkRuntime, res scanResult, summary *resultSummary) *resultSummary {
	if summary == nil {
		summary = &resultSummary{}
	}
	runtimes[res.chunkIdx].tracker.IncrementScanned()

	summary.written++
	switch {
	case strings.EqualFold(res.record.Status, "open"):
		summary.openCount++
	case strings.Contains(strings.ToLower(res.record.Status), "timeout"):
		summary.timeoutCount++
	default:
		summary.closeCount++
	}
	return summary
}

func emitScanResultEvents(stdout io.Writer, logger *scanLogger, ctrl *speedctrl.Controller, progressStep int, runtimes []*chunkRuntime, res scanResult, summary *resultSummary) {
	logger.eventf("scan_result", res.record.IP, res.record.Port, "scanned", statusErrorCause(res.record.Status), map[string]any{
		"status":           res.record.Status,
		"response_time_ms": res.record.ResponseMS,
		"cidr":             res.record.IPCidr,
	})
	if summary == nil || progressStep <= 0 || summary.written%progressStep != 0 {
		return
	}

	rt := runtimes[res.chunkIdx]
	cidr := rt.tracker.CIDR()
	scanned := rt.tracker.ScannedCount()
	total := rt.tracker.TotalCount()
	_, _ = fmt.Fprintf(stdout, "progress cidr=%s scanned=%d/%d paused=%t\n", cidr, scanned, total, ctrl.IsPaused())
	completionRate := 0.0
	if total > 0 {
		completionRate = float64(scanned) / float64(total)
	}
	logger.eventf("scan_progress", "", 0, "progress", "none", map[string]any{
		"cidr":            cidr,
		"scanned_count":   scanned,
		"total_count":     total,
		"completion_rate": completionRate,
		"paused":          ctrl.IsPaused(),
	})
}

func emitCompletionSummary(logger *scanLogger, summary resultSummary, startedAt time.Time, err error) {
	success := err == nil
	cause := "none"
	if err != nil {
		cause = errorCause(err)
	}
	logger.eventf("scan_completion", "", 0, "completion_summary", cause, map[string]any{
		"total_tasks":   summary.written,
		"open_count":    summary.openCount,
		"close_count":   summary.closeCount,
		"timeout_count": summary.timeoutCount,
		"duration_ms":   time.Since(startedAt).Milliseconds(),
		"success":       success,
	})
}
