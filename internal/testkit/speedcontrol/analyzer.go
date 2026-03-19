package speedcontrol

import (
	"fmt"
	"math"
	"sort"
)

const defaultTolerance = 0.15

func Analyze(events []Event, expectation RuleExpectation) ScenarioVerdict {
	expectedTPS := resolveExpectedTPS(expectation)
	tolerance := expectation.Tolerance
	if tolerance <= 0 {
		tolerance = defaultTolerance
	}

	enqueueTS := enqueueTimestamps(events)
	observedTPS := estimateTPS(enqueueTS)
	pauseViolations := countPauseWindowViolations(enqueueTS, expectation.PauseWindows)
	burstEvents := estimateImmediateBurst(enqueueTS, expectation.BurstMaxGapNS)

	tpsPass := true
	if expectedTPS > 0 {
		if observedTPS <= 0 {
			tpsPass = false
		} else {
			diffRatio := math.Abs(observedTPS-expectedTPS) / expectedTPS
			tpsPass = diffRatio <= tolerance
		}
	}

	pausePass := true
	if expectation.RequirePauseBlocking {
		pausePass = pauseViolations == 0
	}

	burstPass := true
	if expectation.MinImmediateBurst > 0 {
		burstPass = burstEvents >= expectation.MinImmediateBurst
	}

	pass := tpsPass && pausePass && burstPass
	attribution := classifyAttribution(expectedTPS, observedTPS, pauseViolations, burstPass)
	expectedText := fmt.Sprintf(
		"expected_tps=%.2f tolerance=±%.1f%% require_pause_blocking=%t min_immediate_burst=%d",
		expectedTPS,
		tolerance*100,
		expectation.RequirePauseBlocking,
		expectation.MinImmediateBurst,
	)
	observedText := fmt.Sprintf(
		"observed_tps=%.2f task_enqueued=%d pause_violations=%d burst_events=%d",
		observedTPS,
		len(enqueueTS),
		pauseViolations,
		burstEvents,
	)

	explanation := "all checks satisfied"
	if !pass {
		explanation = fmt.Sprintf("checks failed: tps_pass=%t pause_pass=%t burst_pass=%t attribution=%s", tpsPass, pausePass, burstPass, attribution)
	}

	return ScenarioVerdict{
		Name:        expectation.Name,
		Pass:        pass,
		Expected:    expectedText,
		Observed:    observedText,
		Attribution: attribution,
		Explanation: explanation,
		Metrics: ScenarioMetrics{
			TaskEnqueued:    len(enqueueTS),
			ObservedTPS:     observedTPS,
			ExpectedTPS:     expectedTPS,
			Tolerance:       tolerance,
			PauseViolations: pauseViolations,
			BurstEvents:     burstEvents,
		},
	}
}

func resolveExpectedTPS(expectation RuleExpectation) float64 {
	if expectation.ExpectedTPS > 0 {
		return expectation.ExpectedTPS
	}

	candidates := make([]float64, 0, 3)
	if expectation.BucketRate > 0 {
		candidates = append(candidates, expectation.BucketRate)
	}
	if expectation.DelaySeconds > 0 {
		candidates = append(candidates, 1.0/expectation.DelaySeconds)
	}
	if expectation.Workers > 0 && expectation.AvgTaskSeconds > 0 {
		candidates = append(candidates, float64(expectation.Workers)/expectation.AvgTaskSeconds)
	}
	if len(candidates) == 0 {
		return 0
	}
	min := candidates[0]
	for _, c := range candidates[1:] {
		if c < min {
			min = c
		}
	}
	return min
}

func enqueueTimestamps(events []Event) []int64 {
	out := make([]int64, 0, len(events))
	for _, e := range events {
		if e.Kind == EventTaskEnqueued {
			out = append(out, e.TimestampNS)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func estimateTPS(timestamps []int64) float64 {
	if len(timestamps) < 2 {
		return 0
	}
	span := float64(timestamps[len(timestamps)-1]-timestamps[0]) / 1e9
	if span <= 0 {
		return 0
	}
	return float64(len(timestamps)-1) / span
}

func countPauseWindowViolations(enqueueTS []int64, windows []PauseWindow) int {
	if len(enqueueTS) == 0 || len(windows) == 0 {
		return 0
	}
	violations := 0
	for _, ts := range enqueueTS {
		for _, w := range windows {
			if ts >= w.StartNS && ts < w.EndNS {
				violations++
				break
			}
		}
	}
	return violations
}

func estimateImmediateBurst(enqueueTS []int64, burstMaxGapNS int64) int {
	if len(enqueueTS) == 0 {
		return 0
	}
	if burstMaxGapNS <= 0 {
		burstMaxGapNS = int64((20 * 1e6)) // 20ms
	}

	burst := 1
	for i := 1; i < len(enqueueTS); i++ {
		gap := enqueueTS[i] - enqueueTS[i-1]
		if gap > burstMaxGapNS {
			break
		}
		burst++
	}
	return burst
}

func classifyAttribution(expectedTPS, observedTPS float64, pauseViolations int, burstPass bool) string {
	if pauseViolations > 0 {
		return "global_gate_not_blocking_dispatch"
	}
	if !burstPass {
		return "burst_not_observed"
	}
	if expectedTPS <= 0 {
		return "insufficient_expected_baseline"
	}
	if observedTPS < expectedTPS {
		return "throughput_lower_than_expected"
	}
	if observedTPS > expectedTPS {
		return "throughput_higher_than_expected"
	}
	return "within_expected_range"
}
