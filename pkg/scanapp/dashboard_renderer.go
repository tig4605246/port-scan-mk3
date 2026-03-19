package scanapp

import (
	"fmt"
	"io"
	"time"
)

const dashboardRefreshSequence = "\x1b[2J\x1b[H"

type dashboardRenderer struct{}

func (r dashboardRenderer) Render(w io.Writer, snap dashboardSnapshot) error {
	_, err := fmt.Fprintf(w,
		"%sProgress: %d/%d (%.1f%%)\n"+
			"Current CIDR: %s\n"+
			"Bucket: %s\n"+
			"Dispatch/s: %.2f\n"+
			"Results/s: %.2f\n"+
			"Controller: %s\n"+
			"API Pressure: %d%%\n"+
			"Last Update: %s\n"+
			"Health: %s\n",
		dashboardRefreshSequence,
		snap.ScannedTasks,
		snap.TotalTasks,
		snap.Percent,
		dashboardValueOrFallback(snap.CurrentCIDR),
		dashboardValueOrFallback(snap.BucketStatus),
		snap.DispatchPerSecond,
		snap.ResultsPerSecond,
		dashboardValueOrFallback(snap.ControllerStatus),
		snap.PressurePercent,
		dashboardLastUpdateText(snap),
		dashboardValueOrFallback(snap.APIHealthText),
	)
	return err
}

func dashboardValueOrFallback(value string) string {
	if value == "" {
		return "-"
	}
	return value
}

func dashboardLastUpdateText(snap dashboardSnapshot) string {
	latest := snap.LastPressureUpdateAt
	if snap.LastPressureFailureAt.After(latest) {
		latest = snap.LastPressureFailureAt
	}
	if latest.IsZero() {
		return "-"
	}
	return latest.UTC().Format(time.RFC3339)
}
