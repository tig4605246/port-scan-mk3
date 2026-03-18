package scanapp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSimplePressureFetcher_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"pressure": 85}`))
	}))
	defer srv.Close()

	fetcher := NewSimplePressureFetcher(srv.URL, srv.Client())
	ctx := context.Background()

	pressure, err := fetcher.Fetch(ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pressure != 85 {
		t.Errorf("expected 85, got %d", pressure)
	}
}

func TestSimplePressureFetcher_MissingField(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"load": 50}`))
	}))
	defer srv.Close()

	fetcher := NewSimplePressureFetcher(srv.URL, srv.Client())
	ctx := context.Background()

	_, err := fetcher.Fetch(ctx)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSimplePressureFetcher_Non200Status(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	fetcher := NewSimplePressureFetcher(srv.URL, srv.Client())
	ctx := context.Background()

	_, err := fetcher.Fetch(ctx)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected status error with 500, got: %v", err)
	}
}

func TestSimplePressureFetcher_StringValue(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"pressure": "85"}`))
	}))
	defer srv.Close()

	fetcher := NewSimplePressureFetcher(srv.URL, srv.Client())
	ctx := context.Background()

	pressure, err := fetcher.Fetch(ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pressure != 85 {
		t.Errorf("expected 85, got %d", pressure)
	}
}

func TestAuthenticatedPressureFetcher_Success(t *testing.T) {
	authCalls := 0
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authCalls++
		r.ParseForm()
		if r.Form.Get("client_id") != "test-client" || r.Form.Get("client_secret") != "test-secret" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"tok123","token_type":"Bearer","expires_in":3600}`))
	}))
	defer authSrv.Close()

	dataSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok123" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"data":{"Percent":42.5},"deviceIP":"5.6.7.8","site":"ALL","protocol":"vxlan"}]`))
	}))
	defer dataSrv.Close()

	fetcher := NewAuthenticatedPressureFetcher(authSrv.URL, dataSrv.URL, "test-client", "test-secret", dataSrv.Client())
	ctx := context.Background()

	pressure, err := fetcher.Fetch(ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pressure != 42 {
		t.Errorf("expected 42, got %d", pressure)
	}
	if authCalls != 1 {
		t.Errorf("expected 1 auth call, got %d", authCalls)
	}
}
