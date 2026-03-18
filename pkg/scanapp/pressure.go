package scanapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// PressureFetcher defines the interface for fetching pressure data.
type PressureFetcher interface {
	Fetch(ctx context.Context) (int, error)
}

// SimplePressureFetcher makes plain HTTP GET requests.
type SimplePressureFetcher struct {
	url    string
	client *http.Client
}

// NewSimplePressureFetcher creates a SimplePressureFetcher.
func NewSimplePressureFetcher(url string, client *http.Client) PressureFetcher {
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Second}
	}
	return &SimplePressureFetcher{url: url, client: client}
}

// Fetch retrieves pressure value from the configured URL.
func (f *SimplePressureFetcher) Fetch(ctx context.Context) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := f.client.Do(req)
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
