package scanapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
			logger.infof("[API] pressure api status=ok pressure=%d%% threshold=%d", pressure, threshold)

			paused := pressure >= threshold
			ctrl.SetAPIPaused(paused)
			if paused != prevPaused {
				if paused {
					logger.infof("[API] 路由器壓力過載，掃描已自動暫停 pressure=%d threshold=%d", pressure, threshold)
				} else {
					logger.infof("[API] 路由器壓力恢復，掃描已自動恢復 pressure=%d threshold=%d", pressure, threshold)
				}
				prevPaused = paused
			}
		}
	}
}

func fetchPressure(client *http.Client, url string) (int, error) {
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return 0, fmt.Errorf("pressure api status=%d", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, err
	}
	raw, ok := body["pressure"]
	if !ok {
		return 0, fmt.Errorf("pressure field missing")
	}
	switch v := raw.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	case string:
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		return n, nil
	default:
		return 0, fmt.Errorf("unsupported pressure field type: %T", raw)
	}
}
