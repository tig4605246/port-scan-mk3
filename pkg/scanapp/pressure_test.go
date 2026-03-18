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
