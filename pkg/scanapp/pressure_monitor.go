package scanapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/xuxiping/port-scan-mk3/pkg/config"
	"github.com/xuxiping/port-scan-mk3/pkg/speedctrl"
)

func startManualPauseMonitor(ctx context.Context, ctrl *speedctrl.Controller, logger *scanLogger) {
	go func() {
		prev := ctrl.ManualPaused()
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				curr := ctrl.ManualPaused()
				if curr != prev {
					if curr {
						logger.infof("[Manual] 接收到按鍵指令，掃描已手動暫停")
					} else {
						logger.infof("[Manual] 掃描已手動恢復")
					}
					prev = curr
				}
			}
		}
	}()
}

func pollPressureAPI(ctx context.Context, cfg config.Config, opts RunOptions, ctrl *speedctrl.Controller, logger *scanLogger, errCh chan<- error) {
	interval := cfg.PressureInterval
	if interval <= 0 {
		interval = 5 * time.Second
	}

	// Use PressureFetcher if provided, otherwise create SimplePressureFetcher for backward compatibility
	var fetcher PressureFetcher
	if opts.PressureFetcher != nil {
		fetcher = opts.PressureFetcher
	} else {
		client := opts.PressureHTTP
		if client == nil {
			client = &http.Client{Timeout: 2 * time.Second}
		}
		fetcher = NewSimplePressureFetcher(cfg.PressureAPI, client)
	}

	var consecutiveFailures int
	pressureObserver := opts.pressureObserver
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pressure, err := fetcher.Fetch(ctx)
			if err != nil {
				consecutiveFailures++
				if pressureObserver != nil {
					pressureObserver.OnPressureFailure(consecutiveFailures, time.Now())
				}
				if consecutiveFailures <= 2 {
					logger.errorf("pressure api request failed (%d/3): %v", consecutiveFailures, err)
					continue
				}
				select {
				case errCh <- fmt.Errorf("pressure api failed 3 times: %w", err):
				default:
				}
				return
			}
			thresholdPause := float64(opts.PauseThreshold)
			thresholdSafe := float64(opts.SafeThreshold)
			rampStep := opts.RampStep

			consecutiveFailures = 0
			logger.infof("[API] pressure api status=ok pressure=%.1f%% pause_thresh=%.1f safe_thresh=%.1f", pressure, thresholdPause, thresholdSafe)

			sampledAt := time.Now()
			if pressureObserver != nil {
				pressureObserver.OnPressureSample(int(pressure), sampledAt)
			}

			if pressure >= thresholdPause {
				ctrl.SetAPIPaused(true)
				ctrl.ResetSpeedMultiplier()
			} else if pressure >= thresholdSafe {
				ctrl.SetAPIPaused(false)
				if ctrl.GetSpeedMultiplier() == 0.0 {
					ctrl.SetSpeedMultiplier(speedctrl.SpeedMultiplierMin)
				}
				ctrl.AdjustSpeedMultiplier(-rampStep)
			} else {
				ctrl.SetAPIPaused(false)
				ctrl.AdjustSpeedMultiplier(rampStep)
			}
		}
	}
}

func fetchPressure(client *http.Client, url string) (float64, error) {
	resp, err := client.Get(url)
	if err != nil {
		return 0.0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return 0.0, fmt.Errorf("pressure api status=%d", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0.0, err
	}
	raw, ok := body["pressure"]
	if !ok {
		return 0.0, fmt.Errorf("pressure field missing")
	}
	n, err := parsePressureValue(raw)
	if err != nil {
		return 0.0, err
	}
	return n, nil
}
