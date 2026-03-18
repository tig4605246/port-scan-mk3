package scanapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
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

// AuthenticatedPressureFetcher handles OAuth-style auth flow.
type AuthenticatedPressureFetcher struct {
	authURL      string
	dataURL      string
	clientID     string
	clientSecret string
	client       *http.Client

	mu          sync.Mutex
	accessToken string
	expiresAt   time.Time
}

// NewAuthenticatedPressureFetcher creates an AuthenticatedPressureFetcher.
func NewAuthenticatedPressureFetcher(authURL, dataURL, clientID, clientSecret string, client *http.Client) PressureFetcher {
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Second}
	}
	return &AuthenticatedPressureFetcher{
		authURL:      authURL,
		dataURL:      dataURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		client:       client,
	}
}

// Fetch retrieves pressure value with token authentication.
func (f *AuthenticatedPressureFetcher) Fetch(ctx context.Context) (int, error) {
	// Get valid token (refresh if needed)
	token, err := f.getToken(ctx)
	if err != nil {
		return 0, fmt.Errorf("auth failed: %w", err)
	}

	// Make data request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.dataURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := f.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return 0, fmt.Errorf("data api status=%d", resp.StatusCode)
	}

	// Parse array response
	var data []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}
	if len(data) == 0 {
		return 0, fmt.Errorf("no data entries in response")
	}

	// Extract Percent from first element's "data" field
	first := data[0]
	dataObj, ok := first["data"].(map[string]any)
	if !ok {
		return 0, fmt.Errorf("data field missing or not object")
	}
	raw, ok := dataObj["Percent"]
	if !ok {
		return 0, fmt.Errorf("Percent field missing")
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
		return 0, fmt.Errorf("unsupported Percent type: %T", raw)
	}
}

func (f *AuthenticatedPressureFetcher) getToken(ctx context.Context) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Check if we have a valid token (with 30s buffer)
	if f.accessToken != "" && time.Now().Add(30*time.Second).Before(f.expiresAt) {
		return f.accessToken, nil
	}

	// Need to refresh token
	form := url.Values{
		"client_id":     {f.clientID},
		"client_secret": {f.clientSecret},
	}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.authURL, strings.NewReader(form))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := f.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("auth status=%d", resp.StatusCode)
	}

	var authResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", fmt.Errorf("failed to decode auth response: %w", err)
	}
	if authResp.AccessToken == "" {
		return "", fmt.Errorf("access_token missing in auth response")
	}
	if authResp.TokenType != "Bearer" {
		return "", fmt.Errorf("unexpected token_type: %s (expected Bearer)", authResp.TokenType)
	}

	f.accessToken = authResp.AccessToken
	f.expiresAt = time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second)

	return f.accessToken, nil
}
