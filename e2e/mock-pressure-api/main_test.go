package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNewPressureHandler_OK(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/pressure", nil)

	newPressureHandler("ok", 42, 0)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	var body map[string]int
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body failed: %v", err)
	}
	if body["pressure"] != 42 {
		t.Fatalf("unexpected pressure: %v", body)
	}
}

func TestNewPressureHandler_Fail(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/pressure", nil)

	newPressureHandler("fail", 0, 0)(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
}

func TestNewPressureHandler_Timeout(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/pressure", nil)

	newPressureHandler("timeout", 7, 1)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	var body map[string]int
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body failed: %v", err)
	}
	if body["pressure"] != 7 {
		t.Fatalf("unexpected pressure: %v", body)
	}
}

func TestGetenvAndGetenvInt(t *testing.T) {
	t.Setenv("E2E_ENV_STR", "value")
	if got := getenv("E2E_ENV_STR", "fallback"); got != "value" {
		t.Fatalf("unexpected getenv: %s", got)
	}
	if got := getenv("E2E_ENV_STR_MISSING", "fallback"); got != "fallback" {
		t.Fatalf("unexpected fallback: %s", got)
	}

	t.Setenv("E2E_ENV_INT", "12")
	if got := getenvInt("E2E_ENV_INT", 1); got != 12 {
		t.Fatalf("unexpected getenvInt: %d", got)
	}
	t.Setenv("E2E_ENV_INT", "bad")
	if got := getenvInt("E2E_ENV_INT", 9); got != 9 {
		t.Fatalf("unexpected getenvInt fallback: %d", got)
	}
	os.Unsetenv("E2E_ENV_INT")
	if got := getenvInt("E2E_ENV_INT", 8); got != 8 {
		t.Fatalf("unexpected getenvInt fallback for missing key: %d", got)
	}
}
