package scanapp

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/xuxiping/port-scan-mk3/pkg/config"
	"github.com/xuxiping/port-scan-mk3/pkg/speedctrl"
)

func TestPressureZones_Transitions(t *testing.T) {
	// Test: Zone C (ramp-up) when pressure 20% - multiplier increases
	// Test: Zone B (ramp-down) when pressure 45% - multiplier decreases
	// Test: Zone A (pause) when pressure 65% - multiplier resets to 0
	// Test: Zone B entry from pause - multiplier starts at floor 0.20

	var mu sync.Mutex
	pressures := []float64{20, 20, 20} // Zone C - ramp-up

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		p := pressures[0]
		if len(pressures) > 1 {
			pressures = pressures[1:]
		}
		mu.Unlock()
		w.Write([]byte(`{"pressure":` + fmt.Sprintf("%.0f", p) + `}`))
	}))
	defer api.Close()

	cfg := config.Config{
		PressureAPI:      api.URL,
		PressureInterval: 10 * time.Millisecond,
		PauseThreshold:   60,
		SafeThreshold:    30,
	}
	ctrl := speedctrl.NewController()
	logOut := &lockedBuffer{}
	logger := newLogger("info", false, logOut)
	errCh := make(chan error, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	go pollPressureAPI(ctx, cfg, RunOptions{
		PauseThreshold: cfg.PauseThreshold,
		SafeThreshold:  cfg.SafeThreshold,
		RampStep:       0.10,
		PressureHTTP:   &http.Client{Timeout: time.Second},
	}, ctrl, logger, errCh)

	time.Sleep(40 * time.Millisecond)

	// After 3 polls at 20% (Zone C), multiplier should be 1.0 + 3*0.10 = 1.30 -> clamped to 1.0
	mult := ctrl.GetSpeedMultiplier()
	if mult < 0.9 || mult > 1.0 {
		t.Fatalf("expected multiplier ~1.0 after ramp-up, got %.2f", mult)
	}
}

func TestPressureZones_ZoneB_FromPause_StartsAtFloor(t *testing.T) {
	// Resume from pause (>=60%) directly into Zone B (30-59%)
	// Multiplier should start at 0.20 (floor), not 0.0

	var mu sync.Mutex
	pressures := []float64{65, 45} // Zone A then Zone B

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		p := pressures[0]
		if len(pressures) > 1 {
			pressures = pressures[1:]
		}
		mu.Unlock()
		w.Write([]byte(`{"pressure":` + fmt.Sprintf("%.0f", p) + `}`))
	}))
	defer api.Close()

	cfg := config.Config{
		PressureAPI:      api.URL,
		PressureInterval: 10 * time.Millisecond,
		PauseThreshold:   60,
		SafeThreshold:    30,
	}
	ctrl := speedctrl.NewController()
	logOut := &lockedBuffer{}
	logger := newLogger("info", false, logOut)
	errCh := make(chan error, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	go pollPressureAPI(ctx, cfg, RunOptions{
		PauseThreshold: cfg.PauseThreshold,
		SafeThreshold:  cfg.SafeThreshold,
		RampStep:       0.10,
		PressureHTTP:   &http.Client{Timeout: time.Second},
	}, ctrl, logger, errCh)

	time.Sleep(25 * time.Millisecond)

	// After Zone A (pause) then Zone B, multiplier should be at floor 0.20
	// NOT 0.0, and NOT still decreasing from 0
	mult := ctrl.GetSpeedMultiplier()
	if mult < 0.19 || mult > 0.21 {
		t.Fatalf("expected multiplier ~0.20 (floor) after Zone A->B, got %.2f", mult)
	}
}
