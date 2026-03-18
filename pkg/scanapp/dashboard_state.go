package scanapp

import (
	"fmt"
	"sync"
	"time"
)

const dashboardRateWindow = 5 * time.Second

type dashboardSnapshot struct {
	ScannedTasks          int
	TotalTasks            int
	Percent               float64
	PressurePercent       int
	CurrentCIDR           string
	BucketStatus          string
	ControllerStatus      string
	DispatchPerSecond     float64
	ResultsPerSecond      float64
	APIHealthText         string
	LastPressureUpdateAt  time.Time
	LastPressureFailureAt time.Time
}

type dashboardState struct {
	mu sync.Mutex

	totalTasks   int
	scannedTasks int

	currentCIDR  string
	bucketStatus string

	manualPaused bool
	apiPaused    bool

	pressurePercent int

	dispatchEvents []time.Time
	resultEvents   []time.Time

	apiHealthText         string
	lastPressureUpdateAt  time.Time
	lastPressureFailureAt time.Time

	now func() time.Time
}

func newDashboardState(total int, now func() time.Time) *dashboardState {
	if total < 0 {
		total = 0
	}
	if now == nil {
		now = time.Now
	}
	return &dashboardState{
		totalTasks:    total,
		apiHealthText: "ok",
		now:           now,
	}
}

func (s *dashboardState) OnTaskEnqueued(cidr string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cidr != "" {
		s.currentCIDR = cidr
	}
	now := s.now()
	s.dispatchEvents = append(s.dispatchEvents, now)
	s.dispatchEvents = pruneDashboardEvents(s.dispatchEvents, now)
}

func (s *dashboardState) OnResult() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.scannedTasks++
	now := s.now()
	s.resultEvents = append(s.resultEvents, now)
	s.resultEvents = pruneDashboardEvents(s.resultEvents, now)
}

func (s *dashboardState) OnBucketStatus(cidr, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cidr != "" {
		s.currentCIDR = cidr
	}
	s.bucketStatus = status
}

func (s *dashboardState) OnController(manualPaused, apiPaused bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.manualPaused = manualPaused
	s.apiPaused = apiPaused
}

func (s *dashboardState) OnPressureSample(pressure int, t time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pressurePercent = pressure
	s.apiHealthText = "ok"
	s.lastPressureUpdateAt = t
}

func (s *dashboardState) OnPressureFailure(streak int, t time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.apiHealthText = fmt.Sprintf("fail streak %d", streak)
	s.lastPressureFailureAt = t
}

func (s *dashboardState) Snapshot() dashboardSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	s.dispatchEvents = pruneDashboardEvents(s.dispatchEvents, now)
	s.resultEvents = pruneDashboardEvents(s.resultEvents, now)

	return dashboardSnapshot{
		ScannedTasks:          s.scannedTasks,
		TotalTasks:            s.totalTasks,
		Percent:               dashboardPercent(s.scannedTasks, s.totalTasks),
		PressurePercent:       s.pressurePercent,
		CurrentCIDR:           s.currentCIDR,
		BucketStatus:          s.bucketStatus,
		ControllerStatus:      dashboardControllerStatus(s.manualPaused, s.apiPaused),
		DispatchPerSecond:     float64(len(s.dispatchEvents)) / dashboardRateWindow.Seconds(),
		ResultsPerSecond:      float64(len(s.resultEvents)) / dashboardRateWindow.Seconds(),
		APIHealthText:         s.apiHealthText,
		LastPressureUpdateAt:  s.lastPressureUpdateAt,
		LastPressureFailureAt: s.lastPressureFailureAt,
	}
}

func pruneDashboardEvents(events []time.Time, now time.Time) []time.Time {
	if len(events) == 0 {
		return events
	}
	cutoff := now.Add(-dashboardRateWindow)
	keep := 0
	for keep < len(events) && events[keep].Before(cutoff) {
		keep++
	}
	if keep == 0 {
		return events
	}
	pruned := make([]time.Time, len(events)-keep)
	copy(pruned, events[keep:])
	return pruned
}

func dashboardPercent(scanned, total int) float64 {
	if total <= 0 {
		return 0
	}
	if scanned > total {
		scanned = total
	}
	return (float64(scanned) / float64(total)) * 100
}

func dashboardControllerStatus(manualPaused, apiPaused bool) string {
	switch {
	case manualPaused && apiPaused:
		return "PAUSED(API+MANUAL)"
	case apiPaused:
		return "PAUSED(API)"
	case manualPaused:
		return "PAUSED(MANUAL)"
	default:
		return "RUNNING"
	}
}
