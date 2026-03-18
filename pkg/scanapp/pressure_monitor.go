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
	threshold := opts.PressureLimit
	if threshold <= 0 {
		threshold = defaultPressureLimit
	}
	thresholdValue := float64(threshold)

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
	var prevPaused bool
	pressureObserver := opts.pressureObserver
	controllerObserver := opts.controllerObserver
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
			consecutiveFailures = 0
			logger.infof("[API] pressure api status=ok pressure=%.1f%% threshold=%.1f", pressure, thresholdValue)

			sampledAt := time.Now()
			if pressureObserver != nil {
				pressureObserver.OnPressureSample(pressure, sampledAt)
			}
			paused := pressure >= thresholdValue
			ctrl.SetAPIPaused(paused)
			if controllerObserver != nil {
				controllerObserver.OnController(ctrl.ManualPaused(), ctrl.APIPaused())
			}
			if paused != prevPaused {
				if paused {
					logger.infof("[API] 路由器壓力過載，掃描已自動暫停 pressure=%.1f threshold=%.1f", pressure, thresholdValue)
				} else {
					logger.infof("[API] 路由器壓力恢復，掃描已自動恢復 pressure=%.1f threshold=%.1f", pressure, thresholdValue)
				}
				prevPaused = paused
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
