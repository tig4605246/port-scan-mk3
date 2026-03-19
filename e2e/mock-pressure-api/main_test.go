package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

func TestAuthHandler_ValidCredentials(t *testing.T) {
	os.Setenv("USE_AUTH", "true")
	os.Setenv("AUTH_CLIENT_ID", "test-client")
	os.Setenv("AUTH_CLIENT_SECRET", "test-secret")
	defer os.Unsetenv("USE_AUTH")
	defer os.Unsetenv("AUTH_CLIENT_ID")
	defer os.Unsetenv("AUTH_CLIENT_SECRET")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader("client_id=test-client&client_secret=test-secret"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler := newMux()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp["token_type"] != "Bearer" {
		t.Fatalf("expected token_type Bearer, got %v", resp["token_type"])
	}
	if resp["expires_in"] != float64(3600) {
		t.Fatalf("expected expires_in 3600, got %v", resp["expires_in"])
	}
	if resp["access_token"] == "" {
		t.Fatalf("expected non-empty access_token")
	}
}

func TestAuthHandler_InvalidCredentials(t *testing.T) {
	os.Setenv("USE_AUTH", "true")
	os.Setenv("AUTH_CLIENT_ID", "test-client")
	os.Setenv("AUTH_CLIENT_SECRET", "test-secret")
	defer os.Unsetenv("USE_AUTH")
	defer os.Unsetenv("AUTH_CLIENT_ID")
	defer os.Unsetenv("AUTH_CLIENT_SECRET")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader("client_id=wrong&client_secret=secret"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler := newMux()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestDataHandler_WithValidToken(t *testing.T) {
	os.Setenv("USE_AUTH", "true")
	os.Setenv("AUTH_CLIENT_ID", "test-client")
	os.Setenv("AUTH_CLIENT_SECRET", "test-secret")
	os.Setenv("PRESSURE_VALUE_1", "85")
	os.Setenv("PRESSURE_VALUE_2", "72")
	defer os.Unsetenv("USE_AUTH")
	defer os.Unsetenv("AUTH_CLIENT_ID")
	defer os.Unsetenv("AUTH_CLIENT_SECRET")
	defer os.Unsetenv("PRESSURE_VALUE_1")
	defer os.Unsetenv("PRESSURE_VALUE_2")

	authRec := httptest.NewRecorder()
	authReq := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader("client_id=test-client&client_secret=test-secret"))
	authReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler := newMux()
	handler.ServeHTTP(authRec, authReq)

	var authResp map[string]any
	json.Unmarshal(authRec.Body.Bytes(), &authResp)
	token := authResp["access_token"].(string)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/data", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("expected 2 data entries, got %d", len(resp))
	}
	percent1 := resp[0]["data"].(map[string]any)["Percent"].(float64)
	percent2 := resp[1]["data"].(map[string]any)["Percent"].(float64)
	if percent1 != 85 || percent2 != 72 {
		t.Fatalf("expected Percent 85 and 72, got %v and %v", percent1, percent2)
	}
}

func TestDataHandler_WithoutToken(t *testing.T) {
	os.Setenv("USE_AUTH", "true")
	defer os.Unsetenv("USE_AUTH")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/data", nil)

	handler := newMux()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}
