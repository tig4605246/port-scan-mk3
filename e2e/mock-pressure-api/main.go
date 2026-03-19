package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	useAuth           func() bool
	validClientID     func() string
	validClientSecret func() string
	pressureValue1    func() string
	pressureValue2    func() string
)

func init() {
	useAuth = func() bool { return getenv("USE_AUTH", "false") == "true" }
	validClientID = func() string { return getenv("AUTH_CLIENT_ID", "test-client") }
	validClientSecret = func() string { return getenv("AUTH_CLIENT_SECRET", "test-secret") }
	pressureValue1 = func() string { return getenv("PRESSURE_VALUE_1", "85") }
	pressureValue2 = func() string { return getenv("PRESSURE_VALUE_2", "72") }
}

var validTokens = struct {
	sync.RWMutex
	tokens map[string]time.Time
}{tokens: make(map[string]time.Time)}

func main() {
	mode := getenv("MODE", "ok")
	addr := getenv("ADDR", ":8080")
	pressure := getenvInt("PRESSURE", 20)
	delayMS := getenvInt("DELAY_MS", 5000)

	mux := newMux()
	mux.HandleFunc("/api/pressure", newPressureHandler(mode, pressure, delayMS))

	mux.HandleFunc("/auth", handleAuth)
	mux.HandleFunc("/data", handleData)

	log.Printf("mock pressure api listening on %s mode=%s useAuth=%v", addr, mode, useAuth())
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth", handleAuth)
	mux.HandleFunc("/data", handleData)
	return mux
}

func newPressureHandler(mode string, pressure, delayMS int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch mode {
		case "ok":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]int{"pressure": pressure})
		case "fail":
			http.Error(w, "mock fail", http.StatusInternalServerError)
		case "timeout":
			time.Sleep(time.Duration(delayMS) * time.Millisecond)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]int{"pressure": pressure})
		default:
			http.Error(w, "unknown mode", http.StatusInternalServerError)
		}
	}
}

func handleAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return
	}

	if !useAuth() {
		http.Error(w, "auth not enabled", http.StatusServiceUnavailable)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "parse error", http.StatusBadRequest)
		return
	}

	clientID := r.Form.Get("client_id")
	clientSecret := r.Form.Get("client_secret")

	if clientID != validClientID() || clientSecret != validClientSecret() {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	token := "mock-token-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	validTokens.Lock()
	validTokens.tokens[token] = time.Now().Add(3600 * time.Second)
	validTokens.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   3600,
	})
}

func handleData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !useAuth() {
		http.Error(w, "auth not enabled", http.StatusServiceUnavailable)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	validTokens.RLock()
	expiry, ok := validTokens.tokens[token]
	validTokens.RUnlock()

	if !ok || time.Now().After(expiry) {
		http.Error(w, "invalid or expired token", http.StatusUnauthorized)
		return
	}

	p1, _ := strconv.Atoi(pressureValue1())
	p2, _ := strconv.Atoi(pressureValue2())

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode([]map[string]any{
		{"data": map[string]any{"Percent": p1}},
		{"data": map[string]any{"Percent": p2}},
	})
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getenvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
