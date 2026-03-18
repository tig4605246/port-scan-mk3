package speedcontrol

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/xuxiping/port-scan-mk3/pkg/ratelimit"
	"github.com/xuxiping/port-scan-mk3/pkg/speedctrl"
)

type scenarioConfig struct {
	name           string
	tasksByCIDR    map[string]int
	bucketRate     int
	bucketCapacity int
	delay          time.Duration
	expectation    RuleExpectation
	hook           func(dispatched int, ctrl *speedctrl.Controller, windows *[]PauseWindow)
}

func RunScenarioMatrix() ([]ScenarioRun, error) {
	scenarios := scenarioDefinitions()
	out := make([]ScenarioRun, 0, len(scenarios))
	for _, sc := range scenarios {
		run, err := runScenario(sc)
		if err != nil {
			return nil, fmt.Errorf("run scenario %s: %w", sc.name, err)
		}
		out = append(out, run)
	}
	return out, nil
}

func scenarioDefinitions() []scenarioConfig {
	manualPauseHook := func(triggerAt int, pauseDuration time.Duration) func(int, *speedctrl.Controller, *[]PauseWindow) {
		triggered := false
		return func(dispatched int, ctrl *speedctrl.Controller, windows *[]PauseWindow) {
			if triggered || dispatched != triggerAt {
				return
			}
			triggered = true
			start := time.Now()
			ctrl.SetManualPaused(true)
			*windows = append(*windows, PauseWindow{
				StartNS: start.UnixNano(),
				EndNS:   start.Add(pauseDuration).UnixNano(),
				Source:  "manual",
			})
			go func() {
				time.Sleep(pauseDuration)
				ctrl.SetManualPaused(false)
			}()
		}
	}

	orGateHook := func() func(int, *speedctrl.Controller, *[]PauseWindow) {
		triggered := false
		return func(dispatched int, ctrl *speedctrl.Controller, windows *[]PauseWindow) {
			if triggered || dispatched != 0 {
				return
			}
			triggered = true
			start := time.Now()
			apiDuration := 40 * time.Millisecond
			manualDuration := 80 * time.Millisecond

			ctrl.SetAPIPaused(true)
			ctrl.SetManualPaused(true)
			*windows = append(*windows, PauseWindow{
				StartNS: start.UnixNano(),
				EndNS:   start.Add(manualDuration).UnixNano(),
				Source:  "api_or_manual",
			})
			go func() {
				time.Sleep(apiDuration)
				ctrl.SetAPIPaused(false)
				time.Sleep(manualDuration - apiDuration)
				ctrl.SetManualPaused(false)
			}()
		}
	}

	return []scenarioConfig{
		{
			name:           "G1_manual_pause",
			tasksByCIDR:    map[string]int{"10.0.0.0/24": 4},
			bucketRate:     200,
			bucketCapacity: 1,
			expectation: RuleExpectation{
				Name:                 "G1_manual_pause",
				RequirePauseBlocking: true,
			},
			hook: manualPauseHook(1, 80*time.Millisecond),
		},
		{
			name:           "G3_or_gate",
			tasksByCIDR:    map[string]int{"10.0.1.0/24": 3},
			bucketRate:     200,
			bucketCapacity: 1,
			expectation: RuleExpectation{
				Name:                 "G3_or_gate",
				RequirePauseBlocking: true,
			},
			hook: orGateHook(),
		},
		{
			name:           "C1_single_cidr_steady_rate",
			tasksByCIDR:    map[string]int{"10.0.2.0/24": 8},
			bucketRate:     20,
			bucketCapacity: 1,
			expectation: RuleExpectation{
				Name:       "C1_single_cidr_steady_rate",
				BucketRate: 20,
				Tolerance:  0.35,
			},
		},
		{
			name:           "C2_single_cidr_burst_then_steady",
			tasksByCIDR:    map[string]int{"10.0.3.0/24": 12},
			bucketRate:     5,
			bucketCapacity: 5,
			expectation: RuleExpectation{
				Name:              "C2_single_cidr_burst_then_steady",
				MinImmediateBurst: 4,
				BurstMaxGapNS:     int64((25 * time.Millisecond).Nanoseconds()),
			},
		},
		{
			name:           "X1_combined_global_pause_with_cidr_bucket",
			tasksByCIDR:    map[string]int{"10.0.4.0/24": 6, "10.0.5.0/24": 6},
			bucketRate:     10,
			bucketCapacity: 2,
			expectation: RuleExpectation{
				Name:                 "X1_combined_global_pause_with_cidr_bucket",
				RequirePauseBlocking: true,
			},
			hook: manualPauseHook(2, 60*time.Millisecond),
		},
	}
}

func runScenario(sc scenarioConfig) (ScenarioRun, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collector := NewCollector(sc.name)
	ctrl := speedctrl.NewController()
	windows := make([]PauseWindow, 0, 2)

	cidrs := make([]string, 0, len(sc.tasksByCIDR))
	for cidr := range sc.tasksByCIDR {
		cidrs = append(cidrs, cidr)
	}
	sort.Strings(cidrs)

	buckets := make(map[string]*ratelimit.LeakyBucket, len(cidrs))
	for _, cidr := range cidrs {
		buckets[cidr] = ratelimit.NewLeakyBucket(sc.bucketRate, sc.bucketCapacity)
	}
	defer func() {
		for _, cidr := range cidrs {
			buckets[cidr].Close()
		}
	}()

	dispatched := 0
	for _, cidr := range cidrs {
		for i := 0; i < sc.tasksByCIDR[cidr]; i++ {
			collector.Record(Event{Kind: EventBucketWaitStart, CIDR: cidr, TaskIndex: i, TimestampNS: time.Now().UnixNano()})
			if err := buckets[cidr].Acquire(ctx); err != nil {
				return ScenarioRun{}, err
			}
			collector.Record(Event{Kind: EventBucketAcquired, CIDR: cidr, TaskIndex: i, TimestampNS: time.Now().UnixNano()})

			if sc.hook != nil {
				sc.hook(dispatched, ctrl, &windows)
			}

			collector.Record(Event{Kind: EventGateWaitStart, CIDR: cidr, TaskIndex: i, TimestampNS: time.Now().UnixNano()})
			select {
			case <-ctx.Done():
				return ScenarioRun{}, ctx.Err()
			case <-ctrl.Gate():
			}
			collector.Record(Event{Kind: EventGateReleased, CIDR: cidr, TaskIndex: i, TimestampNS: time.Now().UnixNano()})

			collector.Record(Event{Kind: EventTaskEnqueued, CIDR: cidr, TaskIndex: i, TimestampNS: time.Now().UnixNano()})
			dispatched++
			if sc.delay > 0 {
				time.Sleep(sc.delay)
			}
		}
	}

	expectation := sc.expectation
	if len(expectation.PauseWindows) == 0 && len(windows) > 0 {
		expectation.PauseWindows = append(expectation.PauseWindows, windows...)
	}
	events := collector.Events()
	verdict := Analyze(events, expectation)

	return ScenarioRun{
		Name:        sc.name,
		Config:      map[string]any{"bucket_rate": sc.bucketRate, "bucket_capacity": sc.bucketCapacity, "tasks_by_cidr": sc.tasksByCIDR},
		Expectation: expectation,
		Events:      events,
		Verdict:     verdict,
	}, nil
}
